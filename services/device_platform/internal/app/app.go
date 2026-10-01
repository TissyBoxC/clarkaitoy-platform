// Package app wires and runs the device platform service.
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
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/config"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/cache"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/database"
	platformhttp "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/transport/http"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/transport/mqtt"
)

// Run starts the HTTP server and waits for a shutdown signal.
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

	databaseStore, err := database.Open(startupCtx, cfg.Database.DSN)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer databaseStore.Close(context.Background())

	redisCache, err := cache.Open(startupCtx, cfg.Redis.Address, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return fmt.Errorf("open redis: %w", err)
	}
	defer redisCache.Close()

	mqttClient, err := mqtt.Open(startupCtx, cfg.MQTT)
	if err != nil {
		return fmt.Errorf("open MQTT: %w", err)
	}
	defer mqttClient.Close()

	server := &http.Server{
		Addr: cfg.HTTP.Address(),
		Handler: platformhttp.NewRouter(platformhttp.RouterOptions{
			Logger:            logger,
			InternalAPIConfig: cfg.Internal,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("device platform started", "address", server.Addr)
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
