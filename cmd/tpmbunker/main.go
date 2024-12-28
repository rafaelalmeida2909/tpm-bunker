package main

import (
	"log"

	"github.com/rafaelalmeida2909/tpm-bunker/pkg/config"
)

/*
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
*/
func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize and start server
	srv := server.New(cfg) // assuming you have a New function in server package
	srv.Start()            // or whatever method you use to start the server
}
