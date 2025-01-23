package tpm

import (
	"context"
	"fmt"
	"sync"
	"time"
	"tpm-bunker/internal/types"
)

type Manager struct {
	Client *TPMClient
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc // Adicionado para controle de cancelamento

	// Estado do dispositivo
	DeviceUUID string
	PublicKey  string
	EK         []byte
	AIK        []byte
}

func NewManager(ctx context.Context) *Manager {
	// Cria um contexto cancelável
	ctx, cancel := context.WithCancel(ctx)

	m := &Manager{
		ctx:    ctx,
		cancel: cancel,
	}

	// Tenta inicializar o cliente TPM com timeout
	initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
	defer initCancel()

	client, err := NewTPMClient(initCtx)
	if err != nil {
		fmt.Printf("Aviso: TPM não disponível: %v\n", err)
		return m
	}

	m.Client = client
	return m
}

func (m *Manager) InitializeDevice(ctx context.Context) error {
	// Verifica se o contexto foi cancelado
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.Client == nil {
		return fmt.Errorf("TPM não está disponível neste dispositivo")
	}

	if m.DeviceUUID != "" {
		return fmt.Errorf("dispositivo já inicializado")
	}

	// Cria um timeout específico para inicialização
	initCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Inicializa o dispositivo com context
	creds, err := m.Client.InitializeDevice(initCtx)
	if err != nil {
		return fmt.Errorf("falha na inicialização do dispositivo: %v", err)
	}

	// Verifica novamente o cancelamento antes de atualizar o estado
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.DeviceUUID = creds.UUID
		m.PublicKey = creds.PublicKey
		m.EK = creds.EK
		m.AIK = creds.AIK
		return nil
	}
}

func (m *Manager) GetDeviceUUID(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		m.mutex.RLock()
		defer m.mutex.RUnlock()
		return m.DeviceUUID, nil
	}
}

func (m *Manager) GetStatus(ctx context.Context) (*types.TPMStatus, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		m.mutex.RLock()
		defer m.mutex.RUnlock()

		status := &types.TPMStatus{
			Available:   m.Client != nil,
			Initialized: m.DeviceUUID != "",
		}

		return status, nil
	}
}

func (m *Manager) GetPublicKey(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		m.mutex.RLock()
		defer m.mutex.RUnlock()
		return m.PublicKey, nil
	}
}

// Método para limpar recursos quando não mais necessários
func (m *Manager) Close() {
	if m.cancel != nil {
		m.cancel()
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.Client != nil {
		m.Client.Close()
		m.Client = nil
	}
}
