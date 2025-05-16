package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mms_api/internal/adapter/out/mock"
	"mms_api/internal/application/service"
	"mms_api/internal/domain/model"
	"mms_api/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAndSaveMMSForRange(t *testing.T) {
	t.Parallel() // Paralelizar teste
	now := time.Now()
	ctx := context.Background()

	tests := []struct {
		name      string
		pair      string
		from      time.Time
		to        time.Time
		setupMock func(*mock.MockMMSRepository, *mock.MockCandleAPI)
		wantErr   bool
	}{
		{
			name: "deve calcular e salvar MMS com sucesso",
			pair: "BRLBTC",
			from: now.AddDate(0, 0, -1),
			to:   now,
			setupMock: func(repo *mock.MockMMSRepository, api *mock.MockCandleAPI) {
				// Mock para retornar candles
				api.GetCandlesFunc = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					candles := make([]model.Candle, 250) // Dados suficientes para calcular MMS
					for i := range candles {
						candles[i] = model.Candle{
							Pair:      pair,
							Timestamp: from.AddDate(0, 0, i),
							Open:      45000.0,
							High:      46000.0,
							Low:       44000.0,
							Close:     45500.0,
						}
					}
					return candles, nil
				}

				// Mock para salvar MMSs
				repo.SaveBatchFunc = func(ctx context.Context, mms []model.MMS) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "deve retornar erro para par inválido",
			pair: "INVALID",
			from: now.AddDate(0, 0, -1),
			to:   now,
			setupMock: func(repo *mock.MockMMSRepository, api *mock.MockCandleAPI) {
				// Não precisa configurar mocks pois deve falhar na validação
			},
			wantErr: true,
		},
		{
			name: "deve retornar erro quando falha ao obter candles",
			pair: "BRLBTC",
			from: now.AddDate(0, 0, -1),
			to:   now,
			setupMock: func(repo *mock.MockMMSRepository, api *mock.MockCandleAPI) {
				api.GetCandlesFunc = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
					return nil, errors.New("erro ao obter candles")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mock.MockMMSRepository{}
			api := &mock.MockCandleAPI{}
			log := logger.NewLogger("[TEST] ")

			if tt.setupMock != nil {
				tt.setupMock(repo, api)
			}

			svc := service.NewMMSService(repo, api, log)

			// Act
			err := svc.CalculateAndSaveMMSForRange(ctx, tt.pair, tt.from, tt.to)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateAndSaveMMSForRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetMMSByPairAndRange(t *testing.T) {
	t.Parallel() // Paralelizar teste
	now := time.Now()
	ctx := context.Background()

	tests := []struct {
		name      string
		pair      string
		from      time.Time
		to        time.Time
		period    int
		setupMock func(*mock.MockMMSRepository)
		wantErr   bool
	}{
		{
			name:   "deve retornar MMS com sucesso",
			pair:   "BRLBTC",
			from:   now.AddDate(0, 0, -10),
			to:     now,
			period: model.Period20,
			setupMock: func(repo *mock.MockMMSRepository) {
				repo.FindByPairAndRangeFunc = func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
					return []model.MMS{
						{
							Pair:      pair,
							Timestamp: now,
							MMS20:     50000.0,
							MMS50:     49000.0,
							MMS200:    48000.0,
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:   "deve retornar erro para par inválido",
			pair:   "INVALID",
			from:   now.AddDate(0, 0, -10),
			to:     now,
			period: model.Period20,
			setupMock: func(repo *mock.MockMMSRepository) {
				// Não precisa configurar mock pois deve falhar na validação
			},
			wantErr: true,
		},
		{
			name:   "deve retornar erro para período inválido",
			pair:   "BRLBTC",
			from:   now.AddDate(0, 0, -10),
			to:     now,
			period: 30, // período inválido
			setupMock: func(repo *mock.MockMMSRepository) {
				// Não precisa configurar mock pois deve falhar na validação
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMMSRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			log := logger.NewLogger("[TEST] ")
			candleAPI := &mock.MockCandleAPI{}

			svc := service.NewMMSService(repo, candleAPI, log)

			result, err := svc.GetMMSByPairAndRange(ctx, tt.pair, tt.from, tt.to, tt.period)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result)
		})
	}
}

func TestCheckDataCompleteness(t *testing.T) {
	t.Parallel() // Paralelizar teste
	ctx := context.Background()

	tests := []struct {
		name      string
		pair      string
		setupMock func(*mock.MockMMSRepository)
		wantErr   bool
	}{
		{
			name: "deve retornar completude com sucesso",
			pair: "BRLBTC",
			setupMock: func(repo *mock.MockMMSRepository) {
				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return true, nil, nil
				}
			},
			wantErr: false,
		},
		{
			name: "deve retornar erro para par inválido",
			pair: "INVALID",
			setupMock: func(repo *mock.MockMMSRepository) {
				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return false, nil, errors.New("par inválido")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMMSRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			log := logger.NewLogger("[TEST] ")
			candleAPI := &mock.MockCandleAPI{}

			svc := service.NewMMSService(repo, candleAPI, log)

			isComplete, missingDates, err := svc.CheckDataCompleteness(ctx, tt.pair)
			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, isComplete)
				assert.Nil(t, missingDates)
				return
			}

			assert.NoError(t, err)
			if tt.pair == "BRLBTC" {
				assert.True(t, isComplete)
				assert.Empty(t, missingDates)
			}
		})
	}
}

func TestAlertOnAPIError(t *testing.T) {
	t.Parallel() // Paralelizar teste
	ctx := context.Background()
	now := time.Now()

	// Setup mocks
	repo := &mock.MockMMSRepository{}
	api := &mock.MockCandleAPI{}
	log := logger.NewLogger("[TEST] ")

	expectedError := errors.New("erro de conexão com a API")

	// Configure API to return error
	api.GetCandlesFunc = func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
		return nil, expectedError
	}

	svc := service.NewMMSService(repo, api, log)

	// Test error case
	err := svc.CalculateAndSaveMMSForRange(ctx, "BRLBTC", now.Add(-24*time.Hour), now)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}
