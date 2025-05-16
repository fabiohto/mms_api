package bootstrap

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"mms_api/config"
	"mms_api/internal/adapter/out/mercadobitcoin"
	"mms_api/internal/adapter/out/persistence/postgres"
	"mms_api/internal/application/port/out"
	"mms_api/internal/application/service"
	dbconfig "mms_api/pkg/db/postgres"
	"mms_api/pkg/logger"
	"mms_api/pkg/monitoring"

	"github.com/go-co-op/gocron"
)

type Worker struct {
	mmsService    service.MMSService
	mmsRepo       out.MMSRepository
	alertMonitor  monitoring.AlertMonitor
	logger        logger.Logger
	db            *sql.DB
	retryInterval time.Duration // Intervalo de retry configurável
}

func NewWorker(cfg *config.Config) (*Worker, error) {
	// Inicializar logger
	l := logger.NewLogger("[WORKER] ")

	// Conectar ao banco de dados
	db, err := dbconfig.NewConnection(cfg.Database)
	if err != nil {
		l.Error("Erro ao conectar ao banco de dados", err)
		return nil, err
	}

	// Inicializar repositório
	mmsRepo := postgres.NewMMSRepository(db, l)

	// Inicializar HTTP client para a API de candles
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Inicializar API de candles
	candleAPI := mercadobitcoin.NewCandleAPI(cfg.MercadoBitcoinBaseURL, httpClient, l)

	// Inicializar serviço
	mmsService := service.NewMMSService(mmsRepo, candleAPI, l)

	// Inicializar monitor de alertas
	alertMonitor := monitoring.NewAlertMonitor(cfg.AlertConfig, l)

	return &Worker{
		mmsService:    mmsService,
		mmsRepo:       mmsRepo,
		alertMonitor:  alertMonitor,
		logger:        l,
		db:            db,
		retryInterval: 1 * time.Hour, // Valor padrão
	}, nil
}

// NewWorkerWithDeps cria um novo worker com dependências injetadas (usado para testes)
func NewWorkerWithDeps(mmsService service.MMSService, mmsRepo out.MMSRepository, alertMonitor monitoring.AlertMonitor, l logger.Logger) *Worker {
	return &Worker{
		mmsService:    mmsService,
		mmsRepo:       mmsRepo,
		alertMonitor:  alertMonitor,
		logger:        l,
		retryInterval: 100 * time.Millisecond, // Valor menor para testes
	}
}

// SetRetryInterval configura o intervalo de retry
func (w *Worker) SetRetryInterval(interval time.Duration) {
	w.retryInterval = interval
}

// Close fecha as conexões do worker
func (w *Worker) Close() error {
	return w.db.Close()
}

// Run executa o processo do worker
func (w *Worker) Run() error {
	// Contexto com cancelamento para permitir tentativas repetidas
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configurações de retry
	maxRetries := 5

	// Pares a serem processados
	pairs := []string{"BRLBTC", "BRLETH"}

	for _, pair := range pairs {
		// Obter última data processada
		lastTimestamp, err := w.mmsRepo.GetLastTimestamp(ctx, pair)
		if err != nil {
			w.logger.Error("Erro ao obter último timestamp", err, "pair", pair)
			continue
		}

		// Se não houver dados, começar do início (último ano)
		var from time.Time
		if lastTimestamp.IsZero() {
			from = time.Now().AddDate(-1, 0, 0)
		} else {
			// Caso contrário, começar do dia seguinte ao último processado
			from = lastTimestamp.AddDate(0, 0, 1)
		}

		// Se a data for posterior a hoje, não há nada a processar
		today := time.Now()
		if from.After(today) {
			w.logger.Info("Dados já atualizados", "pair", pair)
			continue
		}

		// Calcular até ontem
		to := today.AddDate(0, 0, -1)

		// Processar com retry em caso de falha
		success := false
		for attempt := 0; attempt < maxRetries && !success; attempt++ {
			if attempt > 0 {
				w.logger.Info("Tentando novamente", "attempt", attempt+1, "pair", pair)
				time.Sleep(w.retryInterval)
			}

			err := w.mmsService.CalculateAndSaveMMSForRange(ctx, pair, from, to)
			if err == nil {
				success = true
				w.logger.Info("Atualização concluída com sucesso", "pair", pair)
			} else {
				w.logger.Error("Erro na atualização", err, "pair", pair, "attempt", attempt+1)
			}
		}

		if !success {
			w.logger.Error("Falha após todas as tentativas", "pair", pair)
			w.alertMonitor.SendAlert("falha_atualizacao", "Falha na atualização diária de "+pair)
		}

		// Verificar completude dos dados
		isComplete, missingDates, err := w.mmsService.CheckDataCompleteness(ctx, pair)
		if err != nil {
			w.logger.Error("Erro ao verificar completude dos dados", err, "pair", pair)
			continue
		}

		if !isComplete {
			w.logger.Info("Dados incompletos detectados", "pair", pair, "missingDates", missingDates)
			w.alertMonitor.SendAlert("dados_incompletos", "Dados incompletos para "+pair)
		}
	}

	return nil
}

// RunScheduled executa o worker em um intervalo programado
func (w *Worker) RunScheduled(ctx context.Context, interval time.Duration) error {
	scheduler := gocron.NewScheduler(time.UTC)

	// Configurar job para executar no intervalo especificado
	_, err := scheduler.Every(interval).Do(func() {
		if err := w.Run(); err != nil {
			w.logger.Error("Erro na execução programada do worker", err)
			w.alertMonitor.SendAlert("erro_execucao", "Erro na execução programada do worker")
		}
	})

	if err != nil {
		return err
	}

	// Iniciar scheduler em uma goroutine
	scheduler.StartAsync()

	// Aguardar sinal de cancelamento do contexto
	<-ctx.Done()
	scheduler.Stop()

	return nil
}
