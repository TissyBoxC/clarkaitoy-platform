import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:parent_app/features/auth/data/auth_api.dart';
import 'package:parent_app/features/ota/application/app_update_service.dart';
import 'package:parent_app/features/ota/presentation/app_update_page.dart';

void main() {
  testWidgets('update page exposes a visible return control while checking', (
    tester,
  ) async {
    final service = _FakeUpdateCoordinator(
      installedVersionFuture: Completer<String>().future,
    );
    await tester.pumpWidget(_updateApp(service));
    await tester.tap(find.text('打开更新页'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.byTooltip('返回'), findsOneWidget);
    expect(find.text('正在检查更新...'), findsOneWidget);

    await tester.tap(find.byTooltip('返回'));
    await tester.pumpAndSettle();
    expect(find.text('打开更新页'), findsOneWidget);
  });

  testWidgets('system back returns while update download is active', (
    tester,
  ) async {
    final service = _FakeUpdateCoordinator(installCompleter: Completer<void>());
    await tester.pumpWidget(_updateApp(service));
    await tester.tap(find.text('打开更新页'));
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(FilledButton, '立即更新'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, '开始更新'));
    await tester.pump();

    expect(find.text('停止更新'), findsOneWidget);
    await tester.binding.handlePopRoute();
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    expect(find.text('停止更新'), findsWidgets);

    await tester.tap(find.widgetWithText(FilledButton, '停止更新').last);
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(service.cancelToken?.isCancelled, isTrue);
    expect(find.text('打开更新页'), findsOneWidget);
  });

  testWidgets('stop update control cancels without leaving the page', (
    tester,
  ) async {
    final service = _FakeUpdateCoordinator(installCompleter: Completer<void>());
    await tester.pumpWidget(_updateApp(service));
    await tester.tap(find.text('打开更新页'));
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(FilledButton, '立即更新'));
    await tester.pumpAndSettle();
    await tester.tap(find.widgetWithText(FilledButton, '开始更新'));
    await tester.pump();

    await tester.tap(find.widgetWithText(OutlinedButton, '停止更新'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.tap(find.widgetWithText(FilledButton, '停止更新'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));

    expect(service.cancelToken?.isCancelled, isTrue);
    expect(find.text('软件更新'), findsOneWidget);
    expect(find.text('更新已停止，可以稍后重新开始'), findsOneWidget);
  });
}

Widget _updateApp(AppUpdateCoordinator service) {
  return ProviderScope(
    child: MaterialApp(
      home: Builder(
        builder: (context) => Scaffold(
          body: Center(
            child: FilledButton(
              onPressed: () => Navigator.of(context).push(
                MaterialPageRoute<void>(
                  builder: (_) => AppUpdatePage(
                    updateService: service,
                    updateLoader: (_, _) async => _resourceUpdate(),
                  ),
                ),
              ),
              child: const Text('打开更新页'),
            ),
          ),
        ),
      ),
    ),
  );
}

AppUpdateInfo _resourceUpdate() {
  return const AppUpdateInfo(
    kind: 'resource',
    version: '1.0.1',
    downloadUrl: 'https://example.com/resource.zip',
    sha256: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa',
    releaseNotes: '更新内容',
    isMandatory: false,
    minSupportedVersion: '1.0.0',
  );
}

class _FakeUpdateCoordinator implements AppUpdateCoordinator {
  _FakeUpdateCoordinator({
    Future<String>? installedVersionFuture,
    this.installCompleter,
  }) : installedVersionFuture =
           installedVersionFuture ?? Future<String>.value('1.0.0');

  final Future<String> installedVersionFuture;
  final Completer<void>? installCompleter;
  CancelToken? cancelToken;

  @override
  Future<String> installedVersion() {
    return installedVersionFuture;
  }

  @override
  String? updateProblem(AppUpdateInfo update) => null;

  @override
  Future<void> installClientUpdate({
    required AppUpdateInfo update,
    required void Function(AppUpdateProgress progress) onProgress,
    CancelToken? cancelToken,
  }) {
    return _install(cancelToken: cancelToken);
  }

  @override
  Future<void> installResourceUpdate({
    required AppUpdateInfo update,
    required void Function(AppUpdateProgress progress) onProgress,
    CancelToken? cancelToken,
  }) {
    return _install(cancelToken: cancelToken);
  }

  Future<void> _install({CancelToken? cancelToken}) async {
    this.cancelToken = cancelToken;
    if (cancelToken != null) {
      await Future.any([
        installCompleter?.future ?? Future<void>.value(),
        cancelToken.whenCancel,
      ]);
      if (cancelToken.isCancelled) {
        throw cancelToken.cancelError!;
      }
      return;
    }
    await installCompleter?.future;
  }
}
