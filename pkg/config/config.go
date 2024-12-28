package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Configurações do servidor
	ServerAddress string
	ServerPort    string

	// Configurações do TPM
	TPMDevicePath string

	// Configurações do banco de dados
	DatabaseName string
	DatabasePort string
	DatabaseHost string
	DatabaseUser string
	DatabasePass string

	// Configurações de segurança
	JWTSecret    string
	KeyDirectory string
}

func Load() (*Config, error) {
	// Carrega variáveis de ambiente do arquivo .env
	if err := godotenv.Load(); err != nil {
		// Continua mesmo se o arquivo .env não existir
		log.Printf("No .env file found")
	}

	return &Config{
		ServerAddress: getEnvDefault("SERVER_ADDRESS", "localhost"),
		ServerPort:    getEnvDefault("SERVER_PORT", "8080"),
		TPMDevicePath: getEnvDefault("TPM_DEVICE_PATH", "/dev/tpm0"),
		DatabaseName:  getEnvDefault("DATABASE_NAME", "tpmbunker"),
		DatabasePort:  getEnvDefault("DATABASE_PORT", "5432"),
		DatabaseHost:  getEnvDefault("DATABASE_HOST", "localhost"),
		DatabaseUser:  getEnvDefault("DATABASE_USER", "postgres"),
		DatabasePass:  getEnvDefault("DATABASE_PASS", "postgres"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		KeyDirectory:  getEnvDefault("KEY_DIRECTORY", "./keys"),
	}, nil
}

func getEnvDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
