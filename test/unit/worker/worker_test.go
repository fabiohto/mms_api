package worker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mms_api/cmd/worker/bootstrap"
	"mms_api/config"
	"mms_api/internal/application/service"
	"mms_api/internal/domain/model"
	"mms_api/pkg/logger"
	"mms_api/test/unit/mock"

	"github.com/stretchr/testify/assert"
)

func TestWorker_Run(t *testing.T) {
	// Data de referência para testes
	now := time.Date(2025, 5, 15, 0, 0, 0, 0, time.UTC)
	yesterday := now.AddDate(0, 0, -1)
	lastYear := now.AddDate(-1, 0, 0)

	tests := []struct {
		name           string
		setupMocks     func(*mock.MockMMSRepository, *mock.MockCandleAPI, *mock.MockAlertMonitor)
		expectedAlerts []string
		expectedEmails []mock.EmailAlert
		wantErr        bool
	}{
		{
			name: "processamento bem sucedido sem dados anteriores",
			setupMocks: func(repo *mock.MockMMSRepository, api *mock.MockCandleAPI, monitor *mock.MockAlertMonitor) {
				// Mock GetLastTimestamp retornando zero (sem dados)
				repo.GetLastTimestampFunc = func(ctx context.Context, pair string) (time.Time, error) {
					return time.Time{}, nil
				}

				// Mock GetCandles retornando dados históricos
				api.GetCandlesFunc = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					return []model.Candle{
						{
							Timestamp: yesterday,
							Close:     150000.0,
						},
						{
							Timestamp: yesterday.Add(-24 * time.Hour),
							Close:     148000.0,
						},
					}, nil
				}

				// Mock SendAlert para verificar se nenhum alerta é enviado
				monitor.SendAlertFunc = func(alertType string, message string) {
					t.Error("Não deveria enviar alertas em caso de sucesso")
				}

				// Mock SaveBatch sem erros
				repo.SaveBatchFunc = func(ctx context.Context, mms []model.MMS) error {
					return nil
				}

				// Mock CheckDataCompleteness retornando sucesso
				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return true, nil, nil
				}
			},
			wantErr: false,
		},
		{
			name: "erro ao obter candles deve gerar retry e alerta",
			setupMocks: func(repo *mock.MockMMSRepository, api *mock.MockCandleAPI, monitor *mock.MockAlertMonitor) {
				var attempts int

				repo.GetLastTimestampFunc = func(ctx context.Context, pair string) (time.Time, error) {
					return lastYear, nil
				}

				api.GetCandlesFunc = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					attempts++
					return nil, errors.New("erro de conexão")
				}

				monitor.SendAlertFunc = func(alertType string, message string) {
					assert.Equal(t, "falha_atualizacao", alertType)
				}
			},
			expectedAlerts: []string{"falha_atualizacao"},
			wantErr:        false,
		},
		{
			name: "dados incompletos devem gerar alerta",
			setupMocks: func(repo *mock.MockMMSRepository, api *mock.MockCandleAPI, monitor *mock.MockAlertMonitor) {
				repo.GetLastTimestampFunc = func(ctx context.Context, pair string) (time.Time, error) {
					return lastYear, nil
				}

				api.GetCandlesFunc = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					return []model.Candle{
						{
							Timestamp: yesterday,
							Close:     150000.0,
						},
					}, nil
				}

				repo.SaveBatchFunc = func(ctx context.Context, mms []model.MMS) error {
					return nil
				}

				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return false, []time.Time{yesterday.Add(-48 * time.Hour)}, nil
				}

				monitor.SendAlertFunc = func(alertType string, message string) {
					assert.Equal(t, "dados_incompletos", alertType)
				}
			},
			expectedAlerts: []string{"dados_incompletos"},
			wantErr:        false,
		},
		{
			name: "deve enviar alerta por email quando dados estiverem incompletos",
			setupMocks: func(repo *mock.MockMMSRepository, api *mock.MockCandleAPI, monitor *mock.MockAlertMonitor) {
				repo.GetLastTimestampFunc = func(ctx context.Context, pair string) (time.Time, error) {
					return lastYear, nil
				}

				api.GetCandlesFunc = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					return []model.Candle{
						{
							Timestamp: yesterday,
							Close:     150000.0,
						},
					}, nil
				}

				repo.SaveBatchFunc = func(ctx context.Context, mms []model.MMS) error {
					return nil
				}

				// Simular dados incompletos
				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					missingDates := []time.Time{yesterday.Add(-48 * time.Hour)}
					return false, missingDates, nil
				}

				var alertsCalled []string
				monitor.SendAlertFunc = func(alertType string, message string) {
					alertsCalled = append(alertsCalled, alertType)
					monitor.SentEmailAlerts = append(monitor.SentEmailAlerts, mock.EmailAlert{
						Type:    alertType,
						Message: message,
						To:      []string{"destino@email.com"},
					})
				}
			},
			expectedAlerts: []string{"dados_incompletos"},
			expectedEmails: []mock.EmailAlert{
				{
					Type:    "dados_incompletos",
					Message: "Dados incompletos para BRLBTC",
					To:      []string{"destino@email.com"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Criar mocks
			mockRepo := &mock.MockMMSRepository{}
			mockAPI := &mock.MockCandleAPI{}
			mockMonitor := &mock.MockAlertMonitor{}

			// Configurar mocks
			tt.setupMocks(mockRepo, mockAPI, mockMonitor)

			// Criar configuração
			cfg := &config.Config{
				AlertConfig: struct {
					Enabled bool
					URL     string
				}{
					Enabled: true,
					URL:     "http://alert-service",
				},
			}

			// Criar logger
			l := logger.NewLogger("[TEST] ")

			// Criar worker com os mocks
			worker := &bootstrap.Worker{
				MMSService:   service.NewMMSService(mockRepo, mockAPI, l),
				MMSRepo:      mockRepo,
				AlertMonitor: mockMonitor,
				Logger:       l,
			}

			// Executar worker
			err := worker.Run()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
