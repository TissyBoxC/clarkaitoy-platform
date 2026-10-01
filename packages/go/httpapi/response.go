// Package httpapi defines the shared HTTP response envelope for platform
// services. It intentionally excludes transport frameworks so handlers can use
// the standard library, Gin, or another router without changing the contract.
package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/TissyBoxC/sprout-platform/packages/go/observability"
)

const (
	// SchemaVersion identifies the current response contract.
	SchemaVersion = "1.0.0"
)

// Envelope is the common response shape for successful and failed requests.
type Envelope struct {
	SchemaVersion string     `json:"schema_version"`
	RequestID     string     `json:"request_id"`
	Data          any        `json:"data"`
	Error         *ErrorBody `json:"error"`
}

// ErrorBody contains a stable machine-readable code and user-safe message.
type ErrorBody struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Retryable bool          `json:"retryable"`
	Details   *ErrorDetails `json:"details,omitempty"`
}

// ErrorDetails identifies a correctable field without exposing internals.
type ErrorDetails struct {
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason,omitempty"`
}

// WriteSuccess writes a versioned success envelope.
func WriteSuccess(
	response http.ResponseWriter,
	request *http.Request,
	status int,
	data any,
) {
	writeEnvelope(response, status, Envelope{
		SchemaVersion: SchemaVersion,
		RequestID:     requestID(request),
		Data:          data,
		Error:         nil,
	})
}

// WriteError writes a versioned error envelope and keeps internal errors out of
// the response body.
func WriteError(
	response http.ResponseWriter,
	request *http.Request,
	status int,
	code string,
	message string,
	retryable bool,
) {
	writeEnvelope(response, status, Envelope{
		SchemaVersion: SchemaVersion,
		RequestID:     requestID(request),
		Data:          nil,
		Error: &ErrorBody{
			Code:      code,
			Message:   message,
			Retryable: retryable,
		},
	})
}

func requestID(request *http.Request) string {
	if request == nil {
		return ""
	}
	return observability.MetadataFromContext(request.Context()).RequestID
}

func writeEnvelope(response http.ResponseWriter, status int, envelope Envelope) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(envelope)
}
