package agent

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"tpm-bunker/internal/api"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Agent struct {
	ctx    context.Context
	tpmMgr *tpm.Manager
	client *api.APIClient
}

func NewAgent(ctx context.Context, tpmMgr *tpm.Manager, client *api.APIClient) *Agent {
	return &Agent{
		ctx:    ctx,
		tpmMgr: tpmMgr,
		client: client,
	}
}

// GetTPMStatus retorna o status atual do TPM
func (a *Agent) GetTPMStatus() (*types.TPMStatus, error) {
	return a.tpmMgr.GetStatus()
}

// InitializeDevice inicializa o dispositivo com TPM pela primeira vez
func (a *Agent) InitializeDevice() (*types.DeviceInfo, error) {
	status, err := a.tpmMgr.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar status: %v", err)
	}

	if !status.Available {
		return nil, fmt.Errorf("TPM não está disponível neste dispositivo")
	}

	// Inicializa o dispositivo
	err = a.tpmMgr.InitializeDevice()
	if err != nil {
		return nil, fmt.Errorf("falha ao inicializar dispositivo: %v", err)
	}

	// Obtém as informações do dispositivo
	deviceInfo := &types.DeviceInfo{
		UUID:      a.tpmMgr.GetDeviceUUID(),
		PublicKey: a.tpmMgr.GetPublicKey(),
		AIK:       a.tpmMgr.AIK,
		EK:        a.tpmMgr.EK,
	}

	// Verifica a conexão com a API
	if !a.client.CheckConnection() {
		return nil, fmt.Errorf("API não está acessível")
	}

	// Verifica se o dispositivo já está registrado
	isRegistered, err := a.client.IsDeviceRegistered(deviceInfo.UUID)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Erro ao verificar registro do dispositivo: %v", err))
		return nil, fmt.Errorf("falha ao verificar registro do dispositivo: %v", err)
	}

	// Só registra se ainda não estiver registrado
	if !isRegistered {
		runtime.LogInfo(a.ctx, "Registrando novo dispositivo...")
		err = a.client.RegisterDevice(deviceInfo)
		if err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Falha ao registrar dispositivo na API: %v", err))
			return nil, fmt.Errorf("falha ao registrar dispositivo na API: %v", err)
		}
		runtime.LogInfo(a.ctx, "Dispositivo registrado com sucesso")
	} else {
		runtime.LogInfo(a.ctx, "Dispositivo já registrado anteriormente")
	}

	return deviceInfo, nil
}

// AuthLogin tentar logar na API
func (a *Agent) AuthLogin() bool {
	deviceInfo, err := a.GetDeviceInfo()
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Falha ao obter informações do dispositivo: %v", err))
		return false
	}

	runtime.LogInfo(a.ctx, "Realizando login na API...")
	err = a.client.Login(deviceInfo.UUID, a.tpmMgr.EK)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Falha ao realizar login na API: %v", err))
		return false
	}
	runtime.LogInfo(a.ctx, "Login realizado com sucesso")

	return true
}

// IsDeviceInitialized verifica se o dispositivo já foi inicializado
func (a *Agent) IsDeviceInitialized() bool {
	status, _ := a.tpmMgr.GetStatus()
	return status.Initialized
}

// GetDeviceInfo retorna as informações do dispositivo
func (a *Agent) GetDeviceInfo() (*types.DeviceInfo, error) {
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

// CheckTPMPresence verifica se o TPM está presente e acessível
func (a *Agent) CheckTPMPresence() bool {
	hasTPM := tpm.CheckTPMPresence()
	if hasTPM {
		runtime.LogInfo(a.ctx, "TPM presence check successful")
	} else {
		runtime.LogInfo(a.ctx, "TPM not found or not accessible")
	}
	return hasTPM
}

// CheckTPMPresence verifica se o TPM está presente e acessível
func (a *Agent) CheckConnection() bool {
	hasConnection := a.client.CheckConnection()
	if hasConnection {
		runtime.LogInfo(a.ctx, "API connection successful")
	} else {
		runtime.LogInfo(a.ctx, "API connection failed")
	}
	return hasConnection
}

// Encrypt encripta um arquivo e o envia para a API
func (a *Agent) Encrypt(filePath string) ([]byte, error) {
	// Obtém o keyPair do TPM Manager
	keyPair := a.tpmMgr.Client.KeyPair
	if keyPair == nil {
		return nil, fmt.Errorf("chave RSA não inicializada")
	}

	// Encripta o arquivo
	result, err := EncryptFile(filePath, keyPair)
	if err != nil {
		return nil, fmt.Errorf("erro na encriptação: %w", err)
	}

	// Lê o arquivo encriptado
	encryptedData, err := os.ReadFile(result.EncryptedFilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo encriptado: %w", err)
	}

	// Prepara o payload para a API
	payload := &api.EncryptionRequest{
		EncryptedData:    base64.StdEncoding.EncodeToString(encryptedData),
		EncryptedKey:     result.EncryptedSymmetricKey,
		DigitalSignature: result.DigitalSignature,
		Metadata:         result.Metadata,
	}

	header := map[string]string{
		"X-Device-UUID": a.tpmMgr.DeviceUUID,
	}

	// Envia para a API
	response, err := a.client.SendRequest(http.MethodPost, "operations/store_data/", header, payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar arquivo para API: %w", err)
	}

	return response, nil
}

// Decrypt Recupera um arquivo através da API com um operation_id e o descriptografa
func (a *Agent) Decrypt() []byte {
	// implementar
	return nil
}
