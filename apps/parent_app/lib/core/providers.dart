import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'config/app_config.dart';
import 'network/api_client.dart';
import 'storage/flutter_secure_store.dart';
import 'storage/secure_store.dart';

/// Platform secure storage shared by authentication and session recovery.
final secureStoreProvider = Provider<SecureStore>((ref) {
  return FlutterSecureStore();
});

/// Runtime configuration loaded from compile-time values.
final appConfigProvider = Provider<AppConfig>((ref) {
  return AppConfig.fromEnvironment();
});

/// Authenticated HTTP boundary for all feature data sources.
final apiClientProvider = Provider<ApiClient>((ref) {
  final secureStore = ref.watch(secureStoreProvider);
  final appConfig = ref.watch(appConfigProvider);
  late final DioApiClient client;
  client = DioApiClient(
    config: ApiClientConfig(appConfig: appConfig),
    secureStore: secureStore,
    refreshSession: (refreshToken) async {
      // The authentication controller and the HTTP client share the same
      // endpoint but the client only needs the narrow refresh contract.
      final response = await Dio().post<Object?>(
        '${appConfig.apiBaseUrl}/api/v1/auth/refresh',
        data: {'refresh_token': refreshToken},
      );
      final data = response.data;
      if (data is! Map) {
        throw StateError('invalid refresh response');
      }
      final envelope = Map<String, Object?>.from(data);
      final payload = envelope['data'];
      if (payload is! Map) {
        throw StateError('invalid refresh payload');
      }
      final session = Map<String, Object?>.from(payload);
      return AuthSession(
        accessToken: session['access_token'] as String,
        refreshToken: session['refresh_token'] as String,
      );
    },
  );
  return client;
});
