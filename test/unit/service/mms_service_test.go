package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"mms_api/internal/adapter/out/mock"
	"mms_api/internal/application/service"
	"mms_api/internal/domain/model"
)

func TestCalculateAndSaveMMSForRange(t *testing.T) {
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
			logger := &mock.MockLogger{}

			if tt.setupMock != nil {
				tt.setupMock(repo, api)
			}

			svc := service.NewMMSService(repo, api, logger)

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
			name:   "deve retornar MMSs com sucesso",
			pair:   "BRLBTC",
			from:   now.AddDate(0, 0, -10),
			to:     now,
			period: model.Period20,
			setupMock: func(repo *mock.MockMMSRepository) {
				repo.FindByPairAndTimeRange = func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
					return []model.MMS{
						{
							Pair:      pair,
							Timestamp: now,
							MMS20:     50000.0,
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
			// Arrange
			repo := &mock.MockMMSRepository{}
			api := &mock.MockCandleAPI{}
			logger := &mock.MockLogger{}

			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			svc := service.NewMMSService(repo, api, logger)

			// Act
			mms, err := svc.GetMMSByPairAndRange(ctx, tt.pair, tt.from, tt.to, tt.period)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMMSByPairAndRange() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && len(mms) == 0 {
				t.Error("GetMMSByPairAndRange() returned empty result when success was expected")
			}
		})
	}
}

func TestCheckDataCompleteness(t *testing.T) {
	ctx := context.Background()
	pair := "BRLBTC"

	tests := []struct {
		name      string
		setupMock func(*mock.MockMMSRepository)
		want      bool
		wantErr   bool
	}{
		{
			name: "deve retornar true quando dados estão completos",
			setupMock: func(repo *mock.MockMMSRepository) {
				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return true, nil, nil
				}
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "deve retornar false quando há dados faltantes",
			setupMock: func(repo *mock.MockMMSRepository) {
				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					missing := []time.Time{time.Now().AddDate(0, 0, -1)}
					return false, missing, nil
				}
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "deve retornar erro quando falha ao verificar",
			setupMock: func(repo *mock.MockMMSRepository) {
				repo.CheckDataCompletenessFunc = func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
					return false, nil, errors.New("erro ao verificar completude")
				}
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			repo := &mock.MockMMSRepository{}
			api := &mock.MockCandleAPI{}
			logger := &mock.MockLogger{}

			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			svc := service.NewMMSService(repo, api, logger)

			// Act
			got, _, err := svc.CheckDataCompleteness(ctx, pair)

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckDataCompleteness() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("CheckDataCompleteness() = %v, want %v", got, tt.want)
			}
		})
	}
}
