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
	"os"
	"path/filepath"
	"time"
	"tpm-bunker/internal/tpm"
)

type EncryptionResult struct {
	EncryptedFilePath     string            `json:"encrypted_file_path"`
	EncryptedSymmetricKey string            `json:"encrypted_symmetric_key"`
	DigitalSignature      string            `json:"digital_signature"`
	HashOriginal          string            `json:"hash_original"`
	Metadata              map[string]string `json:"metadata"`
	EncryptedData         []byte
}

func EncryptFile(ctx context.Context, inputFilePath string, privateKey *rsa.PrivateKey, tpmMgr *tpm.Manager) (*EncryptionResult, error) {
	// Lê arquivo
	fileData, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	// Criptografa em memória
	encryptedData, encryptedKey, signature, hash, err := encryptInMemory(ctx, fileData, tpmMgr)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(inputFilePath)
	filename := inputFilePath[:len(inputFilePath)-len(ext)]
	encryptedFilePath := filename + "_encrypted" + ext

	return &EncryptionResult{
		EncryptedFilePath:     encryptedFilePath, // Dados já criptografados prontos para API
		EncryptedSymmetricKey: base64.StdEncoding.EncodeToString(encryptedKey),
		DigitalSignature:      base64.StdEncoding.EncodeToString(signature),
		HashOriginal:          base64.StdEncoding.EncodeToString(hash[:]),
		EncryptedData:         encryptedData,
		Metadata: map[string]string{
			"filename":  filepath.Base(inputFilePath),
			"version":   "1.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"algorithm": "AES-256-CBC",
		},
	}, nil
}

func encryptInMemory(ctx context.Context, fileData []byte, tpmMgr *tpm.Manager) (encryptedData, encryptedKey, signature []byte, hash [32]byte, err error) {
	// Gera chave AES
	symmetricKey := make([]byte, 32)
	if _, err = io.ReadFull(rand.Reader, symmetricKey); err != nil {
		return nil, nil, nil, hash, fmt.Errorf("erro ao gerar chave simétrica: %w", err)
	}

	// Gera IV
	iv := make([]byte, aes.BlockSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, nil, hash, fmt.Errorf("erro ao gerar IV: %w", err)
	}

	// Padding e criptografia
	paddedData := padPKCS7(fileData, aes.BlockSize)
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, nil, nil, hash, fmt.Errorf("erro ao criar cipher AES: %w", err)
	}

	// Encripta dados
	mode := cipher.NewCBCEncrypter(block, iv)
	encryptedContent := make([]byte, len(paddedData))
	mode.CryptBlocks(encryptedContent, paddedData)

	// Combina IV e dados
	encryptedData = make([]byte, len(iv)+len(encryptedContent))
	copy(encryptedData[0:aes.BlockSize], iv)
	copy(encryptedData[aes.BlockSize:], encryptedContent)

	// Recupera chave pública
	pubKey, err := tpmMgr.Client.RetrieveRSAKey(ctx)
	if err != nil {
		return nil, nil, nil, hash, fmt.Errorf("erro ao recuperar chave pública: %w", err)
	}

	// Encripta chave simétrica
	encryptedKey, err = rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pubKey,
		symmetricKey,
		nil,
	)
	if err != nil {
		return nil, nil, nil, hash, fmt.Errorf("erro ao encriptar chave simétrica: %w", err)
	}

	// Hash e assinatura
	hash = sha256.Sum256(encryptedData)
	signature, err = tpmMgr.Client.SignData(ctx, hash[:])
	if err != nil {
		return nil, nil, nil, hash, fmt.Errorf("falha na assinatura: %w", err)
	}

	return encryptedData, encryptedKey, signature, hash, nil
}

// padPKCS7 adiciona padding PKCS7
func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
