import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'secure_store.dart';

/// Platform-backed secure storage for access and refresh tokens.
class FlutterSecureStore implements SecureStore {
  FlutterSecureStore({FlutterSecureStorage? storage})
    : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;

  @override
  Future<String?> read(String key) {
    return _storage.read(key: key);
  }

  @override
  Future<void> write(String key, String value) {
    return _storage.write(key: key, value: value);
  }

  @override
  Future<void> delete(String key) {
    return _storage.delete(key: key);
  }
}
