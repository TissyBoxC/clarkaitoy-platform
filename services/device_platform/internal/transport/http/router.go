// Package http exposes device platform health and internal management endpoints.
package http

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/TissyBoxC/sprout-platform/packages/go/httpapi"
	"github.com/TissyBoxC/sprout-platform/packages/go/observability"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/config"
	gatewayservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/ai_gateway/service"
	authservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/auth/service"
	bindingservice "github.com/TissyBoxC/sprout-platform/services/device_platform/internal/modules/device_binding/service"
)

// RouterOptions contains dependencies for the device platform HTTP transport.
type RouterOptions struct {
	Logger            *slog.Logger
	InternalAPIConfig config.InternalAPIConfig
	AuthService       *authservice.Service
	AIService         *gatewayservice.Service
	BindingService    *bindingservice.Service
}

// NewRouter returns the HTTP router for the device platform.
func NewRouter(options RouterOptions) http.Handler {
	logger := options.Logger
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", readyHandler)

	if options.AuthService != nil {
		authHandler := authHandler{service: options.AuthService}
		mux.HandleFunc("POST /api/v1/auth/register", authHandler.register)
		mux.HandleFunc("POST /api/v1/auth/login", authHandler.login)
		mux.HandleFunc("POST /api/v1/admin/auth/login", authHandler.startAdminLogin)
		mux.HandleFunc("POST /api/v1/admin/auth/mfa", authHandler.completeAdminLogin)
		mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.refresh)
		mux.HandleFunc("POST /api/v1/auth/logout", authHandler.logout)
		mux.HandleFunc(
			"GET /api/v1/auth/me",
			authHandler.requireAuthentication(authHandler.me),
		)
		if options.BindingService != nil {
			bindingHandler := deviceBindingHandler{service: options.BindingService}
			mux.HandleFunc(
				"POST /api/v1/devices/binding-tokens",
				authHandler.requireAuthentication(bindingHandler.createToken),
			)
			mux.HandleFunc(
				"POST /api/v1/devices/bind",
				authHandler.requireAuthentication(bindingHandler.bind),
			)
			mux.HandleFunc(
				"GET /api/v1/devices",
				authHandler.requireAuthentication(bindingHandler.list),
			)
			mux.HandleFunc(
				"DELETE /api/v1/devices/{device_id}",
				authHandler.requireAuthentication(bindingHandler.remove),
			)
		}
		if options.AIService != nil {
			adminHandler := adminHandler{service: options.AIService}
			mux.HandleFunc(
				"GET /api/v1/admin/ai-accounts",
				authHandler.requireAdmin(adminHandler.listAIAccounts),
			)
			mux.HandleFunc(
				"PUT /api/v1/admin/ai-accounts/{provider_account_id}",
				authHandler.requireAdmin(adminHandler.updateAIAccount),
			)
		}
	}

	if options.InternalAPIConfig.Enabled {
		mux.Handle(
			"GET /internal/v1/runtime",
			requireServiceToken(
				options.InternalAPIConfig.AuthToken,
				http.HandlerFunc(runtimeHandler),
			),
		)
	}

	return observability.WithRequestLabels(observability.WithRequestMetadata(
		observability.WithAccessLog(
			mux,
			observability.AccessLogOptions{
				Logger:      logger,
				ServiceName: "device-platform",
				Audit:       observability.NewSlogAuditSink(logger),
			},
		),
	))
}

func healthHandler(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "device-platform",
		"time":    time.Now().UTC(),
	})
}

func readyHandler(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, map[string]any{
		"status": "ready",
	})
}

func runtimeHandler(response http.ResponseWriter, request *http.Request) {
	httpapi.WriteSuccess(response, request, http.StatusOK, map[string]any{
		"service":          "device-platform",
		"status":           "running",
		"protocol_version": httpapi.SchemaVersion,
	})
}

func requireServiceToken(expectedToken string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		providedToken := request.Header.Get("Authorization")
		const bearerPrefix = "Bearer "
		if len(providedToken) <= len(bearerPrefix) ||
			providedToken[:len(bearerPrefix)] != bearerPrefix {
			httpapi.WriteError(
				response,
				request,
				http.StatusUnauthorized,
				"unauthenticated",
				"服务认证失败",
				false,
			)
			return
		}

		providedToken = providedToken[len(bearerPrefix):]
		if subtle.ConstantTimeCompare([]byte(providedToken), []byte(expectedToken)) != 1 {
			httpapi.WriteError(
				response,
				request,
				http.StatusUnauthorized,
				"unauthenticated",
				"服务认证失败",
				false,
			)
			return
		}

		next.ServeHTTP(response, request)
	})
}

func writeJSON(response http.ResponseWriter, status int, payload any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(payload)
}
