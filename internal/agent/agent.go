package agent

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"tpm-bunker/internal/api"
	"tpm-bunker/internal/tpm"
	"tpm-bunker/internal/types"
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
		log.Printf("Erro ao verificar registro: %v", err)
		return nil, fmt.Errorf("falha ao verificar registro: %v", err)
	}

	// Registro com timeout específico
	if !isRegistered {
		fmt.Printf("Registrando novo dispositivo...")
		registerCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		if err := a.client.RegisterDevice(registerCtx, deviceInfo); err != nil {
			log.Printf("Falha ao registrar: %v", err)
			return nil, fmt.Errorf("falha ao registrar: %v", err)
		}
		fmt.Printf("Dispositivo registrado com sucesso")
	}

	return deviceInfo, nil
}

// AuthLogin tentar logar na API
func (a *Agent) AuthLogin(ctx context.Context) bool {
	deviceInfo, err := a.GetDeviceInfo(ctx)
	if err != nil {
		log.Printf("Falha ao obter informações: %v", err)
		return false
	}

	// Login com timeout específico
	loginCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	fmt.Printf("Realizando login na API...")
	if err := a.client.Login(loginCtx, deviceInfo.UUID, a.tpmMgr.EK); err != nil {
		log.Printf("Falha no login: %v", err)
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
			fmt.Printf("TPM presence check successful")
		} else {
			fmt.Printf("TPM not found or not accessible")
		}
		return hasTPM
	}
}

// CheckConnection verifica conexão com API
func (a *Agent) CheckConnection(ctx context.Context) bool {
	hasConnection := a.client.CheckConnection(ctx)
	if hasConnection {
		fmt.Printf("API connection successful")
	} else {
		fmt.Printf("API connection failed")
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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		pubKey, err := a.tpmMgr.Client.RetrieveRSASignKey(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve RSA key: %w", err)
		}

		result, err := EncryptFile(ctx, filePath, pubKey, a.tpmMgr)
		if err != nil {
			return nil, fmt.Errorf("encryption error: %w", err)
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

		apiCtx, apiCancel := context.WithTimeout(ctx, 2*time.Minute)
		defer apiCancel()

		return a.client.EncryptRequest(apiCtx, http.MethodPost, "operations/store_data/", header, payload)
	}
}

// Decrypt recupera e descriptografa um arquivo usando um operation_id
func (a *Agent) Decrypt(ctx context.Context, operationID string) (string, error) {
	// Timeout específico para decriptação
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Verifica cancelamento
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Headers para a requisição
		header := map[string]string{
			"X-Device-UUID": a.tpmMgr.DeviceUUID,
		}

		// Contexto específico para a requisição API
		apiCtx, apiCancel := context.WithTimeout(ctx, 2*time.Minute)
		defer apiCancel()

		// Faz a requisição para recuperar os dados encriptados
		log.Printf("Recuperando dados da operação: %s", operationID)
		response, err := a.client.DecryptRequest(apiCtx, http.MethodGet, "operations/retrieve_data/", header, operationID)
		if err != nil {
			return "", fmt.Errorf("erro ao recuperar dados da API: %w", err)
		}

		// Contexto específico para decriptação
		decryptCtx, decryptCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer decryptCancel()

		// Descriptografa os dados
		log.Printf("Iniciando processo de decriptação")
		result, err := DecryptFile(decryptCtx, response, a.tpmMgr)
		if err != nil {
			return "", fmt.Errorf("erro na decriptação: %w", err)
		}

		if !result.Verified {
			return "", fmt.Errorf("falha na verificação da assinatura digital")
		}

		// Obter caminho da pasta Downloads
		downloadPath, err := getDownloadsPath()
		if err != nil {
			return "", fmt.Errorf("erro ao obter pasta de downloads: %w", err)
		}

		// Se o nome do arquivo não estiver disponível, usar um nome padrão
		fileName := response.FileName
		if fileName == "" {
			fileName = fmt.Sprintf("decrypted_file_%s_%s",
				operationID,
				time.Now().Format("20060102_150405"))
		}

		// Caminho completo do arquivo
		filePath := filepath.Join(downloadPath, fileName)

		// Garantir que não sobrescreva arquivo existente
		filePath = ensureUniqueFilePath(filePath)

		// Salvar arquivo
		err = os.WriteFile(filePath, result.DecryptedData, 0600)
		if err != nil {
			return "", fmt.Errorf("erro ao salvar arquivo: %w", err)
		}

		log.Printf("Arquivo salvo com sucesso em: %s", filePath)
		return filePath, nil
	}
}

// getDownloadsPath retorna o caminho da pasta Downloads do usuário
func getDownloadsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("erro ao obter pasta home: %w", err)
	}

	var downloadPath string
	if runtime.GOOS == "windows" {
		downloadPath = filepath.Join(homeDir, "Downloads")
	} else {
		// Linux e outros sistemas Unix-like
		downloadPath = filepath.Join(homeDir, "Downloads")

		// Se a pasta Downloads não existir, tentar XDG_DOWNLOAD_DIR
		if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
			// Tenta ler o arquivo user-dirs.dirs
			xdgConfig := filepath.Join(homeDir, ".config", "user-dirs.dirs")
			if data, err := os.ReadFile(xdgConfig); err == nil {
				for _, line := range strings.Split(string(data), "\n") {
					if strings.HasPrefix(line, "XDG_DOWNLOAD_DIR=") {
						dir := strings.Trim(strings.Split(line, "=")[1], "\"")
						dir = strings.ReplaceAll(dir, "$HOME", homeDir)
						if _, err := os.Stat(dir); err == nil {
							downloadPath = dir
							break
						}
					}
				}
			}
		}
	}

	// Verifica se a pasta existe
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		return "", fmt.Errorf("pasta de downloads não encontrada: %s", downloadPath)
	}

	return downloadPath, nil
}

// ensureUniqueFilePath garante que o caminho do arquivo não existe,
// adicionando um número se necessário
func ensureUniqueFilePath(filePath string) string {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return filePath
	}

	ext := filepath.Ext(filePath)
	baseFilePath := filePath[:len(filePath)-len(ext)]
	counter := 1

	for {
		newPath := fmt.Sprintf("%s_%d%s", baseFilePath, counter, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}
