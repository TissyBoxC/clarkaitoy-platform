import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/core/network/api_client.dart';
import 'package:parent_app/features/auth/data/auth_api.dart';

void main() {
  test(
    'registration sends phone-first guardian and child profile fields',
    () async {
      final client = _RecordingApiClient();
      final api = AuthApi(client);

      await api.register(
        phone: '13800138000',
        phoneVerificationCode: '000000',
        password: 'sprout123',
        guardianFamilyName: '林',
        childNickname: '小芽',
        childBirthday: '2021-06-01',
        guardianConsentVersion: '2026-01',
      );

      expect(client.path, '/api/v1/auth/register');
      expect(client.body, {
        'phone': '13800138000',
        'phone_verification_code': '000000',
        'password': 'sprout123',
        'guardian_family_name': '林',
        'child_nickname': '小芽',
        'child_birthday': '2021-06-01',
        'guardian_consent_version': '2026-01',
      });
    },
  );

  test('login sends the unified phone-or-email identifier', () async {
    final client = _RecordingApiClient();
    final api = AuthApi(client);

    await api.login(identifier: 'guardian@example.com', password: 'sprout123');

    expect(client.path, '/api/v1/auth/login');
    expect(client.body, {
      'identifier': 'guardian@example.com',
      'password': 'sprout123',
    });
  });

  test('binding an email calls the authenticated account endpoint', () async {
    final client = _RecordingApiClient();
    final api = AuthApi(client);

    final account = await api.bindEmail(
      email: 'guardian@example.com',
      password: 'sprout123',
    );

    expect(client.path, '/api/v1/auth/email');
    expect(client.body, {
      'email': 'guardian@example.com',
      'password': 'sprout123',
    });
    expect(account.email, 'guardian@example.com');
  });
}

class _RecordingApiClient implements ApiClient {
  String? path;
  Object? body;

  @override
  Future<Map<String, Object?>> post(String path, {Object? body}) async {
    this.path = path;
    this.body = body;
    return _successEnvelope();
  }

  @override
  Future<Map<String, Object?>> get(
    String path, {
    Map<String, String>? queryParameters,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<Map<String, Object?>> put(String path, {Object? body}) async {
    this.path = path;
    this.body = body;
    return _successEnvelope(email: 'guardian@example.com');
  }

  @override
  Future<Map<String, Object?>> postWithBearerToken(
    String path, {
    Object? body,
    required String bearerToken,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<Map<String, Object?>> delete(String path) {
    throw UnimplementedError();
  }
}

Map<String, Object?> _successEnvelope({String email = ''}) {
  return {
    'data': {
      'access_token': 'access-token',
      'refresh_token': 'refresh-token',
      'expires_in': 900,
      'account': {
        'id': 'account-001',
        'email': email,
        'phone': '13800138000',
        'display_name': '林家长',
        'guardian_family_name': '林',
        'child_nickname': '小芽',
        'child_birthday': '2021-06-01',
        'role': 'parent',
        'status': 'active',
      },
      'ai_account': null,
    },
  };
}
