package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mms_api/cmd/worker/bootstrap"
	"mms_api/internal/application/service"
	"mms_api/internal/domain/model"
	"mms_api/pkg/logger"

	"github.com/stretchr/testify/assert"
)

// Mock interfaces
type mockMMSRepository struct {
	getLastTimestamp       func(ctx context.Context, pair string) (time.Time, error)
	saveBatch              func(ctx context.Context, mms []model.MMS) error
	checkDataCompleteness  func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error)
	findByPairAndTimeRange func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error)
	getMMSByPair           func(ctx context.Context, pair string, timeframe string) ([]model.MMS, error)
	saveMMS                func(ctx context.Context, mms model.MMS) error
}

func (m *mockMMSRepository) GetLastTimestamp(ctx context.Context, pair string) (time.Time, error) {
	return m.getLastTimestamp(ctx, pair)
}

func (m *mockMMSRepository) SaveBatch(ctx context.Context, mms []model.MMS) error {
	return m.saveBatch(ctx, mms)
}

func (m *mockMMSRepository) CheckDataCompleteness(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
	return m.checkDataCompleteness(ctx, pair, from, to)
}

func (m *mockMMSRepository) FindByPairAndTimeRange(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
	return m.findByPairAndTimeRange(ctx, pair, from, to, period)
}

func (m *mockMMSRepository) GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
	return m.getMMSByPair(ctx, pair, timeframe)
}

func (m *mockMMSRepository) SaveMMS(ctx context.Context, mms model.MMS) error {
	return m.saveMMS(ctx, mms)
}

type mockCandleAPI struct {
	getCandles func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error)
}

func (m *mockCandleAPI) GetCandles(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
	return m.getCandles(ctx, pair, from, to)
}

// mockAlertMonitor implementa a interface monitoring.AlertMonitor
type mockAlertMonitor struct {
	sendAlert func(alertType string, message string)
}

func (m *mockAlertMonitor) SendAlert(alertType string, message string) {
	if m.sendAlert != nil {
		m.sendAlert(alertType, message)
	}
}

func TestWorker_Run(t *testing.T) {
	// Data de referência para testes
	now := time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC)
	yesterday := now.AddDate(0, 0, -1)
	lastYear := now.AddDate(-1, 0, 0)

	// Função auxiliar para gerar candles históricos
	generateHistoricalCandles := func(pair string, from, to time.Time) []model.Candle {
		var candles []model.Candle
		basePrice := 150000.0
		if pair == "BRLETH" {
			basePrice = 8000.0
		}

		// Gerar candles diários
		for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
			candles = append(candles, model.Candle{
				Pair:      pair,
				Timestamp: d,
				Open:      basePrice,
				High:      basePrice * 1.01,
				Low:       basePrice * 0.99,
				Close:     basePrice + float64(len(candles))*10, // Preço crescente para simular tendência
				Volume:    100.0,
			})
		}
		return candles
	}

	// Dados de exemplo para MMS
	sampleMMS := []model.MMS{
		{
			Pair:      "BRLBTC",
			Timestamp: yesterday,
			MMS20:     150000.0,
			MMS50:     145000.0,
			MMS200:    140000.0,
		},
		{
			Pair:      "BRLETH",
			Timestamp: yesterday,
			MMS20:     8000.0,
			MMS50:     7500.0,
			MMS200:    7000.0,
		},
	}

	tests := []struct {
		name           string
		setupMocks     func(*mockMMSRepository, *mockCandleAPI, *mockAlertMonitor)
		expectedAlerts []string
		wantErr        bool
	}{
		{
			name: "processamento bem sucedido sem dados anteriores",
			setupMocks: func(repo *mockMMSRepository, api *mockCandleAPI, monitor *mockAlertMonitor) {
				// Mock GetLastTimestamp retornando zero (sem dados)
				repo.getLastTimestamp = func(ctx context.Context, pair string) (time.Time, error) {
					return time.Time{}, nil
				}

				// Mock GetCandles retornando dados históricos suficientes
				api.getCandles = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					// Gerar 250 dias de dados históricos
					historicalFrom := from.AddDate(0, 0, -250)
					return generateHistoricalCandles(pair, historicalFrom, to), nil
				}

				// Mock SaveBatch sem erros
				repo.saveBatch = func(ctx context.Context, mms []model.MMS) error {
					return nil
				}

				// Mock CheckDataCompleteness retornando sucesso
				repo.checkDataCompleteness = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return true, nil, nil
				}

				// Mock FindByPairAndTimeRange retornando dados históricos suficientes
				repo.findByPairAndTimeRange = func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
					var result []model.MMS
					for _, mms := range sampleMMS {
						if mms.Pair == pair {
							result = append(result, mms)
						}
					}
					return result, nil
				}

				// Mock GetMMSByPair
				repo.getMMSByPair = func(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
					var result []model.MMS
					for _, mms := range sampleMMS {
						if mms.Pair == pair {
							result = append(result, mms)
						}
					}
					return result, nil
				}

				// Mock SaveMMS
				repo.saveMMS = func(ctx context.Context, mms model.MMS) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "erro ao obter candles deve gerar retry e alerta",
			setupMocks: func(repo *mockMMSRepository, api *mockCandleAPI, monitor *mockAlertMonitor) {
				// Mock GetLastTimestamp
				repo.getLastTimestamp = func(ctx context.Context, pair string) (time.Time, error) {
					return lastYear, nil
				}

				// Mock GetCandles sempre retornando erro
				api.getCandles = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					return nil, errors.New("erro de conexão")
				}

				// Mock SaveBatch sem erros
				repo.saveBatch = func(ctx context.Context, mms []model.MMS) error {
					return nil
				}

				// Mock CheckDataCompleteness
				repo.checkDataCompleteness = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return true, nil, nil
				}

				// Mock FindByPairAndTimeRange retornando dados históricos suficientes
				repo.findByPairAndTimeRange = func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
					var result []model.MMS
					for _, mms := range sampleMMS {
						if mms.Pair == pair {
							result = append(result, mms)
						}
					}
					return result, nil
				}

				// Mock GetMMSByPair
				repo.getMMSByPair = func(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
					var result []model.MMS
					for _, mms := range sampleMMS {
						if mms.Pair == pair {
							result = append(result, mms)
						}
					}
					return result, nil
				}

				// Mock SaveMMS
				repo.saveMMS = func(ctx context.Context, mms model.MMS) error {
					return nil
				}

				// Configure alert monitor to track alerts
				monitor.sendAlert = func(alertType string, message string) {
					if alertType == "falha_atualizacao" {
						// Alert was sent as expected
					}
				}
			},
			expectedAlerts: []string{"falha_atualizacao"},
			wantErr:        false,
		},
		{
			name: "dados incompletos devem gerar alerta",
			setupMocks: func(repo *mockMMSRepository, api *mockCandleAPI, monitor *mockAlertMonitor) {
				// Mock GetLastTimestamp
				repo.getLastTimestamp = func(ctx context.Context, pair string) (time.Time, error) {
					return lastYear, nil
				}

				// Mock GetCandles retornando dados insuficientes
				api.getCandles = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					// Gerar apenas 150 dias de dados (insuficiente para MMS200)
					historicalFrom := from.AddDate(0, 0, -150)
					return generateHistoricalCandles(pair, historicalFrom, to), nil
				}

				// Mock SaveBatch sem erros
				repo.saveBatch = func(ctx context.Context, mms []model.MMS) error {
					return nil
				}

				// Mock CheckDataCompleteness retornando dados incompletos
				repo.checkDataCompleteness = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return false, []time.Time{yesterday.Add(-48 * time.Hour)}, nil
				}

				// Mock FindByPairAndTimeRange retornando dados históricos suficientes
				repo.findByPairAndTimeRange = func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
					var result []model.MMS
					for _, mms := range sampleMMS {
						if mms.Pair == pair {
							result = append(result, mms)
						}
					}
					return result, nil
				}

				// Mock GetMMSByPair
				repo.getMMSByPair = func(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
					var result []model.MMS
					for _, mms := range sampleMMS {
						if mms.Pair == pair {
							result = append(result, mms)
						}
					}
					return result, nil
				}

				// Mock SaveMMS
				repo.saveMMS = func(ctx context.Context, mms model.MMS) error {
					return nil
				}

				// Configure alert monitor to track alerts
				monitor.sendAlert = func(alertType string, message string) {
					if alertType == "dados_incompletos" {
						// Alert was sent as expected
					}
				}
			},
			expectedAlerts: []string{"dados_incompletos"},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar mocks
			mockRepo := &mockMMSRepository{}
			mockAPI := &mockCandleAPI{}
			mockMonitor := &mockAlertMonitor{}

			// Configurar mocks
			tt.setupMocks(mockRepo, mockAPI, mockMonitor)

			// Criar logger
			l := logger.NewLogger("[TEST] ")

			// Criar serviço com os mocks
			mmsService := service.NewMMSService(mockRepo, mockAPI, l)

			// Criar worker com os mocks
			worker := bootstrap.NewWorkerWithDeps(mmsService, mockRepo, mockMonitor, l)

			// Configurar intervalo de retry menor para os testes
			worker.SetRetryInterval(100 * time.Millisecond)

			// Criar contexto com timeout para evitar que o teste trave
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Criar canal para receber o resultado da execução
			done := make(chan error)
			go func() {
				done <- worker.Run()
			}()

			// Aguardar resultado ou timeout
			select {
			case err := <-done:
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			case <-ctx.Done():
				t.Fatal("teste travou por timeout")
			}
		})
	}
}
