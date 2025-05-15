package out

import (
	"context"
	"time"

	"mms_api/internal/domain/model"
)

// MMSRepository define o contrato para persistência de MMSs
type MMSRepository interface {
	SaveMMS(ctx context.Context, mms model.MMS) error
	SaveBatch(ctx context.Context, mms []model.MMS) error
	GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error)
	FindByPairAndTimeRange(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error)
	CheckDataCompleteness(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error)
	GetLastTimestamp(ctx context.Context, pair string) (time.Time, error)
}
