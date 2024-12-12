package server

import (
	"github.com/rafaelalmeida2909/tpm-bunker/pkg/config"

	"github.com/gin-gonic/gin"
)

type Server struct {
	router         *gin.Engine
	encryptService *EncryptionService
}

func Start(cfg *config.Config) error {
	server := &Server{
		router:         gin.Default(),
		encryptService: NewEncryptionService(),
	}

	server.setupRoutes()
	return server.router.Run(cfg.ServerAddress)
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		api.POST("/upload", s.handleFileUpload)
		api.POST("/attest", s.handleAttestation)
		api.GET("/file/:id", s.handleFileDownload)
	}
}
