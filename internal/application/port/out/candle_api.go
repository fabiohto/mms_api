package out

import (
	"context"
	"time"

	"mms_api/internal/domain/model"
)

// CandleAPI define o contrato para obtenção de candles
type CandleAPI interface {
	GetCandles(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error)
}
