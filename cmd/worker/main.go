package main

import (
	"context"
	"log"
	"time"

	"mms_api/cmd/worker/bootstrap"
	"mms_api/config"
)

func main() {
	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Criar contexto com cancelamento
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicializar worker
	worker, err := bootstrap.NewWorker(cfg)
	if err != nil {
		log.Fatalf("Erro ao inicializar worker: %v", err)
	}
	defer worker.Close()

	// Configurar intervalo de execução (por exemplo, uma vez por dia às 00:00)
	interval := 24 * time.Hour

	// Executar worker com agendamento
	if err := worker.RunScheduled(ctx, interval); err != nil {
		log.Fatalf("Erro ao configurar execução programada do worker: %v", err)
	}
}
