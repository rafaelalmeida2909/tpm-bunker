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
	"log"
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

func EncryptFile(ctx context.Context, inputFilePath string, pubKey *rsa.PublicKey, tpmMgr *tpm.Manager) (*EncryptionResult, error) {
	fileData, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	encryptedData, encryptedKey, signature, hash, err := encryptInMemory(ctx, fileData, pubKey, tpmMgr)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(inputFilePath)
	filename := inputFilePath[:len(inputFilePath)-len(ext)]
	encryptedFilePath := filename + "_encrypted" + ext

	return &EncryptionResult{
		EncryptedFilePath:     encryptedFilePath,
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

func encryptInMemory(ctx context.Context, data []byte, pubKey *rsa.PublicKey, tpmMgr *tpm.Manager) (encryptedData []byte, encryptedKey []byte, signature []byte, hash [32]byte, err error) {
	// Generate random AES key
	symmetricKey := make([]byte, 32)
	if _, err := rand.Read(symmetricKey); err != nil {
		return nil, nil, nil, hash, fmt.Errorf("error generating symmetric key: %w", err)
	}

	// Encrypt AES key with RSA public key
	encryptedKey, err = rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pubKey,
		symmetricKey,
		nil,
	)
	log.Printf("[Encrypt] Encrypted Symmetric Key (Hex): %x", encryptedKey)
	if err != nil {
		return nil, nil, nil, hash, fmt.Errorf("error encrypting symmetric key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, nil, nil, hash, fmt.Errorf("error creating AES cipher: %w", err)
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, nil, nil, hash, fmt.Errorf("error generating IV: %w", err)
	}

	paddedData := padPKCS7(data, aes.BlockSize)

	// Encrypt data
	encryptedData = make([]byte, len(paddedData))
    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(encryptedData, paddedData)

	// Prepend IV to encrypted data
	encryptedData = append(iv, encryptedData...)

	// Calculate hash
	hash_256 := sha256.Sum256(encryptedData)

	// Sign hash using TPM
	signature, err = tpmMgr.Client.SignData(ctx, hash_256[:])
	if err != nil {
		return nil, nil, nil, hash_256, fmt.Errorf("error signing data: %w", err)
	}

	return encryptedData, encryptedKey, signature, hash, nil
}

// padPKCS7 adiciona padding PKCS7
func padPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}
