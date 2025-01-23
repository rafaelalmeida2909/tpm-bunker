package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"tpm-bunker/internal/types"
)

type APIClient struct {
	client    *http.Client
	baseURL   string
	authToken string
}

type DeviceRegistration struct {
	UUID      string `json:"uuid"`
	EKCert    string `json:"ek_certificate"`
	AIK       string `json:"aik"`
	PublicKey string `json:"public_key"`
}

type EncryptionRequest struct {
	EncryptedData    string            `json:"encrypted_data"`
	EncryptedKey     string            `json:"encrypted_symmetric_key "`
	DigitalSignature string            `json:"digital_signature"`
	HashOriginal     string            `json:"hash_original"`
	Metadata         map[string]string `json:"metadata"`
}

type EncryptionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	FileID  string `json:"file_id"`
}

func NewAPIClient(ctx context.Context) *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "http://localhost:8003/api/v1/",
	}
}

func (c *APIClient) CheckConnection(ctx context.Context) bool {
	log.Printf("Verificando conexão com API em: %s", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL, nil)
	if err != nil {
		log.Printf("Erro ao criar request para API: %v", err)
		return false
	}

	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("Erro ao conectar com API: %v", err)
		return false
	}
	defer resp.Body.Close()

	isConnected := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !isConnected {
		log.Printf("Conexão falhou: API retornou status code inválido: %d", resp.StatusCode)
	} else {
		log.Printf("Conexão com API estabelecida com sucesso")
	}

	return isConnected
}
func (c *APIClient) IsDeviceRegistered(ctx context.Context, uuid string) (bool, error) {
	response, err := c.SendRequest(ctx, http.MethodGet, fmt.Sprintf("devices/%s/", uuid), nil, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("erro ao verificar registro do dispositivo: %w", err)
	}

	return response != nil, nil
}

func (c *APIClient) RegisterDevice(ctx context.Context, deviceInfo *types.DeviceInfo) error {
	registration := DeviceRegistration{
		UUID:      deviceInfo.UUID,
		EKCert:    base64.StdEncoding.EncodeToString(deviceInfo.EK),
		AIK:       base64.StdEncoding.EncodeToString(deviceInfo.AIK),
		PublicKey: deviceInfo.PublicKey,
	}

	_, err := c.SendRequest(ctx, http.MethodPost, "devices/", nil, registration)
	if err != nil {
		return fmt.Errorf("falha ao registrar dispositivo: %w", err)
	}

	log.Printf("Dispositivo registrado com sucesso. UUID: %s", deviceInfo.UUID)
	return nil
}

type LoginRequest struct {
	UUID   string `json:"uuid"`
	EKCert string `json:"ek_certificate"`
}

// LoginResponse representa a resposta do login
type LoginResponse struct {
	Token string `json:"token"`
}

func (c *APIClient) Login(ctx context.Context, uuid string, ekCert []byte) error {
	loginData := LoginRequest{
		UUID:   uuid,
		EKCert: base64.StdEncoding.EncodeToString(ekCert),
	}

	response, err := c.SendRequest(ctx, http.MethodPost, "auth/login/", nil, loginData)
	if err != nil {
		return fmt.Errorf("falha ao realizar login: %w", err)
	}

	var loginResponse LoginResponse
	if err := json.Unmarshal(response, &loginResponse); err != nil {
		return fmt.Errorf("falha ao processar resposta do login: %w", err)
	}

	c.setAuthToken(loginResponse.Token)
	log.Printf("Login realizado com sucesso para o dispositivo: %s", uuid)
	return nil
}

// setAuthToken configura o token de autenticação para futuras requisições
func (c *APIClient) setAuthToken(token string) {
	c.authToken = token
}

func (c *APIClient) SendRequest(ctx context.Context, method string, endpoint string, headers map[string]string, data interface{}) ([]byte, error) {
	url := c.baseURL + endpoint

	var body *bytes.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar dados: %w", err)
		}
		body = bytes.NewReader(jsonData)
	} else {
		body = bytes.NewReader(nil)
	}

	// Usa NewRequestWithContext ao invés de NewRequest
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Usa select para permitir cancelamento via context
	respChan := make(chan struct {
		resp *http.Response
		err  error
	})

	go func() {
		resp, err := c.client.Do(req)
		respChan <- struct {
			resp *http.Response
			err  error
		}{resp, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-respChan:
		if result.err != nil {
			return nil, fmt.Errorf("erro ao enviar requisição: %w", result.err)
		}
		defer result.resp.Body.Close()

		respBody, err := io.ReadAll(result.resp.Body)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler resposta: %w", err)
		}

		if result.resp.StatusCode < 200 || result.resp.StatusCode > 299 {
			return nil, fmt.Errorf("requisição falhou com status %d: %s", result.resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}
}

func (c *APIClient) EncryptRequest(ctx context.Context, method string, endpoint string, headers map[string]string, data interface{}) ([]byte, error) {
	// Verify the data is of the correct type
	payload, ok := data.(*EncryptionRequest)
	if !ok {
		return nil, fmt.Errorf("dados inválidos: esperado *api.EncryptionRequest")
	}

	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add encrypted file
	encryptedFile, err := os.Open(payload.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo criptografado: %w", err)
	}
	defer encryptedFile.Close()

	part, err := writer.CreateFormFile("encrypted_data", filepath.Base(payload.EncryptedData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar parte do formulário: %w", err)
	}
	_, err = io.Copy(part, encryptedFile)
	if err != nil {
		return nil, fmt.Errorf("erro ao copiar arquivo: %w", err)
	}

	// Add other form fields
	_ = writer.WriteField("encrypted_symmetric_key", payload.EncryptedKey)
	_ = writer.WriteField("digital_signature", payload.DigitalSignature)
	_ = writer.WriteField("hash_original", payload.HashOriginal)

	// Convert metadata to JSON
	metadataJSON, _ := json.Marshal(payload.Metadata)
	_ = writer.WriteField("metadata", string(metadataJSON))

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("erro ao fechar writer: %w", err)
	}

	// Construct full URL
	url := c.baseURL + endpoint

	// Set Content-Type to multipart/form-data
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	log.Printf("API KEY: %s", c.authToken)

	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Usa select para permitir cancelamento via context
	respChan := make(chan struct {
		resp *http.Response
		err  error
	})

	go func() {
		resp, err := c.client.Do(req)
		respChan <- struct {
			resp *http.Response
			err  error
		}{resp, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-respChan:
		if result.err != nil {
			return nil, fmt.Errorf("erro ao enviar requisição: %w", result.err)
		}
		defer result.resp.Body.Close()

		respBody, err := io.ReadAll(result.resp.Body)
		if err != nil {
			return nil, fmt.Errorf("erro ao ler resposta: %w", err)
		}

		if result.resp.StatusCode < 200 || result.resp.StatusCode > 299 {
			return nil, fmt.Errorf("requisição falhou com status %d: %s", result.resp.StatusCode, string(respBody))
		}

		return respBody, nil
	}
}
