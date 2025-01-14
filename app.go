package main

import (
	"context"
	"fmt"
	"tpm-bunker/internal/agent"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	tpmMgr *tpm.Manager
	agent  *agent.Agent
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup é chamado quando o app inicia
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Agora NewManager sempre retorna um manager, mesmo que o TPM não esteja disponível
	a.tpmMgr = tpm.NewManager(ctx)

	// Inicializa o agente apenas se necessário
	status, _ := a.tpmMgr.GetStatus()
	if status.Available {
		a.agent = agent.NewAgent(ctx, a.tpmMgr)
	}
}

// GetTPMStatus retorna o status atual do TPM
func (a *App) GetTPMStatus() (*tpm.TPMStatus, error) {
	return a.tpmMgr.GetStatus()
}

// InitializeDevice inicializa o dispositivo com TPM pela primeira vez
func (a *App) InitializeDevice() (*types.DeviceInfo, error) {
	status, _ := a.tpmMgr.GetStatus()
	if !status.Available {
		return nil, fmt.Errorf("TPM não está disponível neste dispositivo")
	}

	err := a.tpmMgr.InitializeDevice()
	if err != nil {
		return nil, err
	}

	return &types.DeviceInfo{
		UUID:      a.tpmMgr.GetDeviceUUID(),
		PublicKey: a.tpmMgr.GetPublicKey(),
	}, nil
}

// IsDeviceInitialized verifica se o dispositivo já foi inicializado
func (a *App) IsDeviceInitialized() bool {
	status, _ := a.tpmMgr.GetStatus()
	return status.Initialized
}

// GetDeviceInfo retorna as informações do dispositivo
func (a *App) GetDeviceInfo() (*types.DeviceInfo, error) {
	if a.tpmMgr == nil {
		return nil, fmt.Errorf("TPM Manager não inicializado")
	}

	uuid := a.tpmMgr.GetDeviceUUID()
	pubKey := a.tpmMgr.GetPublicKey()

	runtime.LogInfo(a.ctx, fmt.Sprintf("GetDeviceInfo - UUID: %s", uuid))

	return &types.DeviceInfo{
		UUID:      uuid,
		PublicKey: pubKey,
	}, nil
}

// ExecuteOperation executa uma operação via agente
func (a *App) ExecuteOperation(op types.UserOperation) (*types.APIResponse, error) {
	if a.tpmMgr == nil || a.agent == nil {
		return nil, fmt.Errorf("sistema não inicializado")
	}

	runtime.LogInfo(a.ctx, fmt.Sprintf("ExecuteOperation - Type: %s", op.Type))
	return &types.APIResponse{
		Success: true,
		Message: "Operação não implementada",
	}, nil
}

// shutdown é chamado quando o app é fechado
func (a *App) shutdown(ctx context.Context) {
	// Limpa recursos se necessário
}
