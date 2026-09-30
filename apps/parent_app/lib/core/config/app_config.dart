/// Runtime configuration shared by the parent application.
class AppConfig {
  const AppConfig({required this.apiBaseUrl, required this.requestTimeout});

  final Uri apiBaseUrl;
  final Duration requestTimeout;

  /// Loads settings from compile-time values so local builds stay explicit.
  factory AppConfig.fromEnvironment() {
    const apiBaseUrl = String.fromEnvironment(
      'API_BASE_URL',
      defaultValue: 'https://api.example.invalid',
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
