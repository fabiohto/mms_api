package worker_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mms_api/cmd/worker/bootstrap"
	"mms_api/config"
	"mms_api/pkg/db/postgres"
	"mms_api/pkg/monitoring"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerIntegration(t *testing.T) {
	// Generate test dates
	now := time.Now()
	yearAgo := now.AddDate(-1, 0, 0)

	// Configurar servidor mock para a API do Mercado Bitcoin
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Gerar 200 dias de dados
		var candles []map[string]interface{}
		for i := 0; i < 200; i++ {
			date := yearAgo.AddDate(0, 0, i)
			candles = append(candles, map[string]interface{}{
				"timestamp": date.Format(time.RFC3339),
				"open":      150000.0,
				"high":      155000.0,
				"low":       149000.0,
				"close":     152000.0,
				"volume":    10.5,
				"quantity":  5,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"candles": candles,
		})
	}))
	defer mockServer.Close()

	// Configurar banco de dados de teste
	dbConfig := postgres.Config{
		Host:     "test-db",
		Port:     "5432",
		User:     "test_user",
		Password: "test_password",
		DBName:   "test_db",
	}

	// Criar banco de dados de teste
	db, err := postgres.NewConnectionWithTimeout(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Limpar dados existentes
	_, err = db.Exec("TRUNCATE TABLE mms")
	require.NoError(t, err)

	// Configurar worker com o servidor mock
	cfg := &config.Config{
		Database:              dbConfig,
		MercadoBitcoinBaseURL: mockServer.URL,
		AlertConfig: monitoring.AlertConfig{
			Enabled: false,
			Email: monitoring.EmailConfig{
				Enabled:      false,
				SMTPHost:     "localhost",
				SMTPPort:     1025,
				SMTPUsername: "",
				SMTPPassword: "",
				FromEmail:    "test@example.com",
				ToEmails:     []string{"alert@example.com"},
			},
		},
	}

	// Criar e executar worker
	worker, err := bootstrap.NewWorker(cfg)
	require.NoError(t, err)
	defer worker.Close()

	// Executar worker
	err = worker.Run()
	require.NoError(t, err)

	// Verificar resultados no banco de dados
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM mms").Scan(&count)
	require.NoError(t, err)
	assert.Greater(t, count, 0, "Deve haver registros de MMS no banco")

	// Verificar valores específicos
	rows, err := db.Query(`
		SELECT pair, timestamp, mms20, mms50, mms200 
		FROM mms 
		WHERE pair = 'BRLBTC' 
		AND timestamp BETWEEN $1 AND $2
		ORDER BY timestamp DESC 
		LIMIT 1
	`, yearAgo, now)
	require.NoError(t, err)
	defer rows.Close()

	if rows.Next() {
		var (
			pair      string
			timestamp time.Time
			mms20     float64
			mms50     float64
			mms200    float64
		)
		err := rows.Scan(&pair, &timestamp, &mms20, &mms50, &mms200)
		require.NoError(t, err)

		assert.Equal(t, "BRLBTC", pair)
		assert.NotZero(t, timestamp)
		assert.NotZero(t, mms20)
		assert.NotZero(t, mms50)
		assert.NotZero(t, mms200)
	} else {
		t.Error("No MMS records found in the date range")
	}
}
