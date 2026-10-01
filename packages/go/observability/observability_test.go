package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRequestMetadataGeneratesAndPropagatesIdentifiers(t *testing.T) {
	handler := WithRequestMetadata(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		metadata := MetadataFromContext(request.Context())
		if metadata.RequestID == "" || metadata.TraceID == "" || metadata.SpanID == "" {
			t.Fatal("expected metadata in request context")
		}
		response.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if recorder.Header().Get(RequestIDHeader) == "" {
		t.Fatal("expected request ID response header")
	}
	if recorder.Header().Get(TraceParentHeader) == "" {
		t.Fatal("expected traceparent response header")
	}
}

func TestWithRequestMetadataPreservesValidRequestID(t *testing.T) {
	handler := WithRequestMetadata(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		response.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set(RequestIDHeader, "request-123")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Header().Get(RequestIDHeader) != "request-123" {
		t.Fatal("expected existing request ID to be preserved")
	}
}

func TestWithAccessLogRedactsSensitiveFields(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&output, nil)))
	auditSink := &capturingAuditSink{}

	handler := WithAccessLog(
		http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			logger.Info("test", "api_key", "secret-value", "email", "guardian@example.com")
			response.WriteHeader(http.StatusOK)
		}),
		AccessLogOptions{
			Logger:      logger,
			ServiceName: "test-service",
			Audit:       auditSink,
		},
	)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/test", nil)
	handler = WithRequestMetadata(handler)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	var records []map[string]any
	for _, line := range bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n")) {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("decode log record: %v", err)
		}
		records = append(records, record)
	}
	if len(records) != 2 {
		t.Fatalf("expected application and access records, got %d", len(records))
	}
	if records[0]["api_key"] != redactedValue {
		t.Fatalf("expected api key to be redacted, got %v", records[0]["api_key"])
	}
	if records[0]["email"] != redactedValue {
		t.Fatalf("expected email to be redacted, got %v", records[0]["email"])
	}
	if auditSink.event.Method != http.MethodPost {
		t.Fatalf("expected POST audit event, got %+v", auditSink.event)
	}
	if auditSink.event.RequestID == "" || auditSink.event.TraceID == "" {
		t.Fatalf("expected audit correlation ids, got %+v", auditSink.event)
	}
}

func TestMetadataFromContextWithoutMetadata(t *testing.T) {
	if metadata := MetadataFromContext(context.Background()); metadata != (Metadata{}) {
		t.Fatalf("expected zero metadata, got %+v", metadata)
	}
}

func TestWithRequestLabelsFiltersUnsupportedValues(t *testing.T) {
	handler := WithRequestLabels(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		labels := RequestLabelsFromContext(request.Context())
		if labels.TenantID != "tenant_001" {
			t.Fatalf("expected tenant label, got %q", labels.TenantID)
		}
		if labels.DeviceID != "" {
			t.Fatalf("expected unsafe device label to be dropped, got %q", labels.DeviceID)
		}
		response.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/runtime", nil)
	request.Header.Set(TenantIDHeader, "tenant_001")
	request.Header.Set(DeviceIDHeader, "device id with spaces")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestWithAccessLogIncludesValidatedLabels(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(NewRedactingHandler(slog.NewJSONHandler(&output, nil)))

	handler := WithRequestLabels(WithAccessLog(
		http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusOK)
		}),
		AccessLogOptions{
			Logger:      logger,
			ServiceName: "test-service",
		},
	))

	request := httptest.NewRequest(http.MethodGet, "/internal/v1/runtime", nil)
	request.Header.Set(TenantIDHeader, "tenant_001")
	request.Header.Set(DeviceIDHeader, "device_001")
	request.Header.Set(RequestPurposeHeader, "voice_conversation")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	var record map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &record); err != nil {
		t.Fatalf("decode log record: %v", err)
	}
	if record[TenantIDField] != "tenant_001" {
		t.Fatalf("expected tenant label in access log, got %v", record[TenantIDField])
	}
	if record[DeviceIDField] != "device_001" {
		t.Fatalf("expected device label in access log, got %v", record[DeviceIDField])
	}
	if record[RequestPurposeField] != "voice_conversation" {
		t.Fatalf("expected purpose in access log, got %v", record[RequestPurposeField])
	}
}

type capturingAuditSink struct {
	event AuditEvent
}

func (sink *capturingAuditSink) RecordAudit(event AuditEvent) {
	sink.event = event
}
