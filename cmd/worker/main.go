package main

import (
	"log"

	"mms_api/cmd/worker/bootstrap"
	"mms_api/config"
)

func main() {
	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Inicializar worker
	worker, err := bootstrap.NewWorker(cfg)
	if err != nil {
		log.Fatalf("Erro ao inicializar worker: %v", err)
	}
	defer worker.Close()

	// Executar worker
	if err := worker.Run(); err != nil {
		log.Fatalf("Erro ao executar worker: %v", err)
	}
}
