package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
)

type EncryptionService struct {
	// Cache de chaves públicas dos clientes
	clientKeys map[string]*rsa.PublicKey
}

func NewEncryptionService() *EncryptionService {
	return &EncryptionService{
		clientKeys: make(map[string]*rsa.PublicKey),
	}
}

// Criptografa um arquivo usando a chave pública do cliente
func (s *EncryptionService) EncryptFile(clientID string, data []byte) ([]byte, error) {
	pubKey, exists := s.clientKeys[clientID]
	if !exists {
		return nil, fmt.Errorf("public key not found for client %s", clientID)
	}

	// Gera uma chave simétrica aleatória
	symmetricKey := make([]byte, 32)
	if _, err := rand.Read(symmetricKey); err != nil {
		return nil, err
	}

	// Criptografa a chave simétrica com a chave pública do cliente
	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pubKey,
		symmetricKey,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Criptografa os dados usando a chave simétrica
	// ... implementação da criptografia AES ...

	return encryptedData, nil
}
