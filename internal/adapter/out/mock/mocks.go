package mock

import (
	"context"
	"time"

	"mms_api/internal/domain/model"
)

// MockAlertMonitor é um mock do monitor de alertas para testes
type MockAlertMonitor struct {
	SendAlertFunc    func(alertType string, message string)
	SentEmailAlerts  []EmailAlert
	AlertTypesCalled []string
	MessagesSent     []string
}

type EmailAlert struct {
	Type    string
	Message string
	To      []string
}

func (m *MockAlertMonitor) SendAlert(alertType string, message string) {
	if m.SendAlertFunc != nil {
		m.SendAlertFunc(alertType, message)
	}
	m.AlertTypesCalled = append(m.AlertTypesCalled, alertType)
	m.MessagesSent = append(m.MessagesSent, message)
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

// MockService é um mock do serviço MMS para testes
type MockService struct {
	CalculateAndSaveMMSForRangeFunc func(ctx context.Context, pair string, from, to time.Time) error
	GetMMSByPairAndRangeFunc        func(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error)
	CheckDataCompletenessFunc       func(ctx context.Context, pair string) (bool, []time.Time, error)
	GetMMSByPairFunc                func(ctx context.Context, pair string, timeframe string) ([]model.MMS, error)
}

func (m *MockService) CalculateAndSaveMMSForRange(ctx context.Context, pair string, from, to time.Time) error {
	if m.CalculateAndSaveMMSForRangeFunc != nil {
		return m.CalculateAndSaveMMSForRangeFunc(ctx, pair, from, to)
	}
	return nil
}

func (m *MockService) GetMMSByPairAndRange(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
	if m.GetMMSByPairAndRangeFunc != nil {
		return m.GetMMSByPairAndRangeFunc(ctx, pair, from, to, period)
	}
	return nil, nil
}

func (m *MockService) CheckDataCompleteness(ctx context.Context, pair string) (bool, []time.Time, error) {
	if m.CheckDataCompletenessFunc != nil {
		return m.CheckDataCompletenessFunc(ctx, pair)
	}
	return true, nil, nil
}

func (m *MockService) GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
	if m.GetMMSByPairFunc != nil {
		return m.GetMMSByPairFunc(ctx, pair, timeframe)
	}
	return nil, nil
}
