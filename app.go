package main

import (
	"context"
	"tpm-bunker/internal/agent"
	"tpm-bunker/internal/api"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"
)

// App struct
type App struct {
	ctx   context.Context
	agent *agent.Agent
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup é chamado quando o app inicia
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	tpmMgr := tpm.NewManager(ctx)
	client := api.NewAPIClient()
	a.agent = agent.NewAgent(ctx, tpmMgr, client)
}

// GetTPMStatus retorna o status atual do TPM
func (a *App) GetTPMStatus() (*types.TPMStatus, error) {
	return a.agent.GetTPMStatus()
}

// InitializeDevice inicializa o dispositivo com TPM
func (a *App) InitializeDevice() (*types.DeviceInfo, error) {
	return a.agent.InitializeDevice()
}

// IsDeviceInitialized verifica se o dispositivo está inicializado
func (a *App) IsDeviceInitialized() bool {
	return a.agent.IsDeviceInitialized()
}

// GetDeviceInfo retorna as informações do dispositivo
func (a *App) GetDeviceInfo() (*types.DeviceInfo, error) {
	return a.agent.GetDeviceInfo()
}

// CheckTPMPresence verifica se o TPM está presente
func (a *App) CheckTPMPresence() bool {
	return a.agent.CheckTPMPresence()
}

// CheckConnection verifica a conexão com a API
func (a *App) CheckConnection() bool {
	return a.agent.CheckConnection()
}

// CheckConnection tenta realizar login na API
func (a *App) AuthLogin() bool {
	return a.agent.AuthLogin()
}

// shutdown é chamado quando o app é fechado
func (a *App) shutdown(ctx context.Context) {
	// Cleanup é feito automaticamente pela linguagem
}
