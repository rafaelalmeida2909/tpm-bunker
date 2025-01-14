package types

import "crypto/rsa"

// DeviceInfo contém informações sobre o dispositivo
type DeviceInfo struct {
	UUID      string
	PublicKey *rsa.PublicKey
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
