module github.com/TissyBoxC/sprout-platform/packages/go/httpapi

go 1.27.1

require github.com/TissyBoxC/sprout-platform/packages/go/observability v0.0.0

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel v1.40.0 // indirect
	go.opentelemetry.io/otel/trace v1.40.0 // indirect
)

replace github.com/TissyBoxC/sprout-platform/packages/go/observability => ../observability
