import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/app/app.dart';
import 'package:parent_app/core/providers.dart';
import 'package:parent_app/core/storage/secure_store.dart';

void main() {
  testWidgets('signed-out users see the parent sign-in screen', (tester) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [secureStoreProvider.overrideWithValue(_EmptySecureStore())],
        child: const ParentApp(),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(find.text('欢迎回家'), findsOneWidget);
    expect(find.text('登录'), findsOneWidget);
    expect(find.text('还没有账号？创建家长账号'), findsOneWidget);
  });
}

class _EmptySecureStore implements SecureStore {
  @override
  Future<void> delete(String key) async {}

  @override
  Future<String?> read(String key) async => null;

  @override
  Future<void> write(String key, String value) async {}
}
