import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/core/error/app_exception.dart';
import 'package:parent_app/core/providers.dart';
import 'package:parent_app/core/storage/secure_store.dart';
import 'package:parent_app/features/auth/application/auth_controller.dart';
import 'package:parent_app/features/auth/data/auth_api.dart';

void main() {
  test('restores a persisted refresh token on cold start', () async {
    final secureStore = _MemorySecureStore({'auth_refresh_token': 'refresh'});
    final api = _AuthApiStub();
    final container = ProviderContainer(
      overrides: [
        secureStoreProvider.overrideWithValue(secureStore),
        authApiProvider.overrideWithValue(api),
      ],
    );
    addTearDown(container.dispose);

    final state = await container.read(authControllerProvider.future);

    expect(state.isSignedIn, isTrue);
    expect(api.refreshCount, 1);
    expect(api.meCount, 1);
    expect(secureStore.values['auth_access_token'], 'access-renewed');
    expect(secureStore.values['auth_refresh_token'], 'refresh-renewed');
  });

  test('keeps the session when startup refresh has a network error', () async {
    final secureStore = _MemorySecureStore({'auth_refresh_token': 'refresh'});
    final api = _AuthApiStub(
      refreshError: const AppException(
        kind: AppErrorKind.network,
        message: '网络连接不稳定，请检查后重试',
        retryable: true,
      ),
    );
    final container = ProviderContainer(
      overrides: [
        secureStoreProvider.overrideWithValue(secureStore),
        authApiProvider.overrideWithValue(api),
      ],
    );
    addTearDown(container.dispose);

    final state = await container.read(authControllerProvider.future);

    expect(state.canRetryRestore, isTrue);
    expect(state.isSignedIn, isFalse);
    expect(secureStore.values['auth_refresh_token'], 'refresh');
  });

  test('clears the session when startup refresh is rejected', () async {
    final secureStore = _MemorySecureStore({
      'auth_access_token': 'access',
      'auth_refresh_token': 'refresh',
    });
    final api = _AuthApiStub(
      refreshError: const AppException(
        kind: AppErrorKind.unauthenticated,
        message: '登录已过期，请重新登录',
        retryable: false,
      ),
    );
    final container = ProviderContainer(
      overrides: [
        secureStoreProvider.overrideWithValue(secureStore),
        authApiProvider.overrideWithValue(api),
      ],
    );
    addTearDown(container.dispose);

    final state = await container.read(authControllerProvider.future);

    expect(state.isSignedIn, isFalse);
    expect(state.canRetryRestore, isFalse);
    expect(secureStore.values, isEmpty);
  });

  test('logout clears local session even when server logout fails', () async {
    final secureStore = _MemorySecureStore({
      'auth_access_token': 'access',
      'auth_refresh_token': 'refresh',
    });
    final api = _AuthApiStub(logoutError: StateError('network unavailable'));
    final container = ProviderContainer(
      overrides: [
        secureStoreProvider.overrideWithValue(secureStore),
        authApiProvider.overrideWithValue(api),
      ],
    );
    addTearDown(container.dispose);

    await container.read(authControllerProvider.future);
    await container.read(authControllerProvider.notifier).logout();

    expect(api.logoutCount, 1);
    expect(secureStore.values, isEmpty);
    expect(container.read(authControllerProvider).value?.isSignedIn, isFalse);
  });
}

class _MemorySecureStore implements SecureStore {
  _MemorySecureStore(this.values);

  final Map<String, String> values;

  @override
  Future<void> delete(String key) async {
    values.remove(key);
  }

  @override
  Future<String?> read(String key) async {
    return values[key];
  }

  @override
  Future<void> write(String key, String value) async {
    values[key] = value;
  }
}

class _AuthApiStub implements AuthApi {
  _AuthApiStub({this.refreshError, this.logoutError});

  final Object? refreshError;
  final Object? logoutError;
  int refreshCount = 0;
  int meCount = 0;
  int logoutCount = 0;

  @override
  Future<AuthResult> refresh(String refreshToken) async {
    refreshCount += 1;
    final error = refreshError;
    if (error != null) {
      throw error;
    }
    return const AuthResult(
      accessToken: 'access-renewed',
      refreshToken: 'refresh-renewed',
      account: _account,
    );
  }

  @override
  Future<AuthResult> me() async {
    meCount += 1;
    return const AuthResult(
      accessToken: '',
      refreshToken: '',
      account: _account,
    );
  }

  @override
  Future<void> logout(String refreshToken) async {
    logoutCount += 1;
    final error = logoutError;
    if (error != null) {
      throw error;
    }
  }

  @override
  Future<AuthResult> login({
    required String identifier,
    required String password,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<AuthResult> register({
    required String phone,
    required String phoneVerificationCode,
    required String password,
    required String guardianFamilyName,
    required String childNickname,
    required String childBirthday,
    required String guardianConsentVersion,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<void> sendPhoneVerification({
    required String phone,
    required String purpose,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<ParentAccount> bindEmail({
    required String email,
    required String password,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<AiAccount> updateSelectedModels(List<String> selectedModels) {
    throw UnimplementedError();
  }

  @override
  Future<ParentOverview> overview() {
    throw UnimplementedError();
  }

  @override
  Future<ParentAccount> updateProfile({
    required String displayName,
    required String guardianFamilyName,
    required String childNickname,
    required String childBirthday,
  }) {
    throw UnimplementedError();
  }

  @override
  Future<AppUpdateInfo?> appUpdate({
    required String platform,
    required String currentVersion,
  }) {
    throw UnimplementedError();
  }
}

const _account = ParentAccount(
  id: 'account-001',
  email: 'guardian@example.com',
  phone: '13800138000',
  displayName: '林家长',
  guardianFamilyName: '林',
  childNickname: '小芽',
  childBirthday: '2021-06-01',
  role: 'parent',
  status: 'active',
);
