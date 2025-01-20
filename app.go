package main

import (
	"context"
	"fmt"
	"tpm-bunker/internal/agent"
	"tpm-bunker/internal/api"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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

// AuthLogin tenta realizar login na API
func (a *App) AuthLogin() bool {
	return a.agent.AuthLogin()
}

// EncryptFile encripta o arquivo, envia para a API
func (a *App) EncryptFile(filePath string) error {
	if a.agent == nil {
		return fmt.Errorf("agent não inicializado")
	}

	// Verifica se o device está inicializado
	initialized := a.agent.IsDeviceInitialized()
	if !initialized {
		return fmt.Errorf("device não inicializado. Aguarde a inicialização ser concluída")
	}

	// Tenta encriptar
	_, err := a.agent.Encrypt(filePath)
	if err != nil {
		return fmt.Errorf("erro ao encriptar: %w", err)
	}

	return nil
}

func (a *App) SelectFile() (string, error) {
	// Cria as opções do dialog
	options := runtime.OpenDialogOptions{
		Title: "Selecione um arquivo para criptografar",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Todos os arquivos",
				Pattern:     "*.*",
			},
		},
	}

	// Abre o dialog e retorna o caminho do arquivo selecionado
	return runtime.OpenFileDialog(a.ctx, options)
}

// shutdown é chamado quando o app é fechado
func (a *App) shutdown(ctx context.Context) {
	// Cleanup é feito automaticamente pela linguagem
}
