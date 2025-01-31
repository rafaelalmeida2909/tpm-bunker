package agent

import (
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"
)

type DecryptionResult struct {
	DecryptedData []byte
	Verified      bool
}

func DecryptFile(ctx context.Context, decryptResp *types.DecryptResponse, tpmMgr *tpm.Manager) (*DecryptionResult, error) {
	// Verify digital signature first
	hash := sha256.Sum256(decryptResp.EncryptedData)
	signature, err := base64.StdEncoding.DecodeString(decryptResp.DigitalSignature)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar assinatura: %w", err)
	}

	// Get public key for verification
	pubKey, err := tpmMgr.Client.RetrieveRSAKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao recuperar chave pública: %w", err)
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature)
	if err != nil {
		return nil, fmt.Errorf("assinatura digital inválida: %w", err)
	}

	// Decrypt the data
	decryptedData, err := decryptInMemory(ctx, decryptResp.EncryptedData, decryptResp.EncryptedSymmetricKey, tpmMgr)
	if err != nil {
		return nil, fmt.Errorf("erro na decriptação: %w", err)
	}

	return &DecryptionResult{
		DecryptedData: decryptedData,
		Verified:      true,
	}, nil
}

func decryptInMemory(ctx context.Context, encryptedData, encryptedKey []byte, tpmMgr *tpm.Manager) ([]byte, error) {
	// Extract IV from encrypted data
	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("dados encriptados muito curtos")
	}
	iv := encryptedData[:aes.BlockSize]
	encryptedContent := encryptedData[aes.BlockSize:]

	// Decrypt symmetric key using TPM
	symmetricKey, err := tpmMgr.Client.RSADecrypt(ctx, encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao decriptar chave simétrica: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %w", err)
	}

	// Decrypt content
	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, len(encryptedContent))
	mode.CryptBlocks(decryptedData, encryptedContent)

	// Remove PKCS7 padding
	unpadded, err := unpadPKCS7(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("erro ao remover padding: %w", err)
	}

	return unpadded, nil
}

// unpadPKCS7 removes PKCS7 padding
func unpadPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("dados vazios")
	}

	padding := int(data[len(data)-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("padding inválido")
	}

	// Verify padding is valid
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("padding inconsistente")
		}
	}

	return data[:len(data)-padding], nil
}
