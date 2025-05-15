package service

import (
	"context"
	"errors"
	"time"

	"mms_api/internal/application/port/out"
	"mms_api/internal/domain/model"
	"mms_api/pkg/logger"
)

// MMSService define o contrato para o serviço de MMS
type MMSService interface {
	// Calcular e salvar MMSs para um par em um intervalo
	CalculateAndSaveMMSForRange(ctx context.Context, pair string, from, to time.Time) error

	// Obter MMSs para um par em um intervalo com um período específico
	GetMMSByPairAndRange(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error)

	// Verificar completude dos dados nos últimos 365 dias
	CheckDataCompleteness(ctx context.Context, pair string) (bool, []time.Time, error)

	// Obter MMSs por par e timeframe
	GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error)
}

// mmsServiceImpl implementa a interface MMSService
type mmsServiceImpl struct {
	repo      out.MMSRepository
	candleAPI out.CandleAPI
	logger    logger.Logger
}

// NewMMSService cria uma nova instância do serviço
func NewMMSService(repo out.MMSRepository, candleAPI out.CandleAPI, logger logger.Logger) MMSService {
	return &mmsServiceImpl{
		repo:      repo,
		candleAPI: candleAPI,
		logger:    logger,
	}
}

// CalculateAndSaveMMSForRange implementa o cálculo e persistência de MMSs para um intervalo
func (s *mmsServiceImpl) CalculateAndSaveMMSForRange(ctx context.Context, pair string, from, to time.Time) error {
	// Validar par
	if !model.IsValidPair(pair) {
		return errors.New("par inválido")
	}

	// Precisamos de dados históricos suficientes para calcular a maior MMS (200 dias)
	historicalFrom := from.AddDate(0, 0, -200)

	// Buscar candles da API
	candles, err := s.candleAPI.GetCandles(ctx, pair, historicalFrom, to)
	if err != nil {
		s.logger.Error("falha ao obter candles", "error", err, "pair", pair)
		return err
	}

	if len(candles) < 200 {
		return errors.New("dados insuficientes para calcular MMS")
	}

	// Calcular MMSs
	var mmsEntries []model.MMS

	// A partir do índice 199 (ou seja, temos 200 dias de histórico)
	for i := 199; i < len(candles); i++ {
		// Se a data do candle é anterior à data solicitada, pulamos
		if candles[i].Timestamp.Before(from) {
			continue
		}

		// Calcular MMS 20, 50 e 200
		var sum20, sum50, sum200 float64

		for j := 0; j < 200; j++ {
			// Índice do candle a ser usado no cálculo
			idx := i - j
			closePrice := candles[idx].Close

			// Somar para cada período
			if j < 20 {
				sum20 += closePrice
			}
			if j < 50 {
				sum50 += closePrice
			}
			sum200 += closePrice
		}

		// Criar entrada de MMS
		mms := model.MMS{
			Pair:      pair,
			Timestamp: candles[i].Timestamp,
			MMS20:     sum20 / 20,
			MMS50:     sum50 / 50,
			MMS200:    sum200 / 200,
		}

		mmsEntries = append(mmsEntries, mms)
	}

	// Salvar no banco de dados
	if err := s.repo.SaveBatch(ctx, mmsEntries); err != nil {
		s.logger.Error("falha ao salvar MMSs", "error", err, "pair", pair)
		return err
	}

	return nil
}

// GetMMSByPairAndRange implementa a busca de MMSs por par e intervalo
func (s *mmsServiceImpl) GetMMSByPairAndRange(ctx context.Context, pair string, from, to time.Time, period int) ([]model.MMS, error) {
	// Validações
	if !model.IsValidPair(pair) {
		return nil, errors.New("par inválido")
	}

	if !model.IsValidPeriod(period) {
		return nil, errors.New("período inválido")
	}

	// Verificar se a data de início não é muito antiga (365 dias)
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	if from.Before(oneYearAgo) {
		return nil, errors.New("data de início não pode ser anterior a 365 dias")
	}

	// Buscar do repositório
	return s.repo.FindByPairAndTimeRange(ctx, pair, from, to, period)
}

// CheckDataCompleteness verifica se há dados para todos os dias dos últimos 365 dias
func (s *mmsServiceImpl) CheckDataCompleteness(ctx context.Context, pair string) (bool, []time.Time, error) {
	to := time.Now()
	from := to.AddDate(-1, 0, 0)

	return s.repo.CheckDataCompleteness(ctx, pair, from, to)
}

// GetMMSByPair implementa o método da interface para obter MMSs por par
func (s *mmsServiceImpl) GetMMSByPair(ctx context.Context, pair string, timeframe string) ([]model.MMS, error) {
	return s.repo.GetMMSByPair(ctx, pair, timeframe)
}
