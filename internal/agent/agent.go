package agent

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"time"
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
func (a *Agent) GetTPMStatus(ctx context.Context) (*types.TPMStatus, error) {
	return a.tpmMgr.GetStatus(ctx)
}

// InitializeDevice inicializa o dispositivo com TPM pela primeira vez
func (a *Agent) InitializeDevice(ctx context.Context) (*types.DeviceInfo, error) {
	// Criamos um timeout específico para inicialização
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	status, err := a.tpmMgr.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar status: %v", err)
	}

	if !status.Available {
		return nil, fmt.Errorf("TPM não está disponível neste dispositivo")
	}

	// Inicializa o dispositivo
	if err := a.tpmMgr.InitializeDevice(ctx); err != nil {
		return nil, fmt.Errorf("falha ao inicializar dispositivo: %v", err)
	}

	// Obter UUID
	uuid, err := a.tpmMgr.GetDeviceUUID(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter UUID: %w", err)
	}

	// Obter chave pública
	pubKey, err := a.tpmMgr.GetPublicKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter chave pública: %w", err)
	}

	deviceInfo := &types.DeviceInfo{
		UUID:      uuid,
		PublicKey: pubKey,
		AIK:       a.tpmMgr.AIK,
		EK:        a.tpmMgr.EK,
	}

	// Verifica a conexão com a API
	if !a.client.CheckConnection(ctx) {
		return nil, fmt.Errorf("API não está acessível")
	}

	// Verifica se o dispositivo já está registrado
	isRegistered, err := a.client.IsDeviceRegistered(ctx, deviceInfo.UUID)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Erro ao verificar registro: %v", err))
		return nil, fmt.Errorf("falha ao verificar registro: %v", err)
	}

	// Registro com timeout específico
	if !isRegistered {
		runtime.LogInfo(a.ctx, "Registrando novo dispositivo...")
		registerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := a.client.RegisterDevice(registerCtx, deviceInfo); err != nil {
			runtime.LogError(a.ctx, fmt.Sprintf("Falha ao registrar: %v", err))
			return nil, fmt.Errorf("falha ao registrar: %v", err)
		}
		runtime.LogInfo(a.ctx, "Dispositivo registrado com sucesso")
	}

	return deviceInfo, nil
}

// AuthLogin tentar logar na API
func (a *Agent) AuthLogin(ctx context.Context) bool {
	deviceInfo, err := a.GetDeviceInfo(ctx)
	if err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Falha ao obter informações: %v", err))
		return false
	}

	// Login com timeout específico
	loginCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	runtime.LogInfo(a.ctx, "Realizando login na API...")
	if err := a.client.Login(loginCtx, deviceInfo.UUID, a.tpmMgr.EK); err != nil {
		runtime.LogError(a.ctx, fmt.Sprintf("Falha no login: %v", err))
		return false
	}

	return true
}

// IsDeviceInitialized verifica se o dispositivo já foi inicializado
func (a *Agent) IsDeviceInitialized(ctx context.Context) bool {
	status, _ := a.tpmMgr.GetStatus(ctx)
	return status.Initialized
}

// GetDeviceInfo retorna as informações do dispositivo
func (a *Agent) GetDeviceInfo(ctx context.Context) (*types.DeviceInfo, error) {
	if a.tpmMgr == nil {
		return nil, fmt.Errorf("TPM Manager não inicializado")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		uuid, _ := a.tpmMgr.GetDeviceUUID(ctx)
		pubKey, _ := a.tpmMgr.GetPublicKey(ctx)
		return &types.DeviceInfo{
			UUID:      uuid,
			PublicKey: pubKey,
		}, nil
	}
}

// CheckTPMPresence verifica se o TPM está presente e acessível
func (a *Agent) CheckTPMPresence(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		hasTPM := tpm.CheckTPMPresence(ctx)
		if hasTPM {
			runtime.LogInfo(a.ctx, "TPM presence check successful")
		} else {
			runtime.LogInfo(a.ctx, "TPM not found or not accessible")
		}
		return hasTPM
	}
}

// CheckConnection verifica conexão com API
func (a *Agent) CheckConnection(ctx context.Context) bool {
	hasConnection := a.client.CheckConnection(ctx)
	if hasConnection {
		runtime.LogInfo(a.ctx, "API connection successful")
	} else {
		runtime.LogInfo(a.ctx, "API connection failed")
	}
	return hasConnection
}

// GetOperations recupera as operações entre Agente e API
func (a *Agent) GetOperations(ctx context.Context) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Verifica cancelamento
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		header := map[string]string{
			"X-Device-UUID": a.tpmMgr.DeviceUUID,
		}

		// Envia para API com timeout específico
		apiCtx, apiCancel := context.WithTimeout(ctx, 2*time.Minute)
		defer apiCancel()
		log.Printf("Realizando get de operações")
		return a.client.SendRequest(apiCtx, http.MethodGet, "operations/", header, nil)
	}
}

// Encrypt encripta um arquivo e o envia para a API
func (a *Agent) Encrypt(ctx context.Context, filePath string) ([]byte, error) {
	// Timeout específico para encriptação
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Verifica cancelamento
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Verifica se a chave RSA foi inicializada
		if a.tpmMgr.Client.KeyPair == nil || a.tpmMgr.Client.KeyPair.PublicKey.N == nil {
			log.Printf("Chave RSA não inicializada ou inválida. Tentando recuperar...")
			pubKey, err := a.tpmMgr.Client.RetrieveRSAKey(ctx)
			if err != nil {
				return nil, fmt.Errorf("falha ao recuperar chave RSA: %w", err)
			}
			a.tpmMgr.Client.KeyPair = &rsa.PrivateKey{
				PublicKey: *pubKey,
			}
			log.Printf("Chave RSA recuperada com sucesso.")
		}

		// Continua com a encriptação
		keyPair := a.tpmMgr.Client.KeyPair
		log.Printf("Usando chave RSA: %+v", keyPair.PublicKey)

		result, err := EncryptFile(ctx, filePath, keyPair, a.tpmMgr)
		if err != nil {
			return nil, fmt.Errorf("erro na encriptação: %w", err)
		}

		payload := &api.EncryptionRequest{
			EncryptedData:    result.EncryptedData,
			EncryptedKey:     result.EncryptedSymmetricKey,
			DigitalSignature: result.DigitalSignature,
			HashOriginal:     result.HashOriginal,
			Metadata:         result.Metadata,
		}

		header := map[string]string{
			"X-Device-UUID": a.tpmMgr.DeviceUUID,
		}

		// Envia para API com timeout específico
		apiCtx, apiCancel := context.WithTimeout(ctx, 2*time.Minute)
		defer apiCancel()

		return a.client.EncryptRequest(apiCtx, http.MethodPost, "operations/store_data/", header, payload)
	}
}

// Decrypt Recupera um arquivo através da API com um operation_id e o descriptografa
func (a *Agent) Decrypt() []byte {
	// implementar
	return nil
}
