package in

import (
	"github.com/gin-gonic/gin" // Assumindo Gin como framework web
)

// MMSHandler define o contrato para handlers HTTP relacionados a MMS
type MMSHandler interface {
	// Obter MMSs para um par específico em um intervalo de tempo
	GetMMSByPair(c *gin.Context)
}
