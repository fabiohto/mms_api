package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mms_api/pkg/logger"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	logger     logger.Logger
}

func NewServer(router *gin.Engine, port string, logger logger.Logger) *Server {
	if port == "" {
		port = "8080"
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("0.0.0.0:%s", port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger,
	}
}

// Start initializes the server and handles graceful shutdown
func (s *Server) Start() error {
	// Start server in a goroutine
	go func() {
		s.logger.Info("Starting server on", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Error starting server:", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown:", err)
		return err
	}

	s.logger.Info("Server gracefully stopped")
	return nil
}
