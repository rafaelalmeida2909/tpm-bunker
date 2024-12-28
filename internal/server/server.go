package server

/*
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
*/

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rafaelalmeida2909/tpm-bunker/pkg/config"
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

// Add these handler functions:

func (s *Server) handleFileUpload(c *gin.Context) {
	// Get the file from the request
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// TODO: Implement file encryption and storage logic
	// You might want to call methods from your EncryptionService here

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": file.Filename,
	})
}

func (s *Server) handleAttestation(c *gin.Context) {
	// TODO: Implement TPM attestation logic
	// This would typically involve verifying TPM quotes or measurements

	c.JSON(http.StatusOK, gin.H{
		"message": "Attestation successful",
		// Add any relevant attestation data
	})
}

func (s *Server) handleFileDownload(c *gin.Context) {
	fileID := c.Param("id")

	// TODO: Implement file retrieval and decryption logic
	// You might want to call methods from your EncryptionService here

	c.JSON(http.StatusOK, gin.H{
		"message": "File download endpoint",
		"file_id": fileID,
	})
}
