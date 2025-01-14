package agent

import (
	"tpm-bunker/internal/types"
)

type Orchestrator struct {
	agent     *Agent
	apiClient *APIClient
}

func NewOrchestrator(agent *Agent) *Orchestrator {
	return &Orchestrator{
		agent:     agent,
		apiClient: NewAPIClient(),
	}
}

func (o *Orchestrator) HandleOperation(op types.UserOperation) string {
	// Coordena operações entre TPM e API
	return ""
}
