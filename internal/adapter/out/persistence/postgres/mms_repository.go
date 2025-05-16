package postgres

import (
	"context"
	"database/sql"
	"time"

	"mms_api/internal/domain/model"
	"mms_api/pkg/logger"
)

type MMSRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewMMSRepository(db *sql.DB, logger logger.Logger) *MMSRepository {
	return &MMSRepository{
		db:     db,
		logger: logger,
	}
}

func (r *MMSRepository) GetLastTimestamp(ctx context.Context, pair string) (time.Time, error) {
	var timestamp sql.NullTime
	query := `SELECT MAX(timestamp) FROM mms WHERE pair = $1`

	err := r.db.QueryRowContext(ctx, query, pair).Scan(&timestamp)
	if err == sql.ErrNoRows || !timestamp.Valid {
		return time.Time{}, nil
	}
	if err != nil {
		r.logger.Error("Erro ao buscar último timestamp", err)
		return time.Time{}, err
	}

	return timestamp.Time, nil
}

func (r *MMSRepository) SaveMMS(ctx context.Context, mms model.MMS) error {
	query := `
		INSERT INTO mms (pair, timestamp, mms20, mms50, mms200)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (pair, timestamp)
		DO UPDATE SET 
			mms20 = EXCLUDED.mms20,
			mms50 = EXCLUDED.mms50,
			mms200 = EXCLUDED.mms200
	`

	_, err := r.db.ExecContext(ctx, query, mms.Pair, mms.Timestamp, mms.MMS20, mms.MMS50, mms.MMS200)
	if err != nil {
		r.logger.Error("Erro ao salvar MMS", err)
		return err
	}

	return nil
}

func (r *MMSRepository) SaveBatch(ctx context.Context, mms []model.MMS) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Erro ao iniciar transação", err)
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO mms (pair, timestamp, mms20, mms50, mms200)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (pair, timestamp)
		DO UPDATE SET 
			mms20 = EXCLUDED.mms20,
			mms50 = EXCLUDED.mms50,
			mms200 = EXCLUDED.mms200
	`)
	if err != nil {
		r.logger.Error("Erro ao preparar statement", err)
		return err
	}
	defer stmt.Close()

	for _, m := range mms {
		_, err = stmt.ExecContext(ctx, m.Pair, m.Timestamp, m.MMS20, m.MMS50, m.MMS200)
		if err != nil {
			r.logger.Error("Erro ao salvar MMS", err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		r.logger.Error("Erro ao commitar transação", err)
		return err
	}

	return nil
}

func (r *MMSRepository) FindByPairAndTimeRange(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
	query := `
		SELECT pair, timestamp, mms20, mms50, mms200
		FROM mms
		WHERE pair = $1
		AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, pair, from, to)
	if err != nil {
		r.logger.Error("Erro ao buscar MMS", err)
		return nil, err
	}
	defer rows.Close()

	var result []model.MMS
	for rows.Next() {
		var mms model.MMS
		err := rows.Scan(&mms.Pair, &mms.Timestamp, &mms.MMS20, &mms.MMS50, &mms.MMS200)
		if err != nil {
			r.logger.Error("Erro ao ler MMS do banco", err)
			return nil, err
		}
		result = append(result, mms)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Erro ao iterar sobre resultados", err)
		return nil, err
	}

	return result, nil
}

func (r *MMSRepository) GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
	query := `
		SELECT pair, timestamp, mms20, mms50, mms200
		FROM mms
		WHERE pair = $1
		AND timestamp >= NOW() - $2::interval
		ORDER BY timestamp ASC
	`

	rows, err := r.db.QueryContext(ctx, query, pair, timeframe)
	if err != nil {
		r.logger.Error("Erro ao buscar MMS", err)
		return nil, err
	}
	defer rows.Close()

	var result []model.MMS
	for rows.Next() {
		var mms model.MMS
		err := rows.Scan(&mms.Pair, &mms.Timestamp, &mms.MMS20, &mms.MMS50, &mms.MMS200)
		if err != nil {
			r.logger.Error("Erro ao ler MMS do banco", err)
			return nil, err
		}
		result = append(result, mms)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Erro ao iterar sobre resultados", err)
		return nil, err
	}

	return result, nil
}

func (r *MMSRepository) CheckDataCompleteness(ctx context.Context, pair string, from, to time.Time) (bool, []time.Time, error) {
	// Primeiro, vamos buscar as datas que temos dados
	query := `
		SELECT DISTINCT timestamp::date
		FROM mms
		WHERE pair = $1
		AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp::date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, pair, from, to)
	if err != nil {
		r.logger.Error("Erro ao buscar timestamps", err)
		return false, nil, err
	}
	defer rows.Close()

	// Criar mapa de datas existentes
	existingDates := make(map[time.Time]bool)
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			r.logger.Error("Erro ao ler timestamp", err)
			return false, nil, err
		}
		existingDates[date] = true
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("Erro ao iterar sobre timestamps", err)
		return false, nil, err
	}

	// Verificar cada data no intervalo
	var missingDates []time.Time
	for current := from; !current.After(to); current = current.AddDate(0, 0, 1) {
		currentDate := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.UTC)
		if !existingDates[currentDate] {
			missingDates = append(missingDates, currentDate)
		}
	}

	return len(missingDates) == 0, missingDates, nil
}
