/// Category used by UI layers to choose a safe next action.
enum AppErrorKind {
  unauthenticated,
  insufficientPermission,
  notFound,
  validation,
  rateLimited,
  network,
  serviceUnavailable,
  unexpected,
}

/// Error that has already been translated for user-facing state.
class AppException implements Exception {
  const AppException({
    required this.kind,
    required this.message,
    required this.retryable,
    this.cause,
  });

  final AppErrorKind kind;
  final String message;
  final bool retryable;
  final Object? cause;

  @override
  String toString() => 'AppException($kind)';
}
