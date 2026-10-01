module github.com/TissyBoxC/sprout-platform/services/voice_gateway

go 1.27.1

require (
	github.com/TissyBoxC/sprout-platform/packages/go/httpapi v0.0.0
	github.com/TissyBoxC/sprout-platform/packages/go/observability v0.0.0
	github.com/redis/go-redis/v9 v9.22.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel v1.40.0 // indirect
	go.opentelemetry.io/otel/trace v1.40.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace github.com/TissyBoxC/sprout-platform/packages/go/observability => ../../packages/go/observability

replace github.com/TissyBoxC/sprout-platform/packages/go/httpapi => ../../packages/go/httpapi
