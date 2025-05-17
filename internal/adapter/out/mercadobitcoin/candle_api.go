package mercadobitcoin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"mms_api/internal/domain/model"
	"mms_api/pkg/logger"
)

// Struct para parsing do retorno da API oficial
type apiResponse struct {
	T []int64  `json:"t"`
	O []string `json:"o"`
	C []string `json:"c"`
	H []string `json:"h"`
	L []string `json:"l"`
	V []string `json:"v"`
	Q []string `json:"q"`
}

//type apiResponse struct {
//	Candles []apiCandle `json:"candles"`
//}

// CandleAPI encapsula a comunicação com a API do Mercado Bitcoin
type CandleAPI struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

// NewCandleAPI cria uma nova instância do cliente da API
func NewCandleAPI(baseURL string, httpClient *http.Client, logger logger.Logger) *CandleAPI {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &CandleAPI{
		baseURL:    baseURL,
		httpClient: httpClient,
		logger:     logger,
	}
}

// convertPairFormat converte o formato do par de BRLBTC para BTC-BRL
func convertPairFormat(pair string) string {
	// De "BRLBTC" para "BTC-BRL"
	return pair[3:] + "-BRL"
}

// GetCandles obtém os candles para um par em um intervalo de tempo
func (api *CandleAPI) GetCandles(ctx context.Context, pair string, from, to time.Time) ([]model.Candle, error) {
	url := fmt.Sprintf(
		"%s/candles?symbol=%s&from=%d&to=%d&resolution=1d",
		api.baseURL,
		convertPairFormat(pair),
		from.Unix(),
		to.Unix(),
	)

	// Log da URL chamada
	api.logger.Info("Chamando URL da API do Mercado Bitcoin", "url", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		api.logger.Error("Erro ao criar request", err)
		return nil, err
	}

	resp, err := api.httpClient.Do(req)
	if err != nil {
		api.logger.Error("Erro ao fazer request", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status code inválido: %d", resp.StatusCode)
		api.logger.Error("Resposta inválida da API", err)
		return nil, err
	}

	var response apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		api.logger.Error("Erro ao decodificar resposta", err)
		return nil, err
	}

	candles := make([]model.Candle, 0, len(response.T))
	for i := 0; i < len(response.T); i++ {
		open, _ := strconv.ParseFloat(response.O[i], 64)
		close, _ := strconv.ParseFloat(response.C[i], 64)
		high, _ := strconv.ParseFloat(response.H[i], 64)
		low, _ := strconv.ParseFloat(response.L[i], 64)
		volume, _ := strconv.ParseFloat(response.V[i], 64)
		candle := model.Candle{
			Pair:      pair,
			Timestamp: time.Unix(response.T[i], 0),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}
		candles = append(candles, candle)
	}

	candleCount := len(candles)
	api.logger.Info("Quantidade de candles retornados pela API do Mercado Bitcoin", "count", candleCount)

	return candles, nil
}
