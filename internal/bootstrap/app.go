package bootstrap

import (
	"net/http"
	"time"

	"mms_api/config"
	httpAdapter "mms_api/internal/adapter/in/http"
	"mms_api/internal/adapter/in/http/handlers"
	"mms_api/internal/adapter/in/http/server"
	"mms_api/internal/adapter/out/mercadobitcoin"
	"mms_api/internal/adapter/out/persistence/postgres"
	"mms_api/internal/application/service"
	pgconfig "mms_api/pkg/db/postgres"
	"mms_api/pkg/logger"
)

// App encapsula todas as dependências da aplicação
type App struct {
	server *server.Server
	logger logger.Logger
}

// NewApp inicializa todas as dependências da aplicação
func NewApp(port string) *App {
	// Setup logger
	log := logger.NewLogger("[API] ")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro ao carregar configurações", err)
	}

	// Setup database connection
	db, err := pgconfig.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco de dados", err)
	}

	// Initialize repositories
	mmsRepo := postgres.NewMMSRepository(db, log)

	// Initialize HTTP client for external APIs
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Initialize external APIs
	candleAPI := mercadobitcoin.NewCandleAPI(cfg.MercadoBitcoinBaseURL, httpClient, log)

	// Setup service and handlers
	mmsService := service.NewMMSService(mmsRepo, candleAPI, log)
	mmsHandler := handlers.NewMMSHandler(mmsService, log)

	// Initialize router
	router := httpAdapter.NewRouter(mmsHandler)
	ginEngine := router.SetupRoutes()

	// Create server
	srv := server.NewServer(ginEngine, port, log)

	return &App{
		server: srv,
		logger: log,
	}
}

// Start inicia o servidor HTTP
func (app *App) Start() error {
	return app.server.Start()
}
