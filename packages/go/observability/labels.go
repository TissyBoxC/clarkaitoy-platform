package observability

import (
	"context"
	"net/http"
	"strings"
)

const (
	// TenantIDHeader is the pseudonymous tenant identifier supplied by callers.
	TenantIDHeader = "X-Sprout-Tenant-ID"
	// DeviceIDHeader is the pseudonymous device identifier supplied by callers.
	DeviceIDHeader = "X-Sprout-Device-ID"
	// RequestPurposeHeader describes why the service is processing the request.
	RequestPurposeHeader = "X-Sprout-Request-Purpose"
	// PolicyVersionHeader identifies the applied child-safety policy version.
	PolicyVersionHeader = "X-Sprout-Policy-Version"
	// SessionIDHeader identifies the device session without identifying a child.
	SessionIDHeader = "X-Sprout-Session-ID"
)

const (
	// TenantIDField is the canonical tenant log field.
	TenantIDField = "tenant_id"
	// DeviceIDField is the canonical device log field.
	DeviceIDField = "device_id"
	// RequestPurposeField is the canonical processing-purpose log field.
	RequestPurposeField = "request_purpose"
	// PolicyVersionField is the canonical policy-version log field.
	PolicyVersionField = "policy_version"
	// SessionIDField is the canonical session log field.
	SessionIDField = "session_id"
)

// RequestLabels contains pseudonymous routing and policy context for one request.
//
// Labels must never contain names, contact details, precise locations, free-form
// prompts, or credentials. Invalid values are discarded before logging.
type RequestLabels struct {
	TenantID       string
	DeviceID       string
	RequestPurpose string
	PolicyVersion  string
	SessionID      string
}

type labelsContextKey struct{}

// WithRequestLabels validates supported labels and stores them in the request
// context. Unknown headers remain untouched and unsupported fields are dropped.
func WithRequestLabels(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		labels := RequestLabels{
			TenantID:       validLabel(request.Header.Get(TenantIDHeader)),
			DeviceID:       validLabel(request.Header.Get(DeviceIDHeader)),
			RequestPurpose: validLabel(request.Header.Get(RequestPurposeHeader)),
			PolicyVersion:  validLabel(request.Header.Get(PolicyVersionHeader)),
			SessionID:      validLabel(request.Header.Get(SessionIDHeader)),
		}

		contextWithLabels := context.WithValue(
			request.Context(),
			labelsContextKey{},
			labels,
		)
		next.ServeHTTP(response, request.WithContext(contextWithLabels))
	})
}

// RequestLabelsFromContext returns the validated labels attached by
// WithRequestLabels. Missing labels return the zero value.
func RequestLabelsFromContext(ctx context.Context) RequestLabels {
	labels, hasLabels := ctx.Value(labelsContextKey{}).(RequestLabels)
	if !hasLabels {
		return RequestLabels{}
	}
	return labels
}

func validLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 128 || !labelPattern.MatchString(value) {
		return ""
	}
	return value
}
