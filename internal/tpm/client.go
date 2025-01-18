package tpm

import (
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
	keyPair *rsa.PrivateKey

	// Handles persistentes
	ekHandle  tpmutil.Handle
	aikHandle tpmutil.Handle
	rsaHandle tpmutil.Handle
}

func getPublicKeyPEM(pubKey *rsa.PublicKey) string {
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
func checkTPMDevice() bool {
	if runtime.GOOS == "windows" {
		// No Windows, primeiro tentamos abrir o device para ter certeza que temos acesso
		handle, err := os.OpenFile("\\\\.\\TPM", os.O_RDWR, 0)
		if err != nil {
			if os.IsPermission(err) {
				log.Printf("Sem permissão para acessar o TPM no Windows. Execute como administrador.")
				return false
			} else if os.IsNotExist(err) {
				log.Printf("TPM device não encontrado no Windows")
				return false
			}
			log.Printf("Erro ao acessar TPM no Windows: %v", err)
			return false
		}
		handle.Close()
		return true
	}

	// Para Linux
	if _, err := os.Stat("/dev/tpm0"); err != nil {
		if os.IsNotExist(err) {
			// Tenta o tpmrm0
			if _, err := os.Stat("/dev/tpmrm0"); err != nil {
				log.Printf("Nenhum device TPM encontrado no Linux (/dev/tpm0 ou /dev/tpmrm0)")
				return false
			}
			return true // tpmrm0 exists
		}
		log.Printf("Erro ao verificar TPM no Linux: %v", err)
		return false
	}

	return true // tpm0 exists
}

// NewTPMClient verifica a presença do TPM e inicializa uma nova conexão
func NewTPMClient() (*TPMClient, error) {
	// Primeiro verifica se o device existe
	if !checkTPMDevice() {
		return nil, fmt.Errorf("TPM device não encontrado no sistema")
	}

	// Tenta estabelecer a conexão
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

// CheckTPMPresence é um wrapper para verificação rápida de disponibilidade
func CheckTPMPresence() bool {
	if !checkTPMDevice() {
		return false
	}

	// Tenta uma conexão rápida para verificar se o TPM está funcional
	rwc, err := tpm2.OpenTPM()
	if err != nil {
		log.Printf("TPM device existe mas não pode ser inicializado: %v", err)
		return false
	}
	rwc.Close()
	return true
}

// InitializeDevice configura o dispositivo pela primeira vez
func (c *TPMClient) InitializeDevice() (*types.DeviceInfo, error) {

	// Recupera EK
	ek, err := c.getEndorsementKey()
	if err != nil {
		return nil, fmt.Errorf("falha ao recuperar EK: %v", err)
	}
	c.ek = ek

	// Gera AIK
	aik, err := c.generateAIK()
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar AIK: %v", err)
	}
	c.aik = aik

	// Gera par de chaves RSA
	keyPair, pubKey, err := c.generateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar par de chaves RSA: %v", err)
	}
	c.keyPair = keyPair

	pubKeyPEM := getPublicKeyPEM(pubKey)

	// Gera UUID baseado no EK
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
func (c *TPMClient) getEndorsementKey() ([]byte, error) {
	template := tpm2.Public{
		Type:    tpm2.AlgRSA,
		NameAlg: tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
			tpm2.FlagAdminWithPolicy | tpm2.FlagRestricted | tpm2.FlagDecrypt,
		RSAParameters: &tpm2.RSAParams{
			Symmetric: &tpm2.SymScheme{
				Alg:     tpm2.AlgAES,
				KeyBits: 128,
				Mode:    tpm2.AlgCFB,
			},
			KeyBits: 2048,
		},
	}

	// Tenta criar a chave primária
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

	// Limpa a chave ao sair da função
	defer tpm2.FlushContext(c.rwc, ekHandle)

	return pubKey, nil
}

// generateAIK gera uma nova chave de identidade de atestação
func (c *TPMClient) generateAIK() ([]byte, error) {
	template := tpm2.Public{
		Type:    tpm2.AlgRSA,
		NameAlg: tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
			tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagSign,
		RSAParameters: &tpm2.RSAParams{
			Sign: &tpm2.SigScheme{
				Alg:  tpm2.AlgRSASSA,
				Hash: tpm2.AlgSHA256,
			},
			KeyBits: 2048,
		},
	}

	// Cria AIK sob a hierarquia de endosso
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

	// Limpa a chave ao sair da função
	defer tpm2.FlushContext(c.rwc, aikHandle)

	return pubKey, nil
}

// generateRSAKeyPair gera um novo par de chaves RSA protegido pelo TPM
func (c *TPMClient) generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	template := tpm2.Public{
		Type:    tpm2.AlgRSA,
		NameAlg: tpm2.AlgSHA256,
		Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin |
			tpm2.FlagUserWithAuth | tpm2.FlagSign | tpm2.FlagDecrypt,
		RSAParameters: &tpm2.RSAParams{
			Sign:        nil, // Removendo o esquema de assinatura específico
			KeyBits:     2048,
			ExponentRaw: 0, // Usar expoente padrão (65537)
			ModulusRaw:  nil,
		},
	}

	// Cria par de chaves sob a hierarquia de proprietário
	keyHandle, pubKey, _, _, _, _, err := tpm2.CreatePrimaryEx(
		c.rwc,
		tpmutil.Handle(tpm2.HandleOwner),
		tpm2.PCRSelection{},
		"",
		"",
		template,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao criar par de chaves RSA: %v", err)
	}

	// Limpa a chave ao sair da função
	defer tpm2.FlushContext(c.rwc, keyHandle)

	// Converte o pubKey para *rsa.PublicKey
	rsaPub, err := tpm2.DecodePublic(pubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("falha ao decodificar chave pública: %v", err)
	}

	n := rsaPub.RSAParameters.Modulus()
	e := int(rsaPub.RSAParameters.Exponent())

	if n == nil {
		return nil, nil, fmt.Errorf("módulo RSA nulo")
	}

	publicKey := &rsa.PublicKey{
		N: n,
		E: e,
	}

	// Cria uma estrutura PrivateKey apenas com a parte pública
	privateKey := &rsa.PrivateKey{
		PublicKey: *publicKey,
	}

	return privateKey, publicKey, nil
}

// Close fecha a conexão com o TPM
func (c *TPMClient) Close() error {
	if c.rwc != nil {
		return c.rwc.Close()
	}
	return nil
}
