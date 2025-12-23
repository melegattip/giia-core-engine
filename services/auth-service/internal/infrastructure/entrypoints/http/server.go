package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	pkgLogger "github.com/giia/giia-core-engine/pkg/logger"
)

type Server struct {
	httpServer *http.Server
	logger     pkgLogger.Logger
}

func NewServer(port string, router *gin.Engine, logger pkgLogger.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:           fmt.Sprintf(":%s", port),
			Handler:        router,
			ReadTimeout:    15 * time.Second,
			WriteTimeout:   15 * time.Second,
			IdleTimeout:    60 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info(context.Background(), fmt.Sprintf("HTTP Server starting on %s", s.httpServer.Addr), nil)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error(context.Background(), err, "HTTP Server failed to start", nil)
		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info(ctx, "HTTP Server shutting down", nil)

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error(ctx, err, "HTTP Server shutdown failed", nil)
		return err
	}

	s.logger.Info(ctx, "HTTP Server stopped gracefully", nil)
	return nil
}

func (s *Server) GetAddr() string {
	return s.httpServer.Addr
}
