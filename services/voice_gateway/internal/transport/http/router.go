// Package http exposes voice gateway health and management endpoints.
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/TissyBoxC/sprout-platform/packages/go/observability"
)

// NewRouter returns the initial HTTP router for the voice gateway.
func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /readyz", readyHandler)

	return observability.WithRequestMetadata(observability.WithAccessLog(
		mux,
		observability.AccessLogOptions{
			Logger:      logger,
			ServiceName: "voice-gateway",
			Audit:       observability.NewSlogAuditSink(logger),
		},
	))
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "voice-gateway",
		"time":    time.Now().UTC(),
	})
}

func readyHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ready",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
