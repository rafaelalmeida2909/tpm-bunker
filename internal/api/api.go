package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

func NewAPIClient() *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "http://localhost:8003/api/v1/",
	}
}

func (c *APIClient) CheckConnection() bool {
	log.Printf("Verificando conexão com API em: %s", c.baseURL)

	req, err := http.NewRequest(http.MethodHead, c.baseURL, nil)
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

	// Log do status code recebido
	log.Printf("Resposta da API - Status Code: %d", resp.StatusCode)

	// Verifica se o status code está na faixa 2xx
	isConnected := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !isConnected {
		log.Printf("Conexão falhou: API retornou status code inválido: %d", resp.StatusCode)
	} else {
		log.Printf("Conexão com API estabelecida com sucesso")
	}

	return isConnected
}

func (c *APIClient) IsDeviceRegistered(uuid string) (bool, error) {
	// Tenta obter o dispositivo pelo UUID
	response, err := c.SendRequest(http.MethodGet, fmt.Sprintf("devices/%s/", uuid), nil)
	if err != nil {
		// Se retornar 404, significa que o dispositivo não está registrado
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, fmt.Errorf("erro ao verificar registro do dispositivo: %w", err)
	}

	return response != nil, nil
}

func (c *APIClient) RegisterDevice(deviceInfo *types.DeviceInfo) error {
	// Prepara os dados para registro
	registration := DeviceRegistration{
		UUID:      deviceInfo.UUID,
		EKCert:    base64.StdEncoding.EncodeToString(deviceInfo.EK),  // Converte para base64
		AIK:       base64.StdEncoding.EncodeToString(deviceInfo.AIK), // Converte para base64
		PublicKey: deviceInfo.PublicKey,
	}

	// Envia a requisição para registrar o dispositivo
	_, err := c.SendRequest(http.MethodPost, "devices/", registration)
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

func (c *APIClient) Login(uuid string, ekCert []byte) error {
	// Prepara os dados para login
	loginData := LoginRequest{
		UUID:   uuid,
		EKCert: base64.StdEncoding.EncodeToString(ekCert),
	}

	// Faz a requisição de login
	response, err := c.SendRequest(http.MethodPost, "auth/login/", loginData)
	if err != nil {
		return fmt.Errorf("falha ao realizar login: %w", err)
	}

	// Decodifica a resposta
	var loginResponse LoginResponse
	if err := json.Unmarshal(response, &loginResponse); err != nil {
		return fmt.Errorf("falha ao processar resposta do login: %w", err)
	}

	// Armazena o token para futuras requisições
	c.setAuthToken(loginResponse.Token)

	log.Printf("Login realizado com sucesso para o dispositivo: %s", uuid)
	return nil
}

// setAuthToken configura o token de autenticação para futuras requisições
func (c *APIClient) setAuthToken(token string) {
	c.authToken = token
}

// Modifica SendRequest para incluir o token quando disponível
func (c *APIClient) SendRequest(method string, endpoint string, data interface{}) ([]byte, error) {
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

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Adiciona o token de autenticação se disponível
	if c.authToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar requisição: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("requisição falhou com status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
