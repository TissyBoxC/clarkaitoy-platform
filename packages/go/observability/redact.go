package observability

import (
	"regexp"
	"strings"
)

const redactedValue = "[redacted]"

var sensitiveFields = map[string]struct{}{
	"access_token":  {},
	"api_key":       {},
	"authorization": {},
	"audio":         {},
	"body":          {},
	"child_name":    {},
	"conversation":  {},
	"cookie":        {},
	"email":         {},
	"image":         {},
	"password":      {},
	"prompt":        {},
	"refresh_token": {},
	"secret":        {},
	"token":         {},
}

var (
	bearerPattern = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+`)
	emailPattern  = regexp.MustCompile(`(?i)\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b`)
)

// SanitizeFields removes sensitive values from a slog-style key/value slice.
// Unknown fields are preserved, while free-form values are masked when they
// resemble credentials or personal identifiers.
func SanitizeFields(fields []any) []any {
	sanitized := make([]any, len(fields))
	copy(sanitized, fields)

	for index := 0; index+1 < len(sanitized); index += 2 {
		key, isKey := sanitized[index].(string)
		if !isKey {
			continue
		}
		if isSensitiveField(key) {
			sanitized[index+1] = redactedValue
			continue
		}
		sanitized[index+1] = RedactValue(sanitized[index+1])
	}
	return sanitized
}

// RedactValue masks credentials and email-like values in a scalar field.
func RedactValue(value any) any {
	text, isText := value.(string)
	if !isText {
		return value
	}
	text = bearerPattern.ReplaceAllString(text, redactedValue)
	text = emailPattern.ReplaceAllString(text, redactedValue)
	return text
}

func isSensitiveField(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if _, isSensitive := sensitiveFields[normalized]; isSensitive {
		return true
	}
	return strings.HasSuffix(normalized, "_token") ||
		strings.HasSuffix(normalized, "_secret") ||
		strings.HasSuffix(normalized, "_password")
}
