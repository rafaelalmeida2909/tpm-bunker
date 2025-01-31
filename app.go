package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"tpm-bunker/internal/agent"
	"tpm-bunker/internal/api"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	agent  *agent.Agent
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	a.ctx = ctx
	a.cancel = cancel

	initCtx, initCancel := context.WithTimeout(ctx, 30*time.Second)
	defer initCancel()

	tpmMgr := tpm.NewManager(initCtx)
	client := api.NewAPIClient(initCtx)
	a.agent = agent.NewAgent(ctx, tpmMgr, client)
}

// GetTPMStatus - chamado pelo frontend
func (a *App) GetTPMStatus() (*types.TPMStatus, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	return a.agent.GetTPMStatus(ctx)
}

// InitializeDevice - chamado pelo frontend
func (a *App) InitializeDevice() (*types.DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Minute)
	defer cancel()
	return a.agent.InitializeDevice(ctx)
}

// IsDeviceInitialized - chamado pelo frontend
func (a *App) IsDeviceInitialized() bool {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	return a.agent.IsDeviceInitialized(ctx)
}

// GetDeviceInfo - chamado pelo frontend
func (a *App) GetDeviceInfo() (*types.DeviceInfo, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	return a.agent.GetDeviceInfo(ctx)
}

// CheckTPMPresence - chamado pelo frontend
func (a *App) CheckTPMPresence() bool {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	return a.agent.CheckTPMPresence(ctx)
}

// CheckConnection - chamado pelo frontend
func (a *App) CheckConnection() bool {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Second)
	defer cancel()
	return a.agent.CheckConnection(ctx)
}

// AuthLogin - chamado pelo frontend
func (a *App) AuthLogin() bool {
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	return a.agent.AuthLogin(ctx)
}

// GetOperations - chamado pelo frontend
func (a *App) GetOperations() ([]byte, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()

	// Adicione logs
	response, err := a.agent.GetOperations(ctx)
	if err != nil {
		log.Printf("Erro em GetOperations: %v", err)
		return nil, err
	}

	return response, nil
}

// EncryptFile - chamado pelo frontend
func (a *App) EncryptFile(filePath string) error {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Minute)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in EncryptFile: %v", r)
		}
	}()

	if a.agent == nil {
		return fmt.Errorf("agent não inicializado")
	}

	initialized := a.agent.IsDeviceInitialized(ctx)
	if !initialized {
		return fmt.Errorf("device não inicializado. Aguarde a inicialização")
	}

	done := make(chan error, 1)
	go func() {
		defer close(done)
		encryptCtx, encryptCancel := context.WithTimeout(ctx, 9*time.Minute)
		defer encryptCancel()

		_, err := a.agent.Encrypt(encryptCtx, filePath)
		if err != nil {
			done <- fmt.Errorf("erro ao encriptar: %w", err)
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// DecryptFile - chamado pelo frontend
func (a *App) DecryptFile(operationID string) error {
	ctx, cancel := context.WithTimeout(a.ctx, 10*time.Minute)
	defer cancel()

	// Recuperação de pânico
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in DecryptFile: %v", r)
		}
	}()

	// Verifica se o agent está inicializado
	if a.agent == nil {
		return fmt.Errorf("agent não inicializado")
	}

	// Verifica se o dispositivo está inicializado
	initialized := a.agent.IsDeviceInitialized(ctx)
	if !initialized {
		return fmt.Errorf("device não inicializado. Aguarde a inicialização")
	}

	// Canal para resultado da operação assíncrona
	done := make(chan error, 1)
	go func() {
		defer close(done)

		// Contexto específico para decriptação
		decryptCtx, decryptCancel := context.WithTimeout(ctx, 9*time.Minute)
		defer decryptCancel()

		// Chama a função de decriptação do agent
		filePath, err := a.agent.Decrypt(decryptCtx, operationID)
		if err != nil {
			done <- fmt.Errorf("erro ao decriptar: %w", err)
			return
		}

		// Notifica o frontend sobre o sucesso e o caminho do arquivo
		runtime.EventsEmit(a.ctx, "decryption_complete", map[string]string{
			"status": "success",
			"path":   filePath,
		})

		done <- nil
	}()

	// Aguarda conclusão ou timeout
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SelectFile - chamado pelo frontend
func (a *App) SelectFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("contexto da aplicação não inicializado")
	}

	ctx, cancel := context.WithTimeout(a.ctx, 5*time.Minute)
	defer cancel()

	options := runtime.OpenDialogOptions{
		Title: "Selecione um arquivo para criptografar",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Todos os arquivos",
				Pattern:     "*.*",
			},
		},
	}

	return runtime.OpenFileDialog(ctx, options)
}

func (a *App) shutdown(ctx context.Context) {
	if a.cancel != nil {
		a.cancel()
	}
}
