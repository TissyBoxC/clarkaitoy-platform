// Package app wires and runs the voice gateway.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/TissyBoxC/sprout-platform/packages/go/observability"
	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/config"
	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/platform/cache"
	gatewayhttp "github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/transport/http"
	"github.com/TissyBoxC/sprout-platform/services/voice_gateway/internal/usage"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Run starts the voice gateway and waits for a shutdown signal.
func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(observability.NewRedactingHandler(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: cfg.Log.SlogLevel(),
		},
	)))
	slog.SetDefault(logger)

	startupCtx, startupCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer startupCancel()

	redisCache, err := cache.Open(startupCtx, cfg.Redis.Address, cfg.Redis.Password)
	if err != nil {
		return fmt.Errorf("open redis: %w", err)
	}
	defer redisCache.Close()

	var usageRecorder usage.Recorder
	if cfg.Database.DSN != "" {
		databaseStore, err := pgxpool.New(startupCtx, cfg.Database.DSN)
		if err != nil {
			return fmt.Errorf("open usage database: %w", err)
		}
		defer databaseStore.Close()
		usageRecorder = usage.NewPostgresRecorder(databaseStore)
	}

	server := &http.Server{
		Addr: cfg.HTTP.Address(),
		Handler: gatewayhttp.NewRouter(gatewayhttp.RouterOptions{
			Logger:            logger,
			InternalAPIConfig: cfg.Internal,
			UsageRecorder:     usageRecorder,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("voice gateway started", "address", server.Addr)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("serve: %w", err)
	case sig := <-stopCh:
		logger.Info("shutdown requested", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
