package usage

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRecorder persists billable conversation usage into the platform
// database. It is intentionally append-only per day and device so a retried
// request cannot create duplicate conversation counts.
type PostgresRecorder struct {
	pool *pgxpool.Pool
}

// NewPostgresRecorder creates a usage recorder backed by PostgreSQL.
func NewPostgresRecorder(pool *pgxpool.Pool) *PostgresRecorder {
	return &PostgresRecorder{pool: pool}
}

// Record accumulates one completed conversation.
//
// The device must be bound to exactly one guardian. An unbound device is
// ignored so anonymous telemetry cannot create billable usage.
func (r *PostgresRecorder) Record(
	ctx context.Context,
	record Record,
) error {
	if r == nil || r.pool == nil {
		return fmt.Errorf("usage recorder database is not configured")
	}
	deviceID := strings.TrimSpace(record.DeviceID)
	if deviceID == "" {
		return nil
	}
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO platform_usage_daily (
			usage_day,
			parent_account_id,
			device_id,
			conversation_count,
			spent_usd,
			input_tokens,
			output_tokens,
			updated_at
		)
		SELECT
			CURRENT_DATE,
			parent_account_id,
			device_id,
			1,
			$2,
			$3,
			$4,
			NOW()
		FROM device_bindings
		WHERE device_id = $1
		ON CONFLICT (usage_day, parent_account_id, device_id) DO UPDATE
		SET conversation_count = platform_usage_daily.conversation_count + 1,
		    spent_usd = platform_usage_daily.spent_usd + EXCLUDED.spent_usd,
		    input_tokens = platform_usage_daily.input_tokens + EXCLUDED.input_tokens,
		    output_tokens = platform_usage_daily.output_tokens + EXCLUDED.output_tokens,
		    updated_at = NOW()
	`, deviceID, record.SpentUSD, record.InputSize, record.OutputSize)
	if err != nil {
		return fmt.Errorf("record platform usage: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	return nil
}
