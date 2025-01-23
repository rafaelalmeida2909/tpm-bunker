package agent

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
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

func EncryptFile(ctx context.Context, inputFilePath string, privateKey *rsa.PrivateKey, tpmMgr *tpm.Manager) (*EncryptionResult, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	log.Printf("Iniciando processo de encriptação para o arquivo: %s", inputFilePath)

	// Lê o arquivo original
	fileData, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	// Gera chave AES aleatória (256 bits)
	symmetricKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, symmetricKey); err != nil {
		return nil, fmt.Errorf("erro ao gerar chave simétrica: %w", err)
	}

	// Gera IV aleatório (16 bytes)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("erro ao gerar IV: %w", err)
	}

	// Padding PKCS7
	paddedData := padPKCS7(fileData, aes.BlockSize)

	// Cria cipher AES-CBC
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %w", err)
	}

	// Encripta dados
	mode := cipher.NewCBCEncrypter(block, iv)
	encryptedContent := make([]byte, len(paddedData))
	mode.CryptBlocks(encryptedContent, paddedData)

	// Combina IV e dados encriptados em um novo slice
	encryptedData := make([]byte, len(iv)+len(encryptedContent))
	copy(encryptedData[0:aes.BlockSize], iv)
	copy(encryptedData[aes.BlockSize:], encryptedContent)

	// Recupera a chave pública do TPM
	pubKey, err := tpmMgr.Client.RetrieveRSAKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar chave pública: %w", err)
	}

	runtime.LogInfo(ctx, fmt.Sprintf("PublicKeyPEM: %s", tpm.GetPublicKeyPEM(pubKey)))

	// Encripta a chave simétrica com RSA-OAEP
	encryptedSymmetricKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pubKey,
		symmetricKey,
		nil, // label
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao encriptar chave simétrica: %w", err)
	}

	// Gera o hash para assinatura
	hash := sha256.Sum256(encryptedData)
	log.Printf("Hash calculado no cliente (hex): %x", hash[:])
	// Assina usando o TPM
	signature, err := tpmMgr.Client.SignData(ctx, hash[:])
	if err != nil {
		return nil, fmt.Errorf("falha na assinatura: %w", err)
	}

	// Gera nome do arquivo encriptado
	ext := ".bin"
	filename := inputFilePath[:len(inputFilePath)-len(ext)]
	encryptedFilePath := filename + "_encrypted" + ext
	runtime.LogInfo(ctx, fmt.Sprintf("filename... %s", filename))
	runtime.LogInfo(ctx, fmt.Sprintf("FilePath %s", encryptedFilePath))

	// Salva arquivo encriptado
	if err := os.WriteFile(encryptedFilePath, encryptedData, 0644); err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo encriptado: %w", err)
	}

	// Calcula o hash original

	return &EncryptionResult{
		EncryptedFilePath:     encryptedFilePath,
		EncryptedSymmetricKey: base64.StdEncoding.EncodeToString(encryptedSymmetricKey),
		DigitalSignature:      base64.StdEncoding.EncodeToString(signature),
		HashOriginal:          base64.StdEncoding.EncodeToString(hash[:]),
		Metadata: map[string]string{
			"filename":  filepath.Base(inputFilePath),
			"version":   "1.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"algorithm": "AES-256-CBC",
		},
	}, nil
}

// padPKCS7 adiciona padding PKCS7
func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
