package agent

import (
	"context"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types" // ajuste o import

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Agent struct {
	ctx    context.Context
	tpmMgr *tpm.Manager
}

func NewAgent(ctx context.Context, tpmMgr *tpm.Manager) *Agent {
	return &Agent{
		ctx:    ctx,
		tpmMgr: tpmMgr,
	}
}

func (a *Agent) Initialize() error {
	status, err := a.tpmMgr.GetStatus()
	if err != nil {
		return err
	}

	runtime.EventsEmit(a.ctx, "system:status", status)
	return nil
}

func (a *Agent) ExecuteOperation(op types.UserOperation) (*types.APIResponse, error) {
	// Orquestrar operação entre TPM e API
	// 1. Preparar TPM
	// 2. Fazer request para API
	// 3. Processar resposta
	// 4. Atualizar TPM se necessário
	return a.tpmMgr.HandleOperation(op)
}
