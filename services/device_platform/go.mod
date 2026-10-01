module github.com/TissyBoxC/sprout-platform/services/device_platform

go 1.27.1

require (
	github.com/TissyBoxC/sprout-platform/packages/go/httpapi v0.0.0
	github.com/TissyBoxC/sprout-platform/packages/go/observability v0.0.0
	github.com/eclipse/paho.mqtt.golang v1.5.1
	github.com/jackc/pgx/v5 v5.7.4
	github.com/redis/go-redis/v9 v9.22.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	go.opentelemetry.io/otel v1.40.0 // indirect
	go.opentelemetry.io/otel/trace v1.40.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace github.com/TissyBoxC/sprout-platform/packages/go/observability => ../../packages/go/observability

replace github.com/TissyBoxC/sprout-platform/packages/go/httpapi => ../../packages/go/httpapi
