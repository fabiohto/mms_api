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

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkerIntegration(t *testing.T) {
	// Configurar servidor mock para a API do Mercado Bitcoin
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simular resposta da API de candles com dados incompletos para forçar alerta
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candles": [
				{
					"timestamp": "2025-05-14T00:00:00Z",
					"open": 150000.0,
					"high": 155000.0,
					"low": 149000.0,
					"close": 152000.0,
					"volume": 10.5,
					"quantity": 5
				}
			]
		}`))
	}))
	defer mockServer.Close()

	// Configurar banco de dados de teste
	dbConfig := postgres.Config{
		Host:     "test-db",
		Port:     5432,
		User:     "test_user",
		Password: "test_password",
		DBName:   "test_db",
		SSLMode:  "disable",
	}

	// Criar banco de dados de teste
	db, err := postgres.NewConnection(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Limpar dados existentes
	_, err = db.Exec("TRUNCATE TABLE mms")
	require.NoError(t, err)

	// Configurar worker com o servidor mock
	cfg := &config.Config{
		Database:              dbConfig,
		MercadoBitcoinBaseURL: mockServer.URL,
		AlertConfig: struct {
			Enabled bool
			URL     string
		}{
			Enabled: false,
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
		ORDER BY timestamp DESC 
		LIMIT 1
	`)
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
	}

	// Verificar se o alerta foi enviado por email via MailHog
	mailhogClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := mailhogClient.Get("http://mailhog:8025/api/v2/messages")
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verificar se há mensagens no MailHog
	var messages struct {
		Items []struct {
			Content struct {
				Headers struct {
					Subject []string `json:"subject"`
					To      []string `json:"to"`
				} `json:"headers"`
				Body string `json:"body"`
			} `json:"Content"`
		} `json:"items"`
	}

	err = json.NewDecoder(resp.Body).Decode(&messages)
	require.NoError(t, err)

	// Deve haver pelo menos uma mensagem de alerta
	assert.Greater(t, len(messages.Items), 0, "Deveria haver mensagens de alerta")

	// Verificar o conteúdo do último alerta
	lastMessage := messages.Items[len(messages.Items)-1]
	assert.Contains(t, lastMessage.Content.Headers.Subject[0], "Alerta: dados_incompletos")
	assert.Contains(t, lastMessage.Content.Body, "Dados incompletos para BRLBTC")
}
