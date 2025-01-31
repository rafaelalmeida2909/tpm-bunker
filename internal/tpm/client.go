package tpm

import (
	"context"
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

func (c *TPMClient) SignData(ctx context.Context, hash []byte) ([]byte, error) {
	log.Printf("Tamanho do hash a ser assinado: %d bytes", len(hash))
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}
	if len(hash) != sha256.Size {
		return nil, fmt.Errorf("hash deve ter tamanho SHA256 (32 bytes)")
	}
	if c.rwc == nil {
		return nil, fmt.Errorf("TPM connection not initialized")
	}

	log.Printf("Hash a ser assinado (hex): %x", hash)

	sig, err := tpm2.Sign(
		c.rwc,
		c.rsaHandle,
		"",   // Sem senha
		hash, // Hash pré-calculado
		nil,  // Sem ticket
		&tpm2.SigScheme{
			Alg:  tpm2.AlgRSASSA, // Mesmo algoritmo definido na chave
			Hash: tpm2.AlgSHA256,
		},
	)
	if err != nil {
		log.Printf("Erro na assinatura TPM. Handle: %x, Erro: %v", c.rsaHandle, err)
		return nil, fmt.Errorf("TPM signing failed: %w", err)
	}

	if sig.RSA == nil || len(sig.RSA.Signature) == 0 {
		return nil, fmt.Errorf("invalid signature generated")
	}

	log.Printf("Assinatura gerada (hex): %x", sig.RSA.Signature)
	return sig.RSA.Signature, nil
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
			ekHandle:  tpmutil.Handle(0x81010001), // EK handle
			aikHandle: tpmutil.Handle(0x81008F01), // AIK handle
			rsaHandle: tpmutil.Handle(0x81008F02), // Handle dedicado para chave de assinatura
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

	aik, err := c.generateAIK(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar AIK: %v", err)
	}
	c.aik = aik

	// Verificar existência da chave RSA
	_, _, _, err = tpm2.ReadPublic(c.rwc, c.rsaHandle)
	if err == nil {
		log.Printf("Chave RSA encontrada no TPM")
	} else {
		log.Printf("Chave RSA não encontrada, criando uma nova...")
		keyPair, _, err := c.generateRSAKeyPair(ctx)
		if err != nil {
			return nil, fmt.Errorf("falha ao gerar par de chaves RSA: %v", err)
		}
		c.KeyPair = keyPair

		// Tornar o handle persistente
		err = tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, c.rsaHandle, c.rsaHandle)
		if err != nil {
			return nil, fmt.Errorf("falha ao tornar handle persistente: %v", err)
		}
		log.Printf("Chave RSA criada e persistida com sucesso")
	}

	pubKey, err := c.RetrieveRSAKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao recuperar chave RSA: %v", err)
	}

	pubKeyPEM := GetPublicKeyPEM(pubKey)
	fmt.Printf("PUBKEY %s", pubKeyPEM)
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
				tpm2.FlagRestricted | tpm2.FlagSignerDefault,
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
		// Verificar se o handle persistente já existe
		_, _, _, err := tpm2.ReadPublic(c.rwc, c.rsaHandle)
		if err == nil {
			log.Printf("Handle persistente já existe para rsaHandle")
			pubKey, err := c.RetrieveRSAKey(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("falha ao recuperar chave existente: %v", err)
			}
			return nil, pubKey, nil
		}

		// Create template for RSA key
		template := tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			// Remove FlagRestricted porque pode estar causando o erro
			Attributes: tpm2.FlagFixedTPM |
				tpm2.FlagFixedParent |
				tpm2.FlagSensitiveDataOrigin |
				tpm2.FlagUserWithAuth |
				tpm2.FlagRestricted |
				tpm2.FlagSign | // Apenas FlagSign para assinatura
				tpm2.FlagDecrypt,
			RSAParameters: &tpm2.RSAParams{
				KeyBits:     2048,
				ExponentRaw: 0x10001,
				Sign: &tpm2.SigScheme{
					Alg:  tpm2.AlgRSASSA,
					Hash: tpm2.AlgSHA256,
				},
			},
		}
		// Create primary key in owner hierarchy
		keyHandle, pub, _, _, _, _, err := tpm2.CreatePrimaryEx(
			c.rwc,
			tpm2.HandleOwner,
			tpm2.PCRSelection{},
			"",
			"",
			template,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("falha ao criar chave RSA: %v", err)
		}
		defer tpm2.FlushContext(c.rwc, keyHandle)

		// Limpar handle antigo se existir
		_ = tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, c.rsaHandle, c.rsaHandle)

		// Tornar o handle persistente
		err = tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, keyHandle, c.rsaHandle)
		if err != nil {
			return nil, nil, fmt.Errorf("falha ao tornar handle persistente: %v", err)
		}

		// Decodificar a chave pública diretamente do retorno do TPM
		rsaPub, err := tpm2.DecodePublic(pub)
		if err != nil {
			log.Fatalf("Falha ao decodificar chave pública: %v", err)
		}

		// Valida o exponente da chave pública
		if rsaPub.RSAParameters.ExponentRaw != 0x10001 {
			log.Fatalf("Exponente inválido na chave RSA gerada: %d", rsaPub.RSAParameters.ExponentRaw)
		}

		log.Printf("Chave RSA gerada com sucesso. Exponente: %d", rsaPub.RSAParameters.ExponentRaw)

		pubKey := rsa.PublicKey{
			N: rsaPub.RSAParameters.Modulus(),
			E: 65537,
		}

		// Criar chave privada como wrapper
		privKey := &rsa.PrivateKey{
			PublicKey: pubKey,
		}

		return privKey, &pubKey, nil
	}
}

func (c *TPMClient) RetrieveRSAKey(ctx context.Context) (*rsa.PublicKey, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Lê a chave pública do handle persistente
		pub, _, _, err := tpm2.ReadPublic(c.rwc, c.rsaHandle)
		if err != nil {
			return nil, fmt.Errorf("falha ao ler chave RSA do TPM: %v", err)
		}

		// Serializa a chave pública
		pubBytes, err := pub.Encode()
		if err != nil {
			return nil, fmt.Errorf("falha ao serializar chave pública: %v", err)
		}

		// Decodifica a chave pública
		rsaPub, err := tpm2.DecodePublic(pubBytes)
		if err != nil {
			return nil, fmt.Errorf("falha ao decodificar chave pública RSA: %v", err)
		}

		// Converte para *rsa.PublicKey
		pubKey := rsa.PublicKey{
			N: rsaPub.RSAParameters.Modulus(),
			E: 65537,
		}

		return &pubKey, nil
	}
}

// RSADecrypt decrypts data using the TPM's RSA key
func (c *TPMClient) RSADecrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, c.rsaHandle, c.rsaHandle)
		// Primeiro verifica se a chave existe
		pub, _, _, err := tpm2.ReadPublic(c.rwc, c.rsaHandle)
		if err != nil {
			log.Fatalf("Erro ao ler chave RSA do TPM: %v", err)
		}
		log.Printf("Atributos da chave: %x", pub.Attributes)

		handles, _, err := tpm2.GetCapability(c.rwc, tpm2.CapabilityHandles, 100, uint32(tpm2.PersistentFirst))
		if err != nil {
			log.Fatalf("Erro ao listar handles: %v", err)
		}
		log.Printf("Handles disponíveis: %v", handles)

		// Configura o esquema de decriptação para OAEP com SHA256
		scheme := &tpm2.AsymScheme{
			Alg:  tpm2.AlgOAEP, // Mesmo algoritmo definido na chave
			Hash: tpm2.AlgSHA256,
		}
		fmt.Printf(string(ciphertext))
		// Log para debug
		log.Printf("Tentando decriptar com handle: %x", c.rsaHandle)
		log.Printf("Tamanho do ciphertext: %d bytes", len(ciphertext))

		// Decripta usando OAEP com SHA256
		decrypted, err := tpm2.RSADecrypt(
			c.rwc,
			c.rsaHandle,
			"", // Sem senha
			ciphertext,
			scheme,
			"", // Label vazio para OAEP
		)
		if err != nil {
			log.Printf("Erro na decriptação TPM. Handle: %x, Erro: %v", c.rsaHandle, err)
			return nil, fmt.Errorf("TPM RSA decryption failed: handle %x, error code %v", c.rsaHandle, err)
		}

		return decrypted, nil
	}
}

// Close fecha a conexão com o TPM
func (c *TPMClient) Close() error {
	if c.rwc != nil {
		return c.rwc.Close()
	}
	return nil
}
