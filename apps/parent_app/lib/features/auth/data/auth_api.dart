import '../../../core/network/api_client.dart';

/// Authentication fields returned after register, login, and refresh.
class AuthResult {
  const AuthResult({
    required this.accessToken,
    required this.refreshToken,
    required this.account,
    this.aiAccount,
  });

  final String accessToken;
  final String refreshToken;
  final ParentAccount account;
  final AiAccount? aiAccount;

  factory AuthResult.fromResponse(Map<String, Object?> response) {
    final accountData = _requiredMap(response['account'], 'account');
    return AuthResult(
      accessToken: _requiredString(response['access_token'], 'access_token'),
      refreshToken: _requiredString(response['refresh_token'], 'refresh_token'),
      account: ParentAccount.fromJson(accountData),
      aiAccount: AiAccount.fromNullableJson(response['ai_account']),
    );
  }
}

/// Parent identity safe to display in the application.
class ParentAccount {
  const ParentAccount({
    required this.id,
    required this.email,
    required this.displayName,
    required this.role,
    required this.status,
  });

  final String id;
  final String email;
  final String displayName;
  final String role;
  final String status;

  factory ParentAccount.fromJson(Map<String, Object?> json) {
    return ParentAccount(
      id: _requiredString(json['id'], 'id'),
      email: _requiredString(json['email'], 'email'),
      displayName: _requiredString(json['display_name'], 'display_name'),
      role: _requiredString(json['role'], 'role'),
      status: _requiredString(json['status'], 'status'),
    );
  }
}

/// Parent-safe AI account projection. It never contains a provider API key.
class AiAccount {
  const AiAccount({
    required this.status,
    required this.balanceUsd,
    required this.concurrencyLimit,
    required this.allowedModels,
    required this.providerReady,
  });

  final String status;
  final double balanceUsd;
  final int concurrencyLimit;
  final List<String> allowedModels;
  final bool providerReady;

  static AiAccount? fromNullableJson(Object? value) {
    if (value is! Map) {
      return null;
    }
    return AiAccount.fromJson(Map<String, Object?>.from(value));
  }

  factory AiAccount.fromJson(Map<String, Object?> json) {
    return AiAccount(
      status: _requiredString(json['status'], 'status'),
      balanceUsd: _requiredDouble(json['balance_usd'], 'balance_usd'),
      concurrencyLimit: _requiredInt(
        json['concurrency_limit'],
        'concurrency_limit',
      ),
      allowedModels: _stringList(json['allowed_models']),
      providerReady: json['provider_ready'] == true,
    );
  }
}

/// Authentication API used by the application state.
class AuthApi {
  const AuthApi(this._apiClient);

  final ApiClient _apiClient;

  Future<AuthResult> register({
    required String email,
    required String password,
    required String displayName,
    required String guardianConsentVersion,
  }) async {
    final response = await _apiClient.post(
      '/api/v1/auth/register',
      body: {
        'email': email,
        'password': password,
        'display_name': displayName,
        'guardian_consent_version': guardianConsentVersion,
      },
    );
    return AuthResult.fromResponse(response['data'] as Map<String, Object?>);
  }

  Future<AuthResult> login({
    required String email,
    required String password,
  }) async {
    final response = await _apiClient.post(
      '/api/v1/auth/login',
      body: {'email': email, 'password': password},
    );
    return AuthResult.fromResponse(response['data'] as Map<String, Object?>);
  }

  Future<AuthResult> refresh(String refreshToken) async {
    final response = await _apiClient.post(
      '/api/v1/auth/refresh',
      body: {'refresh_token': refreshToken},
    );
    final data = response['data'] as Map<String, Object?>;
    return AuthResult(
      accessToken: _requiredString(data['access_token'], 'access_token'),
      refreshToken: _requiredString(data['refresh_token'], 'refresh_token'),
      account: const ParentAccount(
        id: '',
        email: '',
        displayName: '',
        role: 'parent',
        status: 'active',
      ),
    );
  }

  Future<void> logout(String refreshToken) async {
    await _apiClient.post(
      '/api/v1/auth/logout',
      body: {'refresh_token': refreshToken},
    );
  }

  Future<AuthResult> me() async {
    final response = await _apiClient.get('/api/v1/auth/me');
    final data = response['data'] as Map<String, Object?>;
    return AuthResult(
      accessToken: '',
      refreshToken: '',
      account: ParentAccount.fromJson(_requiredMap(data['account'], 'account')),
      aiAccount: AiAccount.fromNullableJson(data['ai_account']),
    );
  }
}

Map<String, Object?> _requiredMap(Object? value, String field) {
  if (value is Map) {
    return Map<String, Object?>.from(value);
  }
  throw FormatException('missing $field');
}

String _requiredString(Object? value, String field) {
  if (value is String && value.isNotEmpty) {
    return value;
  }
  throw FormatException('missing $field');
}

double _requiredDouble(Object? value, String field) {
  if (value is num) {
    return value.toDouble();
  }
  throw FormatException('missing $field');
}

int _requiredInt(Object? value, String field) {
  if (value is num) {
    return value.toInt();
  }
  throw FormatException('missing $field');
}

List<String> _stringList(Object? value) {
  if (value is! List) {
    return const [];
  }
  return value.whereType<String>().toList(growable: false);
}
