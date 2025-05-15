package mock

import (
	"context"
	"time"

	"mms_api/internal/domain/model"
)

// MockMMSRepository implementação mock do repositório
type MockMMSRepository struct{}

func NewMockMMSRepository() *MockMMSRepository {
	return &MockMMSRepository{}
}

func (m *MockMMSRepository) SaveMMS(ctx context.Context, mms model.MMS) error {
	return nil
}

func (m *MockMMSRepository) GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
	return []model.MMS{}, nil
}

// MockCandleAPI implementação mock da API de candles
type MockCandleAPI struct{}

func NewMockCandleAPI() *MockCandleAPI {
	return &MockCandleAPI{}
}

func (m *MockCandleAPI) GetCandles(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
	return []model.Candle{}, nil
}
