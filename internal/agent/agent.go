package agent

import (
	"fmt"

	"github.com/google/go-tpm/tpm2"
	"github.com/rafaelalmeida2909/tpm-bunker/pkg/config"
)

type Agent struct {
	tpmHandle  *tpm2.TPMContext
	clientAPI  *ClientAPI
	tpmService *TPMService
}

func Start(cfg *config.Config) error {
	// Initialize TPM connection
	tpmHandle, err := tpm2.OpenTPM()
	if err != nil {
		return fmt.Errorf("failed to open TPM: %v", err)
	}
	defer tpmHandle.Close()

	agent := &Agent{
		tpmHandle:  tpmHandle,
		clientAPI:  NewClientAPI(),
		tpmService: NewTPMService(tpmHandle),
	}

	return agent.Run()
}
