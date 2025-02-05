package main

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"time"
)

type CryptoManager struct {
	rsaPrivateKey *rsa.PrivateKey
	aesKey        []byte
}

func NewCryptoManager() (*CryptoManager, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar chave RSA: %w", err)
	}
	aesKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("erro ao gerar chave AES: %w", err)
	}
	return &CryptoManager{rsaPrivateKey: privateKey, aesKey: aesKey}, nil
}

func (cm *CryptoManager) EncryptData(data []byte) ([]byte, []byte, []byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, nil, fmt.Errorf("erro ao gerar IV: %w", err)
	}

	block, err := aes.NewCipher(cm.aesKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("erro ao criar cipher AES: %w", err)
	}

	padding := aes.BlockSize - (len(data) % aes.BlockSize)
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	encryptedData := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedData, padded)

	encryptedWithIV := append(iv, encryptedData...)
	encryptedKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &cm.rsaPrivateKey.PublicKey, cm.aesKey, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("erro ao encriptar chave simétrica: %w", err)
	}

	hash := sha256.Sum256(encryptedWithIV)
	signature, err := rsa.SignPKCS1v15(rand.Reader, cm.rsaPrivateKey, crypto.SHA256, hash[:])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("erro ao gerar assinatura: %w", err)
	}

	return encryptedWithIV, encryptedKey, signature, nil
}

func (cm *CryptoManager) DecryptData(encryptedData, encryptedKey, signature []byte) ([]byte, error) {
	hash := sha256.Sum256(encryptedData)
	if err := rsa.VerifyPKCS1v15(&cm.rsaPrivateKey.PublicKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, fmt.Errorf("assinatura inválida: %w", err)
	}

	decryptedKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, cm.rsaPrivateKey, encryptedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao decriptar chave simétrica: %w", err)
	}

	if len(encryptedData) < aes.BlockSize {
		return nil, fmt.Errorf("dados encriptados muito curtos")
	}
	iv, encryptedContent := encryptedData[:aes.BlockSize], encryptedData[aes.BlockSize:]
	block, err := aes.NewCipher(decryptedKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cipher AES: %w", err)
	}

decrypted := make([]byte, len(encryptedContent))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, encryptedContent)

	padding := int(decrypted[len(decrypted)-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("padding inválido")
	}
	return decrypted[:len(decrypted)-padding], nil
}

func runBenchmark() {
	sizes := []int64{1, 10, 50, 100, 250, 500, 1024} // MB
	cm, err := NewCryptoManager()
	if err != nil {
		log.Fatalf("Erro ao criar crypto manager: %v", err)
	}

	for _, size := range sizes {
		data := make([]byte, size*1024*1024)
		if _, err := rand.Read(data); err != nil {
			log.Fatalf("Erro ao gerar dados: %v", err)
		}
		
		for i := 0; i < 5; i++ {
			start := time.Now()
			encrypted, encryptedKey, signature, err := cm.EncryptData(data)
			if err != nil {
				log.Fatalf("Erro na encriptação: %v", err)
			}
			duration := time.Since(start)
			throughput := float64(size) / duration.Seconds()
			log.Printf("Encryption - Run %d - Size: %.2f MB, Duration: %v, Throughput: %.2f MB/s", i+1, float64(size), duration, throughput)
			
			start = time.Now()
			_, err = cm.DecryptData(encrypted, encryptedKey, signature)
			if err != nil {
				log.Fatalf("Erro na decriptação: %v", err)
			}
			duration = time.Since(start)
			throughput = float64(size) / duration.Seconds()
			log.Printf("Decryption - Run %d - Size: %.2f MB, Duration: %v, Throughput: %.2f MB/s", i+1, float64(size), duration, throughput)
		}
	}
}

func main() {
	runBenchmark()
}
