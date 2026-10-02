import 'package:dio/dio.dart';

import '../config/app_config.dart';
import '../error/api_error_mapper.dart';
import '../error/app_exception.dart';
import '../storage/secure_store.dart';

/// HTTP boundary used by feature data sources.
abstract interface class ApiClient {
  Future<Map<String, Object?>> get(
    String path, {
    Map<String, String>? queryParameters,
  });

  Future<Map<String, Object?>> post(String path, {Object? body});

  Future<Map<String, Object?>> postWithBearerToken(
    String path, {
    Object? body,
    required String bearerToken,
  });

  Future<Map<String, Object?>> delete(String path);
}

/// Configuration needed to construct the production API client.
class ApiClientConfig {
  const ApiClientConfig({required this.appConfig});

  final AppConfig appConfig;
}

/// Refreshes an access token from the current refresh token.
typedef RefreshSession = Future<AuthSession> Function(String refreshToken);

/// Short-lived access token plus renewable refresh token.
class AuthSession {
  const AuthSession({required this.accessToken, required this.refreshToken});

  final String accessToken;
  final String refreshToken;
}

/// Dio-backed API client with authentication refresh and error mapping.
class DioApiClient implements ApiClient {
  DioApiClient({
    required ApiClientConfig config,
    required SecureStore secureStore,
    required RefreshSession refreshSession,
    Dio? dio,
  }) : _secureStore = secureStore,
       _refreshSession = refreshSession,
       _dio =
           dio ??
           Dio(
             BaseOptions(
               baseUrl: config.appConfig.apiBaseUrl.toString(),
               connectTimeout: config.appConfig.requestTimeout,
               receiveTimeout: config.appConfig.requestTimeout,
               sendTimeout: config.appConfig.requestTimeout,
               headers: const {'Content-Type': 'application/json'},
             ),
           ) {
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: _attachAccessToken,
        onError: _refreshAfterUnauthorized,
      ),
    );
  }

  static const accessTokenKey = 'auth_access_token';
  static const refreshTokenKey = 'auth_refresh_token';

  final Dio _dio;
  final SecureStore _secureStore;
  final RefreshSession _refreshSession;
  Future<String>? _pendingRefresh;

  @override
  Future<Map<String, Object?>> get(
    String path, {
    Map<String, String>? queryParameters,
  }) async {
    try {
      final response = await _dio.get<Object?>(
        path,
        queryParameters: queryParameters,
      );
      return _requireResponseMap(response.data);
    } on Object catch (error) {
      throw mapApiError(error);
    }
  }

  @override
  Future<Map<String, Object?>> post(String path, {Object? body}) async {
    try {
      final response = await _dio.post<Object?>(path, data: body);
      return _requireResponseMap(response.data);
    } on Object catch (error) {
      throw mapApiError(error);
    }
  }

  @override
  Future<Map<String, Object?>> postWithBearerToken(
    String path, {
    Object? body,
    required String bearerToken,
  }) async {
    try {
      final response = await _dio.post<Object?>(
        path,
        data: body,
        options: Options(
          headers: {'Authorization': 'Bearer $bearerToken'},
          extra: const {'skip_auth': true},
        ),
      );
      return _requireResponseMap(response.data);
    } on Object catch (error) {
      throw mapApiError(error);
    }
  }

  @override
  Future<Map<String, Object?>> delete(String path) async {
    try {
      final response = await _dio.delete<Object?>(path);
      return _requireResponseMap(response.data);
    } on Object catch (error) {
      throw mapApiError(error);
    }
  }

  Future<void> _attachAccessToken(
    RequestOptions request,
    RequestInterceptorHandler handler,
  ) async {
    if (request.extra['skip_auth'] == true) {
      handler.next(request);
      return;
    }
    final accessToken = await _secureStore.read(accessTokenKey);
    if (accessToken != null) {
      request.headers['Authorization'] = 'Bearer $accessToken';
    }
    handler.next(request);
  }

  Future<void> _refreshAfterUnauthorized(
    DioException error,
    ErrorInterceptorHandler handler,
  ) async {
    final request = error.requestOptions;
    if (error.response?.statusCode != 401 ||
        request.extra['skip_auth'] == true ||
        request.extra['auth_retried'] == true) {
      handler.next(error);
      return;
    }

    try {
      final accessToken = await _refreshOnce();
      request.extra['auth_retried'] = true;
      request.headers['Authorization'] = 'Bearer $accessToken';
      final response = await _dio.fetch<Object?>(request);
      handler.resolve(response);
    } on Object catch (refreshError) {
      handler.next(refreshError is DioException ? refreshError : error);
    }
  }

  Future<String> _refreshOnce() {
    final pendingRefresh = _pendingRefresh;
    if (pendingRefresh != null) {
      return pendingRefresh;
    }

    final refresh = _performRefresh();
    _pendingRefresh = refresh;
    return refresh.whenComplete(() {
      _pendingRefresh = null;
    });
  }

  Future<String> _performRefresh() async {
    final refreshToken = await _secureStore.read(refreshTokenKey);
    if (refreshToken == null) {
      await _clearSession();
      throw const AppException(
        kind: AppErrorKind.unauthenticated,
        message: '登录已过期，请重新登录',
        retryable: false,
      );
    }

    try {
      final session = await _refreshSession(refreshToken);
      await _secureStore.write(accessTokenKey, session.accessToken);
      await _secureStore.write(refreshTokenKey, session.refreshToken);
      return session.accessToken;
    } on Object {
      await _clearSession();
      rethrow;
    }
  }

  Future<void> _clearSession() async {
    await _secureStore.delete(accessTokenKey);
    await _secureStore.delete(refreshTokenKey);
  }

  Map<String, Object?> _requireResponseMap(Object? data) {
    if (data is Map<String, Object?>) {
      return data;
    }
    if (data is Map) {
      return Map<String, Object?>.from(data);
    }
    throw const AppException(
      kind: AppErrorKind.unexpected,
      message: '操作没有完成，请稍后重试',
      retryable: true,
    );
  }
}
