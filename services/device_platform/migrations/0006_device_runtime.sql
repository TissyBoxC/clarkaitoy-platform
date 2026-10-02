CREATE TABLE IF NOT EXISTS device_runtime_status (
    device_id TEXT PRIMARY KEY REFERENCES device_credentials(device_id) ON DELETE CASCADE,
    last_heartbeat_id TEXT NOT NULL,
    connection_state TEXT NOT NULL,
    transport TEXT NOT NULL,
    network_quality_level TEXT NOT NULL,
    rssi_dbm INTEGER NOT NULL,
    latency_ms INTEGER NOT NULL,
    packet_loss_percent INTEGER NOT NULL,
    time_sync_state TEXT NOT NULL,
    time_sync_source TEXT NOT NULL,
    time_synced_at TIMESTAMPTZ,
    time_offset_ms INTEGER NOT NULL,
    offline_state TEXT NOT NULL,
    offline_reason TEXT NOT NULL,
    fallback_active BOOLEAN NOT NULL DEFAULT FALSE,
    pending_telemetry INTEGER NOT NULL DEFAULT 0,
    firmware_version TEXT NOT NULL DEFAULT '',
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT device_runtime_connection_state_valid
        CHECK (connection_state IN ('offline', 'connecting', 'online')),
    CONSTRAINT device_runtime_transport_valid
        CHECK (transport IN ('none', 'wifi', 'cellular_4g')),
    CONSTRAINT device_runtime_quality_valid
        CHECK (network_quality_level IN ('unknown', 'poor', 'fair', 'good', 'excellent')),
    CONSTRAINT device_runtime_quality_metrics_valid
        CHECK (
            rssi_dbm BETWEEN -127 AND 0
            AND latency_ms BETWEEN 0 AND 60000
            AND packet_loss_percent BETWEEN 0 AND 100
        ),
    CONSTRAINT device_runtime_time_sync_valid
        CHECK (time_sync_state IN ('unsynchronized', 'synchronizing', 'synchronized')),
    CONSTRAINT device_runtime_time_source_valid
        CHECK (time_sync_source IN ('none', 'sntp', 'platform')),
    CONSTRAINT device_runtime_time_offset_valid
        CHECK (time_offset_ms BETWEEN -60000 AND 60000),
    CONSTRAINT device_runtime_offline_state_valid
        CHECK (offline_state IN ('online', 'grace', 'offline')),
    CONSTRAINT device_runtime_offline_reason_valid
        CHECK (
            offline_reason IN (
                'none',
                'network_unavailable',
                'authentication_failed',
                'time_not_synchronized',
                'service_unavailable'
            )
        ),
    CONSTRAINT device_runtime_pending_telemetry_valid
        CHECK (pending_telemetry BETWEEN 0 AND 10000)
);

CREATE INDEX IF NOT EXISTS device_runtime_status_received_at_idx
    ON device_runtime_status(received_at DESC);

CREATE INDEX IF NOT EXISTS device_runtime_status_quality_idx
    ON device_runtime_status(network_quality_level, received_at DESC);

CREATE TABLE IF NOT EXISTS device_runtime_heartbeats (
    device_id TEXT NOT NULL REFERENCES device_credentials(device_id) ON DELETE CASCADE,
    heartbeat_id TEXT NOT NULL,
    reported_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (device_id, heartbeat_id)
);

CREATE INDEX IF NOT EXISTS device_runtime_heartbeats_received_at_idx
    ON device_runtime_heartbeats(received_at DESC);

CREATE TABLE IF NOT EXISTS device_runtime_commands (
    id UUID PRIMARY KEY,
    device_id TEXT NOT NULL REFERENCES device_credentials(device_id) ON DELETE CASCADE,
    command_type TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    status TEXT NOT NULL DEFAULT 'pending',
    requested_by TEXT NOT NULL DEFAULT '',
    request_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ,
    acknowledged_at TIMESTAMPTZ,
    result_code TEXT,
    CONSTRAINT device_runtime_command_type_valid
        CHECK (command_type IN ('refresh_configuration', 'reconnect_network', 'resync_time')),
    CONSTRAINT device_runtime_command_status_valid
        CHECK (status IN ('pending', 'delivered', 'acknowledged', 'failed', 'expired')),
    CONSTRAINT device_runtime_command_request_id_unique UNIQUE (request_id)
);

CREATE INDEX IF NOT EXISTS device_runtime_commands_device_idx
    ON device_runtime_commands(device_id, created_at DESC);
