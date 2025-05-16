package http

import (
	"mms_api/internal/application/port/in"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title MMS API
// @version 1.0
// @description API para cálculo e consulta de Médias Móveis Simples de criptomoedas
// @host localhost:8080
// @BasePath /api/v1
type Router struct {
	mmsHandler in.MMSHandler
}

func NewRouter(mmsHandler in.MMSHandler) *Router {
	return &Router{
		mmsHandler: mmsHandler,
	}
}

// SetupRoutes configures all the routes for the API using Gin framework
func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	router.GET("/health", r.handleHealth())

	// API v1 routes group
	v1 := router.Group("/api/v1")
	{
		mms := v1.Group("/mms")
		{
			mms.GET("", r.mmsHandler.GetMMSByPair) // Get MMS by pair and timeframe
		}
	}

	return router
}

// handleHealth returns the health check handler
func (r *Router) handleHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	}
}
