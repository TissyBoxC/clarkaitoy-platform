// Generated from https://sprout.example/contracts/schemas/envelope.schema.json; do not edit.

export interface APIResponseEnvelope {
    data:           unknown;
    error:          ErrorResponseSchema | null;
    request_id:     string;
    schema_version: SchemaVersion;
}

export interface ErrorResponseSchema {
    code:      string;
    details?:  Details;
    message:   string;
    retryable: boolean;
}

export interface Details {
    field?:  string;
    reason?: string;
}

export type SchemaVersion = "1.0.0";
