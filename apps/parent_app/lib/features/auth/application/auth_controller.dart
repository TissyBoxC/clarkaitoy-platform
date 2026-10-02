import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/error/app_exception.dart';
import '../../../core/providers.dart';
import '../../../core/storage/secure_store.dart';
import '../../../core/network/api_client.dart';
import '../data/auth_api.dart';

const guardianConsentVersion = '2026-01';

/// Immutable authentication state consumed by the router and pages.
class AuthState {
  const AuthState({
    required this.isLoading,
    this.account,
    this.aiAccount,
    this.errorMessage,
  });

  const AuthState.loading() : this(isLoading: true);

  const AuthState.signedOut({this.errorMessage})
    : isLoading = false,
      account = null,
      aiAccount = null;

  final bool isLoading;
  final ParentAccount? account;
  final AiAccount? aiAccount;
  final String? errorMessage;

  bool get isSignedIn => account != null;

  AuthState copyWith({
    bool? isLoading,
    ParentAccount? account,
    AiAccount? aiAccount,
    String? errorMessage,
  }) {
    return AuthState(
      isLoading: isLoading ?? this.isLoading,
      account: account ?? this.account,
      aiAccount: aiAccount ?? this.aiAccount,
      errorMessage: errorMessage,
    );
  }
}

/// Owns session restoration, registration, login, and sign-out.
class AuthController extends AsyncNotifier<AuthState> {
  late final AuthApi _authApi;
  late final SecureStore _secureStore;

  @override
  Future<AuthState> build() async {
    _authApi = ref.read(authApiProvider);
    _secureStore = ref.read(secureStoreProvider);
    final refreshToken = await _secureStore.read(DioApiClient.refreshTokenKey);
    if (refreshToken == null || refreshToken.isEmpty) {
      return const AuthState.signedOut();
    }
    try {
      final session = await _authApi.refresh(refreshToken);
      final account = await _authApi.me();
      await _saveSession(session);
      return AuthState(
        isLoading: false,
        account: account.account,
        aiAccount: account.aiAccount,
      );
    } on Object {
      await _clearSession();
      return const AuthState.signedOut();
    }
  }

  Future<void> register({
    required String email,
    required String password,
    required String displayName,
  }) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final session = await _authApi.register(
        email: email,
        password: password,
        displayName: displayName,
        guardianConsentVersion: guardianConsentVersion,
      );
      await _saveSession(session);
      return AuthState(
        isLoading: false,
        account: session.account,
        aiAccount: session.aiAccount,
      );
    });
  }

  Future<void> login({required String email, required String password}) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final session = await _authApi.login(email: email, password: password);
      await _saveSession(session);
      return AuthState(
        isLoading: false,
        account: session.account,
        aiAccount: session.aiAccount,
      );
    });
  }

  Future<void> logout() async {
    state = const AsyncLoading();
    final refreshToken = await _secureStore.read(DioApiClient.refreshTokenKey);
    if (refreshToken != null) {
      try {
        await _authApi.logout(refreshToken);
      } on Object {
        // A failed server revoke must not keep the device signed in.
      }
    }
    await _clearSession();
    state = const AsyncData(AuthState.signedOut());
  }

  Future<void> refreshAccount() async {
    final account = await _authApi.me();
    state = AsyncData(
      AuthState(
        isLoading: false,
        account: account.account,
        aiAccount: account.aiAccount,
      ),
    );
  }

  Future<void> _saveSession(AuthResult session) async {
    await _secureStore.write(DioApiClient.accessTokenKey, session.accessToken);
    await _secureStore.write(
      DioApiClient.refreshTokenKey,
      session.refreshToken,
    );
  }

  Future<void> _clearSession() async {
    await _secureStore.delete(DioApiClient.accessTokenKey);
    await _secureStore.delete(DioApiClient.refreshTokenKey);
  }
}

final authApiProvider = Provider<AuthApi>((ref) {
  return AuthApi(ref.read(apiClientProvider));
});

final authControllerProvider = AsyncNotifierProvider<AuthController, AuthState>(
  AuthController.new,
);

/// Converts a state error into copy the interface can safely display.
String authErrorMessage(Object? error) {
  if (error is AppException) {
    return error.message;
  }
  return '操作没有完成，请稍后重试';
}
