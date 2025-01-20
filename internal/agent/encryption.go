package agent

import (
	"bytes"
	"crypto"
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
)

type EncryptionResult struct {
	EncryptedFilePath     string            `json:"encrypted_file_path"`
	EncryptedSymmetricKey string            `json:"encrypted_symmetric_key"`
	DigitalSignature      string            `json:"digital_signature"`
	Metadata              map[string]string `json:"metadata"`
}

func EncryptFile(inputFilePath string, privateKey *rsa.PrivateKey) (*EncryptionResult, error) {
	// Lê o arquivo original
	fileData, err := os.ReadFile(inputFilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	// Gera chave AES aleatória
	symmetricKey := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, symmetricKey); err != nil {
		return nil, fmt.Errorf("erro ao gerar chave simétrica: %w", err)
	}

	// Gera IV aleatório
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("erro ao gerar IV: %w", err)
	}

	// Cria cipher AES
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %w", err)
	}

	// Padding PKCS7
	paddedData := padPKCS7(fileData, aes.BlockSize)

	// Encripta dados com AES-CBC
	mode := cipher.NewCBCEncrypter(block, iv)
	encryptedContent := make([]byte, len(paddedData))
	mode.CryptBlocks(encryptedContent, paddedData)

	// Combina IV com conteúdo encriptado
	encryptedData := append(iv, encryptedContent...)

	// Encripta a chave simétrica com RSA
	publicKey := &privateKey.PublicKey
	encryptedSymmetricKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		publicKey,
		symmetricKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao encriptar chave simétrica: %w", err)
	}

	// Gera assinatura digital
	hashed := sha256.Sum256(encryptedData)
	signature, err := rsa.SignPSS(
		rand.Reader,
		privateKey,
		crypto.SHA256,
		hashed[:],
		&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto},
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar assinatura: %w", err)
	}

	// Verifica assinatura localmente
	err = rsa.VerifyPSS(
		&privateKey.PublicKey,
		crypto.SHA256,
		hashed[:],
		signature,
		&rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto},
	)
	if err != nil {
		return nil, fmt.Errorf("erro na verificação local da assinatura: %w", err)
	}

	// Gera nome do arquivo encriptado
	ext := filepath.Ext(inputFilePath)
	filename := inputFilePath[:len(inputFilePath)-len(ext)]
	encryptedFilePath := filename + "_encrypted" + ext

	// Salva arquivo encriptado
	if err := os.WriteFile(encryptedFilePath, encryptedData, 0644); err != nil {
		return nil, fmt.Errorf("erro ao salvar arquivo encriptado: %w", err)
	}

	return &EncryptionResult{
		EncryptedFilePath:     encryptedFilePath,
		EncryptedSymmetricKey: base64.StdEncoding.EncodeToString(encryptedSymmetricKey),
		DigitalSignature:      base64.StdEncoding.EncodeToString(signature),
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
