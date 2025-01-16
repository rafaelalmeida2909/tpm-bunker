package types

// DeviceInfo contém informações sobre o dispositivo
type DeviceInfo struct {
	UUID      string
	PublicKey string
	EK        []byte
	AIK       []byte
}

// APIResponse representa uma resposta da API
type APIResponse struct {
	Success bool
	Message string
}

// UserOperation representa uma operação solicitada pelo usuário
type UserOperation struct {
	Type string
	Data interface{}
}

// TPMStatus representa o status do TPM
type TPMStatus struct {
	Available   bool `json:"available"`
	Initialized bool `json:"initialized"`
}
