package http

import (
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	request := httptest.NewRequest(stdhttp.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	NewRouter().ServeHTTP(recorder, request)

	if recorder.Code != stdhttp.StatusOK {
		t.Fatalf("expected status %d, got %d", stdhttp.StatusOK, recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID response header")
	}
}
