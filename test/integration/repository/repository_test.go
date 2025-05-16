package repository_test

import (
	"context"
	"testing"
	"time"

	"mms_api/internal/adapter/out/persistence/postgres"
	"mms_api/internal/domain/model"
	pgdb "mms_api/pkg/db/postgres"
	"mms_api/pkg/logger"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMMSRepository_Integration(t *testing.T) {
	// Configurar banco de dados de teste
	dbConfig := pgdb.Config{
		Host:     "test-db",
		Port:     "5432",
		User:     "test_user",
		Password: "test_password",
		DBName:   "test_db",
	}

	// Criar conexão com timeout
	db, err := pgdb.NewConnectionWithTimeout(dbConfig)
	require.NoError(t, err)
	defer db.Close()

	// Criar logger para testes
	l := logger.NewLogger("[TEST] ")

	// Criar repositório
	repo := postgres.NewMMSRepository(db, l)

	// Limpar dados existentes
	_, err = db.Exec("TRUNCATE TABLE mms")
	require.NoError(t, err)

	t.Run("SaveBatch e FindByPairAndTimeRange", func(t *testing.T) {
		ctx := context.Background()
		now := time.Now()

		// Criar dados de teste
		testData := []model.MMS{
			{
				Pair:      "BRLBTC",
				Timestamp: now,
				MMS20:     150000.0,
				MMS50:     148000.0,
				MMS200:    145000.0,
			},
			{
				Pair:      "BRLBTC",
				Timestamp: now.Add(-24 * time.Hour),
				MMS20:     149000.0,
				MMS50:     147000.0,
				MMS200:    144000.0,
			},
		}

		// Salvar dados
		err := repo.SaveBatch(ctx, testData)
		require.NoError(t, err)

		// Buscar dados
		from := now.Add(-48 * time.Hour)
		to := now.Add(24 * time.Hour)
		result, err := repo.FindByPairAndTimeRange(ctx, "BRLBTC", from, to, model.Period20)
		require.NoError(t, err)

		// Verificar resultados
		assert.Len(t, result, 2)
		assert.Equal(t, testData[0].Pair, result[0].Pair)
		assert.Equal(t, testData[0].MMS20, result[0].MMS20)
	})

	t.Run("CheckDataCompleteness", func(t *testing.T) {
		ctx := context.Background()
		now := time.Now()

		// Criar dados com um gap
		testData := []model.MMS{
			{
				Pair:      "BRLETH",
				Timestamp: now,
				MMS20:     2500.0,
				MMS50:     2400.0,
				MMS200:    2300.0,
			},
			// Gap de um dia aqui
			{
				Pair:      "BRLETH",
				Timestamp: now.Add(-48 * time.Hour),
				MMS20:     2400.0,
				MMS50:     2300.0,
				MMS200:    2200.0,
			},
		}

		// Salvar dados
		err := repo.SaveBatch(ctx, testData)
		require.NoError(t, err)

		// Verificar completude
		from := now.Add(-48 * time.Hour)
		to := now
		isComplete, missingDates, err := repo.CheckDataCompleteness(ctx, "BRLETH", from, to)
		require.NoError(t, err)

		assert.False(t, isComplete)
		assert.NotEmpty(t, missingDates)
	})

	t.Run("GetMMSByPair", func(t *testing.T) {
		ctx := context.Background()

		// Buscar dados para um timeframe específico
		result, err := repo.GetMMSByPair(ctx, "BRLBTC", "1d")
		require.NoError(t, err)

		// Verificar resultados
		assert.NotEmpty(t, result)
		for _, mms := range result {
			assert.Equal(t, "BRLBTC", mms.Pair)
		}
	})
}
