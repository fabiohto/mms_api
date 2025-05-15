package http

import (
	"mms_api/internal/application/port/in"

	"github.com/gin-gonic/gin"
)

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
