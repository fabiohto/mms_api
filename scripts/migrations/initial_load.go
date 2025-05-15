package main

import (
	"context"
	"log"
	"time"

	"mms-service/config"
	"mms-service/internal/adapter/out/mercadobitcoin"
	"mms-service/internal/adapter/out/persistence/postgres"
	"mms-service/internal/application/service"
	"mms-service/pkg/logger"
)

func main() {
	// Carregar configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Inicializar logger
	l := logger.NewLogger()

	// Conectar ao banco de dados
	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	// Inicializar repositório
	mmsRepo := postgres.NewMMSRepository(db, l)

	// Inicializar API de candles
	candleAPI := mercadobitcoin.NewCandleAPI(cfg.MercadoBitcoinBaseURL, nil, l)

	// Inicializar serviço
	mmsService := service.NewMMSService(mmsRepo, candleAPI, l)

	// Calcular período para carga inicial (últimos 365 dias)
	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	// Executar carga inicial para cada par
	pairs := []string{"BRLBTC", "BRLETH"}
	ctx := context.Background()

	for _, pair := range pairs {
		log.Printf("Iniciando carga para %s de %s até %s", pair, from.Format("2006-01-02"), to.Format("2006-01-02"))

		if err := mmsService.CalculateAndSaveMMSForRange(ctx, pair, from, to); err != nil {
			log.Fatalf("Erro ao calcular e salvar MMS para %s: %v", pair, err)
		}

		log.Printf("Carga para %s concluída com sucesso", pair)
	}

	log.Println("Carga inicial concluída com sucesso!")
}
