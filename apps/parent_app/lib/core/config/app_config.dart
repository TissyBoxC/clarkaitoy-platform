/// Runtime configuration shared by the parent application.
class AppConfig {
  const AppConfig({required this.apiBaseUrl, required this.requestTimeout});

  final Uri apiBaseUrl;
  final Duration requestTimeout;

  /// Loads settings from compile-time values so local builds stay explicit.
  ///
  /// Release and CI builds must pass API_BASE_URL produced from
  /// deploy/public-endpoints.env; the local Android emulator default is only a
  /// development convenience and must never ship in a release artifact.
  factory AppConfig.fromEnvironment() {
    const apiBaseUrl = String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'http://10.0.2.2:8081',
    );
    const timeoutSeconds = int.fromEnvironment(
      'API_TIMEOUT_SECONDS',
      defaultValue: 15,
    );

    return AppConfig(
      apiBaseUrl: Uri.parse(apiBaseUrl),
      requestTimeout: Duration(seconds: timeoutSeconds),
    );
  }
}
