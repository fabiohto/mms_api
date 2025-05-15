package mock

import (
	"context"
	"time"

	"mms_api/internal/domain/model"
)

// MockMMSRepository é um mock do repositório MMS para testes
type MockMMSRepository struct {
	SaveBatchFunc             func(ctx context.Context, mms []model.MMS) error
	FindByPairAndRangeFunc    func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error)
	CheckDataCompletenessFunc func(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error)
	GetLastTimestampFunc      func(ctx context.Context, pair string) (time.Time, error)
	GetMMSByPairFunc          func(ctx context.Context, pair string, timeframe string) ([]model.MMS, error)
	SaveMMSFunc               func(ctx context.Context, mms model.MMS) error
}

func (m *MockMMSRepository) SaveBatch(ctx context.Context, mms []model.MMS) error {
	if m.SaveBatchFunc != nil {
		return m.SaveBatchFunc(ctx, mms)
	}
	return nil
}

func (m *MockMMSRepository) FindByPairAndTimeRange(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
	if m.FindByPairAndRangeFunc != nil {
		return m.FindByPairAndRangeFunc(ctx, pair, from, to, period)
	}
	return nil, nil
}

func (m *MockMMSRepository) CheckDataCompleteness(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
	if m.CheckDataCompletenessFunc != nil {
		return m.CheckDataCompletenessFunc(ctx, pair, from, to)
	}
	return true, nil, nil
}

func (m *MockMMSRepository) GetLastTimestamp(ctx context.Context, pair string) (time.Time, error) {
	if m.GetLastTimestampFunc != nil {
		return m.GetLastTimestampFunc(ctx, pair)
	}
	return time.Time{}, nil
}

func (m *MockMMSRepository) GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
	if m.GetMMSByPairFunc != nil {
		return m.GetMMSByPairFunc(ctx, pair, timeframe)
	}
	return nil, nil
}

func (m *MockMMSRepository) SaveMMS(ctx context.Context, mms model.MMS) error {
	if m.SaveMMSFunc != nil {
		return m.SaveMMSFunc(ctx, mms)
	}
	return nil
}

// MockCandleAPI é um mock da API de candles para testes
type MockCandleAPI struct {
	GetCandlesFunc func(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error)
}

func (m *MockCandleAPI) GetCandles(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
	if m.GetCandlesFunc != nil {
		return m.GetCandlesFunc(ctx, pair, from, to)
	}
	return nil, nil
}

// MockLogger é um mock do logger para testes
type MockLogger struct {
	InfoFunc  func(args ...interface{})
	ErrorFunc func(args ...interface{})
	FatalFunc func(args ...interface{})
}

func (m *MockLogger) Info(args ...interface{}) {
	if m.InfoFunc != nil {
		m.InfoFunc(args...)
	}
}

func (m *MockLogger) Error(args ...interface{}) {
	if m.ErrorFunc != nil {
		m.ErrorFunc(args...)
	}
}

func (m *MockLogger) Fatal(args ...interface{}) {
	if m.FatalFunc != nil {
		m.FatalFunc(args...)
	}
}
