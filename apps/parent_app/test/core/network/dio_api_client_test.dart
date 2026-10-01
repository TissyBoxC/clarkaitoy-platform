import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/core/config/app_config.dart';
import 'package:parent_app/core/network/api_client.dart';
import 'package:parent_app/core/storage/secure_store.dart';

void main() {
  test('refreshes once and retries the original request', () async {
    final secureStore = InMemorySecureStore();
    await secureStore.write(DioApiClient.accessTokenKey, 'expired');
    await secureStore.write(DioApiClient.refreshTokenKey, 'refresh');

    final dio = Dio(BaseOptions(baseUrl: 'https://api.example.test'));
    dio.httpClientAdapter = _SequenceAdapter();

    var refreshCount = 0;
    final client = DioApiClient(
      config: ApiClientConfig(
        appConfig: AppConfig(
          apiBaseUrl: Uri.parse('https://api.example.test'),
          requestTimeout: const Duration(seconds: 1),
        ),
      ),
      secureStore: secureStore,
      refreshSession: (refreshToken) async {
        refreshCount += 1;
        return const AuthSession(
          accessToken: 'renewed',
          refreshToken: 'refresh-2',
        );
      },
      dio: dio,
    );

    final response = await client.get('/api/parent/v1/devices');

    expect(response, isNotEmpty);
    expect(refreshCount, 1);
    expect(await secureStore.read(DioApiClient.accessTokenKey), 'renewed');
  });
}

class InMemorySecureStore implements SecureStore {
  final Map<String, String> _values = {};

  @override
  Future<void> delete(String key) async {
    _values.remove(key);
  }

  @override
  Future<String?> read(String key) async {
    return _values[key];
  }

  @override
  Future<void> write(String key, String value) async {
    _values[key] = value;
  }
}

class _SequenceAdapter implements HttpClientAdapter {
  var _requestCount = 0;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    _requestCount += 1;
    if (_requestCount == 1) {
      return ResponseBody.fromString(
        jsonEncode(<String, Object?>{
          'schema_version': '1.0.0',
          'request_id': 'request-001',
          'data': null,
          'error': <String, Object?>{
            'code': 'unauthenticated',
            'message': '登录已过期，请重新登录',
            'retryable': false,
          },
        }),
        401,
        headers: <String, List<String>>{
          Headers.contentTypeHeader: <String>[Headers.jsonContentType],
        },
      );
    }

    return ResponseBody.fromString(
      jsonEncode(<String, Object?>{'data': <String, Object?>{}}),
      200,
      headers: <String, List<String>>{
        Headers.contentTypeHeader: <String>[Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
