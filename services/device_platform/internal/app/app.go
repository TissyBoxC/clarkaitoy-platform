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
	aiProvider "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/provider"
	aiRepository "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/repository"
	aiService "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/service"
	authRepository "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/repository"
	authService "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/service"
	bindingRepository "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/repository"
	bindingService "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/service"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/cache"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/database"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/platform/security"
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

	tokenIssuer, err := security.NewHMACTokenIssuer(cfg.Auth.AccessTokenSecret)
	if err != nil {
		return fmt.Errorf("create token issuer: %w", err)
	}
	credentialCipher, err := aiService.NewAESGCMCipher(cfg.Auth.CredentialKey)
	if err != nil {
		return fmt.Errorf("create AI credential cipher: %w", err)
	}
	mfaCipher, err := authService.NewAESGCMTOTPCipher(cfg.Auth.MFACredentialKey)
	if err != nil {
		return fmt.Errorf("create MFA credential cipher: %w", err)
	}
	aiClient, err := aiProvider.New(cfg.AI.BaseURL, cfg.AI.ServiceToken, nil)
	if err != nil {
		return fmt.Errorf("create AI provider client: %w", err)
	}
	aiAccountService, err := aiService.New(aiService.Options{
		Repository:         aiRepository.NewPostgresRepository(databaseStore.Pool()),
		Provider:           aiClient,
		Cipher:             credentialCipher,
		DefaultBalanceUSD:  cfg.AI.DefaultBalanceUSD,
		DefaultModels:      cfg.AI.DefaultModels,
		DefaultConcurrency: cfg.AI.DefaultConcurrency,
	})
	if err != nil {
		return fmt.Errorf("create AI account service: %w", err)
	}
	var phoneVerifier authService.PhoneVerifier
	if cfg.Auth.PhoneVerificationMode == "local" {
		phoneVerifier = authService.NewLocalPhoneVerifier(
			authRepository.NewPostgresRepository(databaseStore.Pool()),
			nil,
		)
	}
	parentAuthService, err := authService.New(authService.Options{
		Repository:      authRepository.NewPostgresRepository(databaseStore.Pool()),
		TokenIssuer:     tokenIssuer,
		AIProvisioner:   aiAccountService,
		MFACipher:       mfaCipher,
		PhoneVerifier:   phoneVerifier,
		MFAChallengeTTL: cfg.Auth.MFAChallengeTTL,
		AccessTTL:       cfg.Auth.AccessTokenTTL,
		RefreshTTL:      cfg.Auth.RefreshTokenTTL,
	})
	if err != nil {
		return fmt.Errorf("create authentication service: %w", err)
	}
	deviceBindingService, err := bindingService.New(bindingService.Options{
		Repository:    bindingRepository.NewPostgresRepository(databaseStore.Pool()),
		TokenTTL:      15 * time.Minute,
		ProofVerifier: security.ECDSAProofVerifier{},
	})
	if err != nil {
		return fmt.Errorf("create device binding service: %w", err)
	}

	server := &http.Server{
		Addr: cfg.HTTP.Address(),
		Handler: platformhttp.NewRouter(platformhttp.RouterOptions{
			Logger:            logger,
			InternalAPIConfig: cfg.Internal,
			AuthService:       parentAuthService,
			AIService:         aiAccountService,
			BindingService:    deviceBindingService,
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
