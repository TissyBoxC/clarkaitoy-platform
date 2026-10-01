import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/core/error/api_error_mapper.dart';
import 'package:parent_app/core/error/app_exception.dart';

void main() {
  test('maps expired login without retry', () {
    final exception = mapApiError(
      DioException(
        requestOptions: RequestOptions(path: '/api/parent/v1/auth/me'),
        response: Response<void>(
          requestOptions: RequestOptions(path: '/api/parent/v1/auth/me'),
          statusCode: 401,
        ),
      ),
    );

    expect(exception.kind, AppErrorKind.unauthenticated);
    expect(exception.message, '登录已过期，请重新登录');
    expect(exception.retryable, isFalse);
  });

  test('maps connection failures to retryable network state', () {
    final exception = mapApiError(
      DioException(
        requestOptions: RequestOptions(path: '/api/parent/v1/devices'),
        type: DioExceptionType.connectionError,
      ),
    );

    expect(exception.kind, AppErrorKind.network);
    expect(exception.retryable, isTrue);
  });
}
