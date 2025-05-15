package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"mms_api/internal/application/service"
	"mms_api/internal/domain/model"
	"mms_api/pkg/logger"
)

// MMSResponse representa a resposta da API para consulta de MMS
type MMSResponse struct {
	Timestamp int64   `json:"timestamp"`
	MMS       float64 `json:"mms"`
}

// mmsHandler implementa os handlers HTTP para MMS
type mmsHandler struct {
	mmsService service.MMSService
	logger     logger.Logger
}

// NewMMSHandler cria um novo handler para MMS
func NewMMSHandler(mmsService service.MMSService, logger logger.Logger) *mmsHandler {
	return &mmsHandler{
		mmsService: mmsService,
		logger:     logger,
	}
}

// GetMMSByPair implementa o handler para a rota GET /:pair/mms
func (h *mmsHandler) GetMMSByPair(c *gin.Context) {
	// Extrair o par dos parâmetros da URL
	pair := c.Param("pair")
	if !model.IsValidPair(pair) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Par inválido. Use BRLBTC ou BRLETH"})
		return
	}

	// Extrair parâmetros de consulta
	fromStr := c.Query("from")
	toStr := c.DefaultQuery("to", "")
	rangeStr := c.Query("range")

	// Validar e converter timestamp de início
	fromTs, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'from' inválido"})
		return
	}
	from := time.Unix(fromTs, 0)

	// Validar e converter timestamp de fim (default: dia anterior)
	var to time.Time
	if toStr == "" {
		to = time.Now().AddDate(0, 0, -1)
	} else {
		toTs, err := strconv.ParseInt(toStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'to' inválido"})
			return
		}
		to = time.Unix(toTs, 0)
	}

	// Validar e converter intervalo de dias
	period, err := strconv.Atoi(rangeStr)
	if err != nil || !model.IsValidPeriod(period) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parâmetro 'range' inválido. Use 20, 50 ou 200"})
		return
	}

	// Obter dados do serviço
	result, err := h.mmsService.GetMMSByPairAndRange(c.Request.Context(), pair, from, to, period)
	if err != nil {
		h.logger.Error("erro ao buscar MMS", "error", err, "pair", pair)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar a requisição"})
		return
	}

	// Converter para o formato de resposta
	var response []MMSResponse
	for _, mms := range result {
		var value float64

		// Obter o valor correto com base no período
		switch period {
		case model.Period20:
			value = mms.MMS20
		case model.Period50:
			value = mms.MMS50
		case model.Period200:
			value = mms.MMS200
		}

		response = append(response, MMSResponse{
			Timestamp: mms.Timestamp.Unix(),
			MMS:       value,
		})
	}

	c.JSON(http.StatusOK, response)
}
