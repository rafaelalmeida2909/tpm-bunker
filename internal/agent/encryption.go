package agent

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"time"
	"tpm-bunker/internal/tpm"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type EncryptionResult struct {
	EncryptedFilePath     string            `json:"encrypted_file_path"`
	EncryptedSymmetricKey string            `json:"encrypted_symmetric_key"`
	DigitalSignature      string            `json:"digital_signature"`
	HashOriginal          string            `json:"hash_original"`
	Metadata              map[string]string `json:"metadata"`
}

func signDataSafe(ctx context.Context, privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	runtime.LogInfo(ctx, "Iniciando assinatura segura")

	// Verifica a chave
	if privateKey == nil || privateKey.D == nil || privateKey.PublicKey.N == nil {
		return nil, fmt.Errorf("chave privada inválida")
	}

	// Cria um novo hash dos dados
	hash := sha256.Sum256(data)
	runtime.LogInfo(ctx, fmt.Sprintf("Hash calculado, tamanho: %d bytes", len(hash)))

	// Canal para resultado com buffer
	resultChan := make(chan struct {
		signature []byte
		err       error
	}, 1)

	// Executa a assinatura em uma goroutine isolada
	go func() {
		// Copia a chave privada para evitar acesso concorrente
		privKeyClone := &rsa.PrivateKey{
			PublicKey: rsa.PublicKey{
				N: new(big.Int).Set(privateKey.PublicKey.N),
				E: privateKey.PublicKey.E,
			},
			D:      new(big.Int).Set(privateKey.D),
			Primes: make([]*big.Int, len(privateKey.Primes)),
		}
		for i, prime := range privateKey.Primes {
			privKeyClone.Primes[i] = new(big.Int).Set(prime)
		}

		// Tenta a assinatura em memória
		sig, err := func() ([]byte, error) {
			defer func() {
				if r := recover(); r != nil {
					runtime.LogError(ctx, fmt.Sprintf("Recuperado de panic durante assinatura: %v", r))
				}
			}()

			opts := &rsa.PSSOptions{
				SaltLength: 32, // Usando um valor fixo
				Hash:       crypto.SHA256,
			}

			return rsa.SignPSS(rand.Reader, privKeyClone, crypto.SHA256, hash[:], opts)
		}()

		resultChan <- struct {
			signature []byte
			err       error
		}{sig, err}
	}()

	// Aguarda com timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		if result.err != nil {
			return nil, fmt.Errorf("erro na assinatura: %w", result.err)
		}
		runtime.LogInfo(ctx, fmt.Sprintf("Assinatura concluída, tamanho: %d bytes", len(result.signature)))
		return result.signature, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout durante assinatura")
	}
}

func EncryptFile(ctx context.Context, inputFilePath string, privateKey *rsa.PrivateKey) (*EncryptionResult, error) {
	if privateKey == nil {
		return nil, fmt.Errorf("chave privada não inicializada")
	}

	if privateKey.PublicKey.N == nil {
		return nil, fmt.Errorf("chave privada inválida (módulo nulo)")
	}

	// Verifica cancelamento antes de começar
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Lê o arquivo original
	fileData, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	runtime.LogInfo(ctx, fmt.Sprintf("File Data... %v", fileData))

	// Gera chave AES aleatória
	symmetricKey := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, symmetricKey); err != nil {
		return nil, fmt.Errorf("erro ao gerar chave simétrica: %w", err)
	}

	runtime.LogInfo(ctx, fmt.Sprint("symmetricKey... %s", symmetricKey))

	// Gera IV aleatório
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("erro ao gerar IV: %w", err)
	}

	runtime.LogInfo(ctx, fmt.Sprint("IV... %s", iv))

	// Verifica cancelamento antes da encriptação
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Cria cipher AES
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %w", err)
	}

	runtime.LogInfo(ctx, fmt.Sprint("Block... %s", block))

	// Padding PKCS7
	paddedData := padPKCS7(fileData, aes.BlockSize)

	runtime.LogInfo(ctx, fmt.Sprint("paddedData... %s", paddedData))

	// Encripta dados com AES-CBC
	mode := cipher.NewCBCEncrypter(block, iv)
	encryptedContent := make([]byte, len(paddedData))
	mode.CryptBlocks(encryptedContent, paddedData)

	runtime.LogInfo(ctx, fmt.Sprint("encryptedContent... %s", encryptedContent))

	// Combina IV com conteúdo encriptado
	encryptedData := append(iv, encryptedContent...)

	runtime.LogInfo(ctx, fmt.Sprint("encryptedData... %s", encryptedData))

	// Verifica cancelamento antes da encriptação RSA
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Encripta a chave simétrica com RSA
	publicKey := &privateKey.PublicKey
	encryptedSymmetricKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		symmetricKey,
		nil,
	)
	runtime.LogInfo(ctx, fmt.Sprint("publicKey... %s", tpm.GetPublicKeyPEM(publicKey)))
	if err != nil {
		return nil, fmt.Errorf("erro ao encriptar chave simétrica: %w", err)
	}

	// Verifica cancelamento antes da assinatura
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Gera assinatura digital
	runtime.LogInfo(ctx, "Iniciando processo de assinatura")
	signature, err := signDataSafe(ctx, privateKey, encryptedData)
	if err != nil {
		runtime.LogError(ctx, fmt.Sprintf("Erro na assinatura: %v", err))
		return nil, fmt.Errorf("falha na assinatura: %w", err)
	}

	runtime.LogInfo(ctx, "Verificando assinatura")
	hashedData := sha256.Sum256(encryptedData)
	err = rsa.VerifyPSS(
		&privateKey.PublicKey,
		crypto.SHA256,
		hashedData[:],
		signature,
		&rsa.PSSOptions{
			SaltLength: 32,
			Hash:       crypto.SHA256,
		},
	)
	if err != nil {
		runtime.LogError(ctx, fmt.Sprintf("Erro na verificação: %v", err))
		return nil, fmt.Errorf("falha na verificação: %w", err)
	}

	runtime.LogInfo(ctx, "Assinatura verificada com sucesso")

	// Verifica cancelamento antes de salvar o arquivo
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Gera nome do arquivo encriptado
	ext := filepath.Ext(inputFilePath)
	runtime.LogInfo(ctx, fmt.Sprint("ext... %s", ext))
	filename := inputFilePath[:len(inputFilePath)-len(ext)]
	runtime.LogInfo(ctx, fmt.Sprint("filename... %s", filename))
	encryptedFilePath := filename + "_encrypted" + ext
	runtime.LogInfo(ctx, fmt.Sprint("encryptedFilePath... %s", encryptedFilePath))

	// Salva arquivo encriptado
	if err := os.WriteFile(encryptedFilePath, encryptedData, 0644); err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo encriptado: %w", err)
	}

	// Calcula o hash original
	hashOriginal := base64.StdEncoding.EncodeToString(hashedData[:])

	return &EncryptionResult{
		EncryptedFilePath:     encryptedFilePath,
		EncryptedSymmetricKey: base64.StdEncoding.EncodeToString(encryptedSymmetricKey),
		DigitalSignature:      base64.StdEncoding.EncodeToString(signature),
		HashOriginal:          hashOriginal,
		Metadata: map[string]string{
			"filename": filepath.Base(inputFilePath),
		},
	}, nil
}

// padPKCS7 adiciona padding PKCS7
func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
