import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/features/device/presentation/device_qr_scan_page.dart';

void main() {
  testWidgets('scan page exposes a return control without a camera', (
    tester,
  ) async {
    final observer = _NavigationObserver();
    await tester.pumpWidget(
      ProviderScope(
        child: MaterialApp(
          navigatorObservers: [observer],
          home: Builder(
            builder: (context) => Scaffold(
              body: Center(
                child: FilledButton(
                  onPressed: () => Navigator.of(context).push(
                    MaterialPageRoute<void>(
                      builder: (_) => DeviceQrScanPage(
                        scannerViewBuilder: (_, _) =>
                            const ColoredBox(color: Colors.black),
                      ),
                    ),
                  ),
                  child: const Text('打开扫码页'),
                ),
              ),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('打开扫码页'));
    await tester.pumpAndSettle();

    expect(find.byTooltip('返回'), findsOneWidget);
    expect(find.text('扫描设备绑定码'), findsOneWidget);

    await tester.tap(find.byTooltip('返回'));
    await tester.pumpAndSettle();

    expect(find.text('打开扫码页'), findsOneWidget);
    expect(observer.popCount, 1);
  });

  testWidgets('system back uses the same cleanup path', (tester) async {
    final observer = _NavigationObserver();
    await tester.pumpWidget(
      ProviderScope(
        child: MaterialApp(
          navigatorObservers: [observer],
          home: Builder(
            builder: (context) => Scaffold(
              body: Center(
                child: FilledButton(
                  onPressed: () => Navigator.of(context).push(
                    MaterialPageRoute<void>(
                      builder: (_) => DeviceQrScanPage(
                        scannerViewBuilder: (_, _) =>
                            const ColoredBox(color: Colors.black),
                      ),
                    ),
                  ),
                  child: const Text('打开扫码页'),
                ),
              ),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('打开扫码页'));
    await tester.pumpAndSettle();
    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();

    expect(find.text('打开扫码页'), findsOneWidget);
    expect(observer.popCount, 1);
  });
}

class _NavigationObserver extends NavigatorObserver {
  int popCount = 0;

  @override
  void didPop(Route<dynamic> route, Route<dynamic>? previousRoute) {
    popCount += 1;
    super.didPop(route, previousRoute);
  }
}
