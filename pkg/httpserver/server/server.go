package httpserver

import (
	"context"
	"errors"
	"fmt"
	"go_home/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
}

type HTTPServer struct {
	router *gin.Engine
	config Config
	logger *logger.Logger
}

func NewHTTPServer(config Config, logger *logger.Logger) *HTTPServer {
	router := gin.New()

	return &HTTPServer{
		router: router,
		config: config,
		logger: logger,
	}
}

func (h *HTTPServer) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:              h.config.Addr,
		Handler:           h.router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)

	h.logger.Info("starting HTTP server", zap.String("addr", h.config.Addr))

	go func() {
		defer close(errCh)

		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			h.logger.Error("HTTP server failed", zap.Error(err))

			return fmt.Errorf("listen and serve HTTP: %w", err)
		}
	case <-ctx.Done():
		h.logger.Info("starting HTTP server graceful shutdown", zap.Duration("timeout", h.config.ShutdownTimeout))

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			h.config.ShutdownTimeout,
		)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			closeErr := server.Close()

			h.logger.Error("HTTP server forced shutdown failed", zap.Error(closeErr))

			return errors.Join(
				fmt.Errorf("shutdown HTTP server: %w", err),
				closeErr,
			)
		}

		h.logger.Info("HTTP server stopped gracefully")
	}

	return nil
}

func (h *HTTPServer) Router() *gin.Engine {
	return h.router
}
