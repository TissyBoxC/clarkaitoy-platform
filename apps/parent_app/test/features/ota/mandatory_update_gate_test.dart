import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/features/auth/data/auth_api.dart';
import 'package:parent_app/features/ota/application/app_update_service.dart';
import 'package:parent_app/features/ota/presentation/mandatory_update_gate.dart';

void main() {
  testWidgets('mandatory client update blocks the application', (tester) async {
    final service = _FakeUpdateCoordinator();
    await tester.pumpWidget(
      _gateApp(
        service: service,
        updateLoader: (_, _) async => _clientUpdate(isMandatory: true),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('需要更新'), findsOneWidget);
    expect(find.text('立即更新'), findsOneWidget);
    // There is no dismissal affordance: no "later" and no close control.
    expect(find.text('稍后更新'), findsNothing);
    expect(find.text('返回'), findsNothing);
  });

  testWidgets('optional client update does not block the application', (
    tester,
  ) async {
    final service = _FakeUpdateCoordinator();
    await tester.pumpWidget(
      _gateApp(
        service: service,
        updateLoader: (_, _) async => _clientUpdate(isMandatory: false),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('需要更新'), findsNothing);
    expect(find.text('应用首页'), findsOneWidget);
  });

  testWidgets('mandatory resource update never blocks a client update loop', (
    tester,
  ) async {
    final service = _FakeUpdateCoordinator();
    await tester.pumpWidget(
      _gateApp(
        service: service,
        updateLoader: (_, _) async => const AppUpdateInfo(
          kind: 'resource',
          version: '1.0.1',
          downloadUrl: 'https://example.com/resource.zip',
          sha256:
              'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
              'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
          releaseNotes: '内容更新',
          isMandatory: true,
          minSupportedVersion: '1.0.0',
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(find.text('需要更新'), findsNothing);
    expect(find.text('应用首页'), findsOneWidget);
  });

  testWidgets('install runs without any dismissal control appearing', (
    tester,
  ) async {
    final service = _FakeUpdateCoordinator();
    await tester.pumpWidget(
      _gateApp(
        service: service,
        updateLoader: (_, _) async => _clientUpdate(isMandatory: true),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    await tester.tap(find.text('立即更新'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(service.didInstallClientUpdate, isTrue);
    expect(find.text('等待安装完成'), findsOneWidget);
    expect(find.text('稍后更新'), findsNothing);
  });
}

Widget _gateApp({
  required AppUpdateCoordinator service,
  required Future<AppUpdateInfo?> Function(
    AppUpdateCoordinator service,
    String currentVersion,
  )
  updateLoader,
}) {
  return ProviderScope(
    overrides: [mandatoryUpdateCheckProvider.overrideWithValue(true)],
    child: MaterialApp(
      home: MandatoryUpdateGate(
        updateService: service,
        updateLoader: updateLoader,
        child: const Scaffold(body: Center(child: Text('应用首页'))),
      ),
    ),
  );
}

AppUpdateInfo _clientUpdate({required bool isMandatory}) {
  return AppUpdateInfo(
    kind: 'client',
    version: '1.0.1',
    downloadUrl: 'https://example.com/app.apk',
    sha256: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
    releaseNotes: '重要修复',
    isMandatory: isMandatory,
    minSupportedVersion: '1.0.0',
  );
}

class _FakeUpdateCoordinator implements AppUpdateCoordinator {
  bool didInstallClientUpdate = false;

  @override
  Future<String> installedVersion() async => '1.0.0';

  @override
  String? updateProblem(AppUpdateInfo update) => null;

  @override
  Future<void> installClientUpdate({
    required AppUpdateInfo update,
    required void Function(AppUpdateProgress progress) onProgress,
    CancelToken? cancelToken,
  }) async {
    didInstallClientUpdate = true;
    onProgress(
      const AppUpdateProgress(stage: AppUpdateStage.installing, fraction: 1),
    );
  }

  @override
  Future<void> installResourceUpdate({
    required AppUpdateInfo update,
    required void Function(AppUpdateProgress progress) onProgress,
    CancelToken? cancelToken,
  }) async {}
}
