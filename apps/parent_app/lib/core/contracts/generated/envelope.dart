// Generated from https://sprout.example/contracts/schemas/envelope.schema.json; do not edit.

class ApiResponseEnvelope {
  final dynamic data;
  final ErrorResponseSchema? error;
  final String requestId;
  final String schemaVersion;

  ApiResponseEnvelope({
    required this.data,
    required this.error,
    required this.requestId,
    required this.schemaVersion,
  });
}

class ErrorResponseSchema {
  final String code;
  final Details? details;
  final String message;
  final bool retryable;

  ErrorResponseSchema({
    required this.code,
    this.details,
    required this.message,
    required this.retryable,
  });
}

class Details {
  final String? field;
  final String? reason;

  Details({this.field, this.reason});
}

class SchemaVersion {
  static const String the100 = '1.0.0';
}
