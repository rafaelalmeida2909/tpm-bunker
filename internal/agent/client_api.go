package agent

import (
	"net/http"
)

type ClientAPI struct {
	tpmService *TPMService
	router     *http.ServeMux
}

func NewClientAPI(tpmService *TPMService) *ClientAPI {
	api := &ClientAPI{
		tpmService: tpmService,
		router:     http.NewServeMux(),
	}
	api.setupRoutes()
	return api
}

func (api *ClientAPI) setupRoutes() {
	// Rota para geração de chaves TPM
	api.router.HandleFunc("/generate-keys", api.handleGenerateKeys)
	// Rota para upload de arquivo local
	api.router.HandleFunc("/upload", api.handleFileUpload)
	// Rota para solicitação de descriptografia
	api.router.HandleFunc("/decrypt", api.handleDecryptRequest)
}
