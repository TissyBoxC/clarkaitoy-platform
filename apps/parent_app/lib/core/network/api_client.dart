import '../config/app_config.dart';

/// Minimal HTTP boundary used by feature data sources.
///
/// The concrete Dio adapter is added when the platform API becomes available;
/// keeping this boundary small prevents UI code from depending on a client.
abstract interface class ApiClient {
  Future<Map<String, Object?>> get(
    String path, {
    Map<String, String>? queryParameters,
  });

  Future<Map<String, Object?>> post(String path, {Object? body});
}

/// Configuration needed to construct the production API client.
class ApiClientConfig {
  const ApiClientConfig({required this.appConfig, this.accessToken});

  final AppConfig appConfig;
  final String? accessToken;
}
