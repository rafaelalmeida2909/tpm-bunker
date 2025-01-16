package agent

import (
	"context"

	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"

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

func (a *Agent) Initialize() string {
	//status, err := a.tpmMgr.GetStatus()
	err := "teste"
	if err != "" {
		return err
	}

	runtime.EventsEmit(a.ctx, "system:status", "status")
	return ""
}

func (a *Agent) ExecuteOperation(op types.UserOperation) string {
	// Orquestrar operação entre TPM e API
	// 1. Preparar TPM
	// 2. Fazer request para API
	// 3. Processar resposta
	// 4. Atualizar TPM se necessário
	return ""
}
