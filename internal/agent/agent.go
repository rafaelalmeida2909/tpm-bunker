package agent

import (
	"fmt"
	"io"

	"github.com/google/go-tpm/tpm2"
	"github.com/rafaelalmeida2909/tpm-bunker/pkg/config"
)

type Agent struct {
	tpm        io.ReadWriteCloser
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
		tpm:        rwc,
		tpmService: NewTPMService(rwc),
		clientAPI:  NewClientAPI(nil),
	}

	return agent.Run()
}
