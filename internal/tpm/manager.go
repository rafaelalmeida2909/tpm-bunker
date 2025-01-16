package tpm

import (
	"context"
	"fmt"
	"sync"
	"tpm-bunker/internal/types"
)

// Manager gerencia as operações do TPM e mantém o estado
type Manager struct {
	client *TPMClient
	mutex  sync.RWMutex
	ctx    context.Context

	// Estado do dispositivo
	deviceUUID string
	publicKey  string
	EK         []byte
	AIK        []byte
}

// NewManager cria uma nova instância do gerenciador TPM
func NewManager(ctx context.Context) *Manager {
	m := &Manager{
		ctx: ctx,
	}

	// Tenta inicializar o cliente TPM
	client, err := NewTPMClient()
	if err != nil {
		// Apenas loga o erro e continua com client nulo
		fmt.Printf("Aviso: TPM não disponível: %v\n", err)
		return m
	}

	m.client = client
	return m
}

// InitializeDevice realiza a configuração inicial do dispositivo
func (m *Manager) InitializeDevice() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Verifica se o TPM está disponível
	if m.client == nil {
		return fmt.Errorf("TPM não está disponível neste dispositivo")
	}

	// Verifica se já foi inicializado
	if m.deviceUUID != "" {
		return fmt.Errorf("dispositivo já inicializado")
	}

	// Inicializa o dispositivo através do client
	creds, err := m.client.InitializeDevice()
	if err != nil {
		return fmt.Errorf("falha na inicialização do dispositivo: %v", err)
	}

	// Armazena as credenciais relevantes
	m.deviceUUID = creds.UUID
	m.publicKey = creds.PublicKey
	m.EK = creds.EK
	m.AIK = creds.AIK

	return nil
}

// IsInitialized verifica se o dispositivo está inicializado
func (m *Manager) IsInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.deviceUUID != ""
}

// GetDeviceUUID retorna o UUID do dispositivo
func (m *Manager) GetDeviceUUID() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.deviceUUID
}

// GetStatus retorna o status atual do TPM
func (m *Manager) GetStatus() (*types.TPMStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	status := &types.TPMStatus{
		Available:   m.client != nil,
		Initialized: m.deviceUUID != "",
	}

	return status, nil
}

// GetPublicKey retorna a chave pública RSA do dispositivo
func (m *Manager) GetPublicKey() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.publicKey
}
