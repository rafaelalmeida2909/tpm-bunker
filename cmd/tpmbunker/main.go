package main

import (
	"config"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rafaelalmeida2909/tpm-bunker/internal/agent"
	"github.com/rafaelalmeida2909/tpm-bunker/internal/server"
)

func main() {
	// Contexto com cancelamento para shutdown gracioso
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Carrega configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Inicia o agente em uma goroutine separada
	agentErrCh := make(chan error, 1)
	go func() {
		if err := agent.Start(ctx, cfg); err != nil {
			agentErrCh <- err
		}
	}()

	// Inicia o servidor em uma goroutine separada
	serverErrCh := make(chan error, 1)
	go func() {
		if err := server.Start(ctx, cfg); err != nil {
			serverErrCh <- err
		}
	}()

	// Gerencia o shutdown gracioso
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-agentErrCh:
		log.Printf("Agent error: %v", err)
	case err := <-serverErrCh:
		log.Printf("Server error: %v", err)
	case sig := <-sigCh:
		log.Printf("Received signal: %v", sig)
	}

	// Inicia o processo de shutdown
	cancel()
	log.Println("Shutting down gracefully...")
}
