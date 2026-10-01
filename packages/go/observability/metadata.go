// Package observability provides request correlation, structured access logs,
// audit events, and redaction shared by platform services.
package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// RequestIDHeader is the public correlation ID used by clients and logs.
	RequestIDHeader = "X-Request-ID"
	// TraceParentHeader is the W3C distributed tracing header.
	TraceParentHeader = "traceparent"
)

// Metadata contains the correlation identifiers attached to one request.
type Metadata struct {
	RequestID string
	TraceID   string
	SpanID    string
}

type metadataContextKey struct{}

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

// WithRequestMetadata assigns request and trace identifiers, returns them in
// response headers, and stores them in the request context for downstream code.
func WithRequestMetadata(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		metadata := newMetadata(request)

		response.Header().Set(RequestIDHeader, metadata.RequestID)
		response.Header().Set(
			TraceParentHeader,
			traceparent(metadata.TraceID, metadata.SpanID),
		)

		contextWithMetadata := context.WithValue(
			request.Context(),
			metadataContextKey{},
			metadata,
		)
		next.ServeHTTP(response, request.WithContext(contextWithMetadata))
	})
}

// MetadataFromContext returns correlation metadata attached by
// WithRequestMetadata. Missing metadata returns the zero value.
func MetadataFromContext(ctx context.Context) Metadata {
	metadata, isMetadata := ctx.Value(metadataContextKey{}).(Metadata)
	if !isMetadata {
		return Metadata{}
	}
	return metadata
}

func newMetadata(request *http.Request) Metadata {
	requestID := strings.TrimSpace(request.Header.Get(RequestIDHeader))
	if !requestIDPattern.MatchString(requestID) {
		requestID = uuid.NewString()
	}

	traceID, parentSpanID := extractTraceContext(request)
	if traceID == "" {
		traceID = randomHex(16)
	}
	if parentSpanID == "" {
		parentSpanID = randomHex(8)
	}

	return Metadata{
		RequestID: requestID,
		TraceID:   traceID,
		SpanID:    randomHex(8),
	}
}

func extractTraceContext(request *http.Request) (string, string) {
	carrier := propagation.HeaderCarrier(request.Header)
	extractedContext := propagation.TraceContext{}.Extract(
		request.Context(),
		carrier,
	)
	spanContext := trace.SpanContextFromContext(extractedContext)
	if !spanContext.IsValid() {
		return "", ""
	}
	return spanContext.TraceID().String(), spanContext.SpanID().String()
}

func randomHex(byteCount int) string {
	value := make([]byte, byteCount)
	if _, err := rand.Read(value); err != nil {
		// crypto/rand failure means the process cannot safely continue issuing
		// correlations; a panic is preferable to silently reusing identifiers.
		panic("observability: random source unavailable")
	}
	return hex.EncodeToString(value)
}

func traceparent(traceID string, spanID string) string {
	return "00-" + traceID + "-" + spanID + "-00"
}
