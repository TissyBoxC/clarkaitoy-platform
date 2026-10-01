package http

import (
	"bytes"
	"encoding/json"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/TissyBoxC/sprout-platform/packages/go/httpapi"
	"github.com/TissyBoxC/sprout-platform/services/device_platform/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	request := httptest.NewRequest(stdhttp.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	NewRouter(newTestRouterOptions()).ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusOK {
		t.Fatalf("expected status %d, got %d", stdhttp.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
}

func TestHealthEndpointReturnsTraceParent(t *testing.T) {
	request := httptest.NewRequest(stdhttp.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	NewRouter(newTestRouterOptions()).ServeHTTP(recorder, request)

	if recorder.Header().Get("traceparent") == "" {
		t.Fatal("expected traceparent response header")
	}
}

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil))
}

func TestInternalAPIIsHiddenWhenDisabled(t *testing.T) {
	request := httptest.NewRequest(stdhttp.MethodGet, "/internal/v1/runtime", nil)
	recorder := httptest.NewRecorder()

	NewRouter(newTestRouterOptions()).ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusNotFound {
		t.Fatalf("expected status %d, got %d", stdhttp.StatusNotFound, recorder.Code)
	}
}

func TestInternalAPIRejectsMissingServiceToken(t *testing.T) {
	options := newTestRouterOptions()
	options.InternalAPIConfig = config.InternalAPIConfig{
		Enabled:   true,
		AuthToken: strings.Repeat("t", 32),
	}

	request := httptest.NewRequest(stdhttp.MethodGet, "/internal/v1/runtime", nil)
	recorder := httptest.NewRecorder()
	NewRouter(options).ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", stdhttp.StatusUnauthorized, recorder.Code)
	}
	envelope := decodeEnvelope(t, recorder)
	if envelope.Error == nil || envelope.Error.Code != "unauthenticated" {
		t.Fatalf("unexpected error envelope: %+v", envelope)
	}
}

func TestInternalAPIReturnsRuntimeEnvelope(t *testing.T) {
	options := newTestRouterOptions()
	options.InternalAPIConfig = config.InternalAPIConfig{
		Enabled:   true,
		AuthToken: strings.Repeat("t", 32),
	}

	request := httptest.NewRequest(stdhttp.MethodGet, "/internal/v1/runtime", nil)
	request.Header.Set("Authorization", "Bearer "+strings.Repeat("t", 32))
	recorder := httptest.NewRecorder()
	NewRouter(options).ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusOK {
		t.Fatalf("expected status %d, got %d", stdhttp.StatusOK, recorder.Code)
	}
	envelope := decodeEnvelope(t, recorder)
	if envelope.SchemaVersion != httpapi.SchemaVersion {
		t.Fatalf("unexpected schema version: %q", envelope.SchemaVersion)
	}
	if envelope.RequestID == "" {
		t.Fatal("expected response request ID")
	}
}

func TestInternalAPIContractDocumentsRuntimeEndpoint(t *testing.T) {
	contract, err := os.ReadFile("../../contracts/http/openapi.yaml")
	if err != nil {
		t.Fatalf("read internal API contract: %v", err)
	}
	if !bytes.Contains(contract, []byte("/internal/v1/runtime")) {
		t.Fatal("expected runtime endpoint in internal API contract")
	}
}

type testEnvelope struct {
	SchemaVersion string             `json:"schema_version"`
	RequestID     string             `json:"request_id"`
	Error         *httpapi.ErrorBody `json:"error"`
}

func decodeEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) testEnvelope {
	t.Helper()

	var envelope testEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response envelope: %v", err)
	}
	return envelope
}

func newTestRouterOptions() RouterOptions {
	return RouterOptions{Logger: newTestLogger()}
}
