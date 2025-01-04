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

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.tpmMgr = tpm.NewManager(ctx)
	a.agent = agent.NewAgent(ctx, a.tpmMgr)

	if err := a.agent.Initialize(); err != nil {
		runtime.LogError(ctx, fmt.Sprintf("Falha na inicialização: %v", err))
	}
}

// GetTPMStatus retorna o status do TPM
func (a *App) GetTPMStatus() (*types.TPMStatus, error) {
	return a.tpmMgr.GetStatus()
}

// ExecuteOperation executa uma operação via agente
func (a *App) ExecuteOperation(op types.UserOperation) (*types.APIResponse, error) {
	return a.agent.ExecuteOperation(op)
}
