package tpm

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"tpm-bunker/internal/types"

	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"github.com/google/uuid"
)

type TPMClient struct {
	rwc     io.ReadWriteCloser
	ek      []byte
	aik     []byte
	KeyPair *rsa.PrivateKey

	// Handles persistentes
	ekHandle  tpmutil.Handle
	aikHandle tpmutil.Handle
	rsaHandle tpmutil.Handle
}

func GetPublicKeyPEM(pubKey *rsa.PublicKey) string {
	pubASN1, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		log.Fatalf("Falha ao serializar chave pública: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})
	return string(pubPEM)
}

// checkTPMDevice verifica apenas a existência do arquivo de dispositivo TPM
func checkTPMDevice(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		if runtime.GOOS == "windows" {
			handle, err := os.OpenFile("\\\\.\\TPM", os.O_RDWR, 0)
			if err != nil {
				if os.IsPermission(err) {
					log.Printf("Sem permissão para acessar o TPM no Windows")
					return false
				}
				log.Printf("Erro ao acessar TPM no Windows: %v", err)
				return false
			}
			defer handle.Close()
			return true
		}

		// Para Linux
		if _, err := os.Stat("/dev/tpm0"); err != nil {
			if os.IsNotExist(err) {
				if _, err := os.Stat("/dev/tpmrm0"); err != nil {
					log.Printf("Nenhum device TPM encontrado")
					return false
				}
				return true
			}
			log.Printf("Erro ao verificar TPM: %v", err)
			return false
		}
		return true
	}
}

// NewTPMClient verifica a presença do TPM e inicializa uma nova conexão
func NewTPMClient(ctx context.Context) (*TPMClient, error) {
	if !checkTPMDevice(ctx) {
		return nil, fmt.Errorf("TPM device não encontrado")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		rwc, err := tpm2.OpenTPM()
		if err != nil {
			return nil, fmt.Errorf("falha ao inicializar TPM: %v", err)
		}

		return &TPMClient{
			rwc:       rwc,
			ekHandle:  tpmutil.Handle(0x81000001),
			aikHandle: tpmutil.Handle(0x81000002),
			rsaHandle: tpmutil.Handle(0x81000003),
		}, nil
	}
}

// CheckTPMPresence é um wrapper para verificação rápida de disponibilidade
func CheckTPMPresence(ctx context.Context) bool {
	if !checkTPMDevice(ctx) {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	default:
		rwc, err := tpm2.OpenTPM()
		if err != nil {
			log.Printf("TPM device existe mas não pode ser inicializado: %v", err)
			return false
		}
		defer rwc.Close()
		return true
	}
}

// InitializeDevice configura o dispositivo pela primeira vez
func (c *TPMClient) InitializeDevice(ctx context.Context) (*types.DeviceInfo, error) {
	// Verifica cancelamento inicial
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	ek, err := c.getEndorsementKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao recuperar EK: %v", err)
	}
	c.ek = ek

	// Verifica cancelamento antes de gerar AIK
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	aik, err := c.generateAIK(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar AIK: %v", err)
	}
	c.aik = aik

	// Verifica cancelamento antes de gerar RSA
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	keyPair, pubKey, err := c.generateRSAKeyPair(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar par de chaves RSA: %v", err)
	}
	c.KeyPair = keyPair

	pubKeyPEM := GetPublicKeyPEM(pubKey)
	deviceUUID, err := generateTPMBasedUUID(ek)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar UUID: %v", err)
	}

	return &types.DeviceInfo{
		UUID:      deviceUUID,
		EK:        ek,
		AIK:       aik,
		PublicKey: pubKeyPEM,
	}, nil
}

// generateTPMBasedUUID gera um UUID v5 usando o EK como namespace
func generateTPMBasedUUID(ek []byte) (string, error) {
	// Cria um hash do EK para usar como namespace
	hash := sha256.Sum256(ek)
	namespace, err := uuid.FromBytes(hash[:16]) // Usa os primeiros 16 bytes do hash
	if err != nil {
		return "", fmt.Errorf("falha ao criar namespace do UUID: %v", err)
	}

	// Gera UUID v5 usando o namespace e o EK completo como nome
	deviceUUID := uuid.NewSHA1(namespace, ek)

	return deviceUUID.String(), nil
}

// getEndorsementKey recupera a chave de endosso do TPM
func (c *TPMClient) getEndorsementKey(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		template := tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent |
				tpm2.FlagSensitiveDataOrigin | tpm2.FlagAdminWithPolicy |
				tpm2.FlagRestricted | tpm2.FlagDecrypt,
			RSAParameters: &tpm2.RSAParams{
				Symmetric: &tpm2.SymScheme{
					Alg:     tpm2.AlgAES,
					KeyBits: 128,
					Mode:    tpm2.AlgCFB,
				},
				KeyBits: 2048,
			},
		}

		ekHandle, pubKey, _, _, _, _, err := tpm2.CreatePrimaryEx(
			c.rwc,
			tpm2.HandleEndorsement,
			tpm2.PCRSelection{},
			"",
			"",
			template,
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao criar EK: %v", err)
		}
		defer tpm2.FlushContext(c.rwc, ekHandle)

		return pubKey, nil
	}
}

// generateAIK gera uma nova chave de identidade de atestação
func (c *TPMClient) generateAIK(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		template := tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent |
				tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth |
				tpm2.FlagRestricted | tpm2.FlagSign,
			RSAParameters: &tpm2.RSAParams{
				Sign: &tpm2.SigScheme{
					Alg:  tpm2.AlgRSASSA,
					Hash: tpm2.AlgSHA256,
				},
				KeyBits: 2048,
			},
		}

		aikHandle, pubKey, _, _, _, _, err := tpm2.CreatePrimaryEx(
			c.rwc,
			tpm2.HandleEndorsement,
			tpm2.PCRSelection{},
			"",
			"",
			template,
		)
		if err != nil {
			return nil, fmt.Errorf("falha ao criar AIK: %v", err)
		}
		defer tpm2.FlushContext(c.rwc, aikHandle)

		return pubKey, nil
	}
}

// generateRSAKeyPair gera um novo par de chaves RSA protegido pelo TPM
func (c *TPMClient) generateRSAKeyPair(ctx context.Context) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
		// Gera um par de chaves RSA diretamente
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, nil, fmt.Errorf("falha ao gerar par de chaves RSA: %v", err)
		}

		// Salva a chave no TPM
		template := tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent |
				tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth |
				tpm2.FlagSign | tpm2.FlagDecrypt,
			RSAParameters: &tpm2.RSAParams{
				KeyBits:     2048,
				ExponentRaw: uint32(privateKey.PublicKey.E),
				ModulusRaw:  privateKey.PublicKey.N.Bytes(),
			},
		}

		keyHandle, _, _, _, _, _, err := tpm2.CreatePrimaryEx(
			c.rwc,
			tpmutil.Handle(tpm2.HandleOwner),
			tpm2.PCRSelection{},
			"",
			"",
			template,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("falha ao salvar chave no TPM: %v", err)
		}
		defer tpm2.FlushContext(c.rwc, keyHandle)

		// Armazena a chave privada completa no cliente
		c.KeyPair = privateKey

		return privateKey, &privateKey.PublicKey, nil
	}
}

// Close fecha a conexão com o TPM
func (c *TPMClient) Close() error {
	if c.rwc != nil {
		return c.rwc.Close()
	}
	return nil
}
