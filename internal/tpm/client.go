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
	"time"
	"tpm-bunker/internal/types"

	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"github.com/google/uuid"
)

type TPMClient struct {
	rwc io.ReadWriteCloser
	ek  []byte
	aik []byte

	// Handles persistentes
	ekHandle      tpmutil.Handle
	aikHandle     tpmutil.Handle
	signHandle    tpmutil.Handle // Handle para chave de assinatura
	decryptHandle tpmutil.Handle // Handle para chave de criptografia/decriptação
}

func (c *TPMClient) SignData(ctx context.Context, hash []byte) ([]byte, error) {
	caps, _, err := tpm2.GetCapability(c.rwc, tpm2.CapabilityAlgs, 100, 0)
	if err != nil {
		log.Printf("[SignData] Erro ao listar algoritmos suportados: %v", err)
	} else {
		log.Printf("[SignData] Algoritmos suportados pelo TPM: %v", caps)
	}

	log.Printf("[SignData] Starting signature operation")
	log.Printf("[SignData] Hash length: %d bytes", len(hash))
	log.Printf("[SignData] Hash value (hex): %x", hash)
	log.Printf("[SignData] Using sign handle: 0x%x", c.signHandle)

	// Check if handle exists and read its properties
	pub, _, _, err := tpm2.ReadPublic(c.rwc, c.signHandle)
	if err != nil {
		log.Printf("[SignData] ERROR: Failed to read public key: %v", err)
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}
	log.Printf("[SignData] Key attributes: 0x%x", pub.Attributes)
	log.Printf("[SignData] Key algorithm: %v", pub.Type)
	log.Printf("[SignData] Signing scheme: %v", pub.RSAParameters.Sign.Alg)

	signature, err := tpm2.Sign(
		c.rwc,
		c.signHandle,
		"", // Sem senha
		hash,
		nil,
		&tpm2.SigScheme{
			Alg:  tpm2.AlgRSASSA, // Mesmo algoritmo definido na chave
			Hash: tpm2.AlgSHA256,
		},
	)
	log.Printf("[SignData] Key scheme: %v", pub.RSAParameters.Sign)

	if err != nil {
		return nil, fmt.Errorf("erro ao assinar dados: %w", err)
	}

	log.Printf("[SignData] Signature length: %d bytes", len(signature.RSA.Signature))
	log.Printf("[SignData] Signature created successfully")

	return signature.RSA.Signature, nil
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

		client := &TPMClient{rwc: rwc}

		// Buscar os handles dinâmicos
		client.ekHandle = tpmutil.Handle(0x81010001)  // Handle fixo para EK
		client.aikHandle = tpmutil.Handle(0x81008F01) // Handle fixo para AIK

		// Procurar os handles para assinatura e decriptação
		signHandle, err := findRSAHandle(rwc, true)
		if err != nil {
			log.Printf("Chave de assinatura não encontrada, criando nova...")
			client.signHandle = tpmutil.Handle(0x81008F02) // Definir um novo handle
		} else {
			log.Printf("Chave de assinatura encontrada no handle: 0x%x", signHandle)
			client.signHandle = signHandle
		}

		decryptHandle, err := findRSAHandle(rwc, false)
		if err != nil {
			log.Printf("Chave de decriptação não encontrada, criando nova...")
			client.decryptHandle = tpmutil.Handle(0x81008F03) // Definir um novo handle
		} else {
			log.Printf("Chave de decriptação encontrada no handle: 0x%x", decryptHandle)
			client.decryptHandle = decryptHandle
		}

		return client, nil
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

func listTPMHandles(rwc io.ReadWriteCloser) ([]tpmutil.Handle, error) {
	handlesRaw, _, err := tpm2.GetCapability(rwc, tpm2.CapabilityHandles, 100, uint32(tpm2.PersistentFirst))
	if err != nil {
		return nil, fmt.Errorf("erro ao listar handles: %w", err)
	}

	var handles []tpmutil.Handle
	for _, h := range handlesRaw {
		if handle, ok := h.(tpmutil.Handle); ok {
			handles = append(handles, handle)
		} else {
			log.Printf("Aviso: handle inválido encontrado no TPM: %v", h)
		}
	}

	return handles, nil
}

// Verificar qual handle pertence a uma chave RSA
func findRSAHandle(rwc io.ReadWriteCloser, isSigningKey bool) (tpmutil.Handle, error) {
	handlesRaw, _, err := tpm2.GetCapability(rwc, tpm2.CapabilityHandles, 100, uint32(tpm2.PersistentFirst))
	if err != nil {
		return 0, err
	}

	activeHandles := make(map[tpmutil.Handle]bool)
	for _, h := range handlesRaw {
		if handle, ok := h.(tpmutil.Handle); ok {
			activeHandles[handle] = true
		}
	}

	for handle := range activeHandles {
		pub, _, _, err := tpm2.ReadPublic(rwc, handle)
		if err != nil {
			continue
		}

		if pub.Type == tpm2.AlgRSA {
			if isSigningKey && (pub.Attributes&tpm2.FlagSign != 0) {
				return handle, nil
			} else if !isSigningKey && (pub.Attributes&tpm2.FlagDecrypt != 0) {
				return handle, nil
			}
		}
	}

	return 0, fmt.Errorf("no suitable RSA key found")
}

// InitializeDevice configura o dispositivo pela primeira vez
func (c *TPMClient) InitializeDevice(ctx context.Context) (*types.DeviceInfo, error) {
	log.Println("[InitializeDevice] Iniciando inicialização do dispositivo TPM")
	maxRetries := 1 // Número máximo de tentativas
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		select {
		case <-ctx.Done():
			log.Println("[InitializeDevice] Cancelado pelo contexto")
			return nil, ctx.Err()
		default:
			log.Printf("[InitializeDevice] Tentativa %d/%d\n", i+1, maxRetries+1)

			// Obtendo chave de endosso (EK)
			log.Println("[InitializeDevice] Recuperando Endorsement Key (EK)...")
			ek, err := c.getEndorsementKey(ctx)
			if err != nil {
				lastErr = fmt.Errorf("falha ao recuperar EK: %v", err)
				log.Println("[InitializeDevice] Erro ao recuperar EK:", err)
				continue
			}
			c.ek = ek
			log.Println("[InitializeDevice] EK recuperada com sucesso")

			// Gerando AIK (Attestation Identity Key)
			log.Println("[InitializeDevice] Gerando AIK (Attestation Identity Key)...")
			aik, err := c.generateAIK(ctx)
			if err != nil {
				lastErr = fmt.Errorf("falha ao gerar AIK: %v", err)
				log.Println("[InitializeDevice] Erro ao gerar AIK:", err)
				continue
			}
			c.aik = aik
			log.Println("[InitializeDevice] AIK gerada com sucesso")

			// Buscando handle da chave RSA de assinatura
			log.Println("[InitializeDevice] Buscando handle da chave RSA de assinatura...")
			signHandle, err := findRSAHandle(c.rwc, true)
			if err == nil {
				log.Printf("[InitializeDevice] Chave de assinatura encontrada no handle: 0x%x\n", signHandle)
				c.signHandle = signHandle
			} else {
				log.Println("[InitializeDevice] Chave de assinatura não encontrada, gerando uma nova...")
				_, err := c.generateRSAKeyPair(ctx)
				if err != nil {
					lastErr = fmt.Errorf("falha ao gerar chaves RSA: %v", err)
					log.Println("[InitializeDevice] Erro ao gerar chaves RSA:", err)
					if i < maxRetries {
						log.Printf("[InitializeDevice] Tentando novamente... (%d/%d)\n", i+1, maxRetries)
						continue
					}
					return nil, lastErr
				}
			}

			// Recuperando chave pública RSA
			log.Println("[InitializeDevice] Recuperando chave pública RSA de assinatura...")
			pubKey, err := c.RetrieveRSASignKey(ctx)
			if err != nil {
				lastErr = fmt.Errorf("falha ao recuperar chave RSA: %v", err)
				log.Println("[InitializeDevice] Erro ao recuperar chave RSA:", err)
				continue
			}
			log.Println("[InitializeDevice] Chave pública RSA recuperada com sucesso")
			// Convertendo chave pública para formato PEM
			pubKeyPEM := GetPublicKeyPEM(pubKey)
			log.Println("[InitializeDevice] Chave pública convertida para formato PEM")
			log.Println("PUBKEYPEM %s", pubKeyPEM)

			// Gerando UUID baseado na EK
			log.Println("[InitializeDevice] Gerando UUID baseado na EK...")
			deviceUUID, err := generateTPMBasedUUID(ek)
			if err != nil {
				log.Println("[InitializeDevice] Erro ao gerar UUID:", err)
				return nil, fmt.Errorf("falha ao gerar UUID: %v", err)
			}
			log.Println("[InitializeDevice] UUID gerado com sucesso:", deviceUUID)

			// Retornando informações do dispositivo
			log.Println("[InitializeDevice] Inicialização do dispositivo concluída com sucesso!")
			return &types.DeviceInfo{
				UUID:      deviceUUID,
				EK:        ek,
				AIK:       aik,
				PublicKey: pubKeyPEM,
			}, nil
		}
	}

	log.Println("[InitializeDevice] Falha ao inicializar o dispositivo após todas as tentativas")
	return nil, lastErr
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
func (c *TPMClient) generateRSAKeyPair(ctx context.Context) (*rsa.PublicKey, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		log.Printf("[generateRSAKeyPair] Starting key generation")

		signTemplate := tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM |
				tpm2.FlagFixedParent |
				tpm2.FlagSensitiveDataOrigin |
				tpm2.FlagUserWithAuth |
				tpm2.FlagSign,
			RSAParameters: &tpm2.RSAParams{
				KeyBits:     2048,
				ExponentRaw: 0x10001,
				Sign: &tpm2.SigScheme{
					Alg:  tpm2.AlgRSASSA,
					Hash: tpm2.AlgSHA256,
				},
			},
		}

		log.Printf("[generateRSAKeyPair] Template attributes: 0x%x", signTemplate.Attributes)

		signKeyHandle, signPub, _, _, _, _, err := tpm2.CreatePrimaryEx(
			c.rwc,
			tpm2.HandleOwner,
			tpm2.PCRSelection{},
			"",
			"",
			signTemplate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create signing key: %v", err)
		}
		defer tpm2.FlushContext(c.rwc, signKeyHandle)

		decryptTemplate := tpm2.Public{
			Type:    tpm2.AlgRSA,
			NameAlg: tpm2.AlgSHA256,
			Attributes: tpm2.FlagFixedTPM |
				tpm2.FlagFixedParent |
				tpm2.FlagSensitiveDataOrigin |
				tpm2.FlagUserWithAuth |
				tpm2.FlagDecrypt,
			RSAParameters: &tpm2.RSAParams{
				KeyBits:     2048,
				ExponentRaw: 0x10001,
			},
		}

		decryptKeyHandle, _, _, _, _, _, err := tpm2.CreatePrimaryEx(
			c.rwc,
			tpm2.HandleOwner,
			tpm2.PCRSelection{},
			"",
			"",
			decryptTemplate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create decryption key: %v", err)
		}
		defer tpm2.FlushContext(c.rwc, decryptKeyHandle)

		if err := tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, c.signHandle, c.signHandle); err == nil {
			time.Sleep(100 * time.Millisecond)
		}

		if err := tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, c.decryptHandle, c.decryptHandle); err != nil {
			log.Printf("[generateRSAKeyPair] Warning: Failed to clear existing decrypt handle: %v", err)
		}

		// Make handles persistent with logging
		log.Printf("[generateRSAKeyPair] Persisting signing key to handle 0x%x", c.signHandle)
		if err := tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, signKeyHandle, c.signHandle); err != nil {
			return nil, fmt.Errorf("failed to persist signing handle: %v", err)
		}

		log.Printf("[generateRSAKeyPair] Persisting decryption key to handle 0x%x", c.decryptHandle)
		if err := tpm2.EvictControl(c.rwc, "", tpm2.HandleOwner, decryptKeyHandle, c.decryptHandle); err != nil {
			return nil, fmt.Errorf("failed to persist decryption handle: %v", err)
		}

		rsaPub, err := tpm2.DecodePublic(signPub)
		if err != nil {
			return nil, fmt.Errorf("failed to decode public key: %v", err)
		}

		return &rsa.PublicKey{
			N: rsaPub.RSAParameters.Modulus(),
			E: 65537,
		}, nil
	}
}

func (c *TPMClient) RetrieveRSASignKey(ctx context.Context) (*rsa.PublicKey, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Read public key from signing handle
		pub, _, _, err := tpm2.ReadPublic(c.rwc, c.signHandle)
		if err != nil {
			return nil, fmt.Errorf("failed to read RSA key from TPM: %v", err)
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
		// Using decryptHandle directly, removed reference to rsaHandle
		pub, _, _, err := tpm2.ReadPublic(c.rwc, c.decryptHandle)
		if err != nil {
			return nil, fmt.Errorf("failed to read decryption key: %v", err)
		}
		log.Printf("Atributos da chave: %x", pub.Attributes)

		handles, _, err := tpm2.GetCapability(c.rwc, tpm2.CapabilityHandles, 100, uint32(tpm2.PersistentFirst))
		if err != nil {
			log.Fatalf("Erro ao listar handles: %v", err)
		}
		log.Printf("Handles disponíveis: %v", handles)

		fmt.Printf(string(ciphertext))
		// Log para debug
		log.Printf("Tentando decriptar com handle: %x", c.decryptHandle)
		log.Printf("Tamanho do ciphertext: %d bytes", len(ciphertext))

		// Decripta usando OAEP com SHA256
		decrypted, err := tpm2.RSADecrypt(
			c.rwc,
			c.decryptHandle,
			"", // Sem senha
			ciphertext,
			&tpm2.AsymScheme{
				Alg:  tpm2.AlgOAEP,
				Hash: tpm2.AlgSHA256,
			},
			"",
		)
		if err != nil {
			return nil, fmt.Errorf("erro na decriptação TPM: %w", err)
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
