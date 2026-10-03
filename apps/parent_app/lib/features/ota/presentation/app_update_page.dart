import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/error/app_exception.dart';
import '../../../shared/widgets/app_reveal.dart';
import '../../auth/application/auth_controller.dart';
import '../../auth/data/auth_api.dart';
import '../application/app_update_service.dart';

/// Checks the platform release service and installs client or resource updates.
class AppUpdatePage extends ConsumerStatefulWidget {
  const AppUpdatePage({super.key});

  @override
  ConsumerState<AppUpdatePage> createState() => _AppUpdatePageState();
}

class _AppUpdatePageState extends ConsumerState<AppUpdatePage> {
  late final AppUpdateService _updateService;
  bool _isChecking = true;
  bool _isInstalling = false;
  _UpdateCheckResult? _result;
  AppUpdateProgress? _progress;
  String? _checkError;
  String? _actionError;

  @override
  void initState() {
    super.initState();
    _updateService = ref.read(appUpdateServiceProvider);
    _check();
  }

  Future<void> _check() async {
    setState(() {
      _isChecking = true;
      _checkError = null;
      _actionError = null;
    });
    try {
      final currentVersion = await _updateService.installedVersion();
      final update = await ref
          .read(authApiProvider)
          .appUpdate(
            platform: Platform.isAndroid ? 'android' : 'ios',
            currentVersion: currentVersion,
          );
      if (!mounted) {
        return;
      }
      setState(() {
        _isChecking = false;
        _result = _UpdateCheckResult(
          currentVersion: currentVersion,
          update: update,
        );
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isChecking = false;
        _result = null;
        _checkError = _errorMessage(error);
      });
    }
  }

  Future<void> _install() async {
    final update = _result?.update;
    if (update == null || _isInstalling) {
      return;
    }
    final blockingProblem = _updateService.updateProblem(update);
    if (blockingProblem != null) {
      setState(() => _actionError = blockingProblem);
      return;
    }
    final confirmed = await _confirmInstall(context, update);
    if (confirmed != true || !mounted) {
      return;
    }

    setState(() {
      _isInstalling = true;
      _actionError = null;
      _progress = const AppUpdateProgress(
        stage: AppUpdateStage.downloading,
        fraction: 0,
      );
    });
    try {
      if (update.isClientUpdate) {
        await _updateService.installClientUpdate(
          update: update,
          onProgress: _onProgress,
        );
      } else {
        await _updateService.installResourceUpdate(
          update: update,
          onProgress: _onProgress,
        );
      }
      if (!mounted) {
        return;
      }
      setState(() {
        _isInstalling = false;
        _progress = null;
      });
      if (update.isClientUpdate) {
        _showMessage('请按系统提示完成安装');
      } else {
        _showMessage('内容已更新，可以继续使用');
        await _check();
      }
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isInstalling = false;
        _actionError = _errorMessage(error);
      });
    }
  }

  void _onProgress(AppUpdateProgress progress) {
    if (!mounted) {
      return;
    }
    setState(() => _progress = progress);
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }

  @override
  Widget build(BuildContext context) {
    final update = _result?.update;
    return Scaffold(
      appBar: AppBar(title: const Text('软件更新')),
      body: _buildBody(update),
    );
  }

  Widget _buildBody(AppUpdateInfo? update) {
    if (_isChecking) {
      return const _CheckingUpdateState();
    }
    if (_checkError != null) {
      return _UpdateError(message: _checkError!, onRetry: _check);
    }
    if (update == null) {
      return _LatestVersionState(onCheckAgain: _check);
    }
    return _UpdateAvailableState(
      update: update,
      currentVersion: _result!.currentVersion,
      progress: _progress,
      actionError: _actionError ?? _updateService.updateProblem(update),
      isInstalling: _isInstalling,
      onInstall: _install,
      onCheckAgain: _check,
    );
  }
}

class _UpdateCheckResult {
  const _UpdateCheckResult({
    required this.currentVersion,
    required this.update,
  });

  final String currentVersion;
  final AppUpdateInfo? update;
}

class _CheckingUpdateState extends StatelessWidget {
  const _CheckingUpdateState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          CircularProgressIndicator(),
          SizedBox(height: 16),
          Text('正在检查更新...'),
        ],
      ),
    );
  }
}

class _LatestVersionState extends StatelessWidget {
  const _LatestVersionState({required this.onCheckAgain});

  final VoidCallback onCheckAgain;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: AppReveal(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                Icons.verified_rounded,
                size: 56,
                color: Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(height: 18),
              Text('已经是最新版本', style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height: 8),
              const Text('当前版本运行正常，暂时没有需要更新。'),
              const SizedBox(height: 18),
              OutlinedButton(
                onPressed: onCheckAgain,
                child: const Text('重新检查更新'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _UpdateAvailableState extends StatelessWidget {
  const _UpdateAvailableState({
    required this.update,
    required this.currentVersion,
    required this.progress,
    required this.actionError,
    required this.isInstalling,
    required this.onInstall,
    required this.onCheckAgain,
  });

  final AppUpdateInfo update;
  final String currentVersion;
  final AppUpdateProgress? progress;
  final String? actionError;
  final bool isInstalling;
  final VoidCallback onInstall;
  final VoidCallback onCheckAgain;

  @override
  Widget build(BuildContext context) {
    final isClientUpdate = update.isClientUpdate;
    final isBlocked = actionError != null;
    return ListView(
      padding: const EdgeInsets.all(20),
      children: [
        AppReveal(
          child: Card(
            child: Padding(
              padding: const EdgeInsets.all(22),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        width: 46,
                        height: 46,
                        decoration: BoxDecoration(
                          color: Theme.of(context).colorScheme.primaryContainer,
                          borderRadius: BorderRadius.circular(16),
                        ),
                        child: Icon(
                          isClientUpdate
                              ? Icons.system_update_alt_rounded
                              : Icons.auto_awesome_rounded,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              isClientUpdate ? '发现新版本' : '有新的内容可以更新',
                              style: Theme.of(context).textTheme.titleLarge,
                            ),
                            const SizedBox(height: 4),
                            Text('当前 $currentVersion，可更新到 ${update.version}'),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 18),
                  Text(
                    isClientUpdate
                        ? '更新会下载并安装新的应用版本，你的账号、设备绑定、资料和对话记录都会保留。'
                        : '更新只替换资源内容，不需要重新安装，完成后即可继续使用。',
                  ),
                  if (update.releaseNotes.isNotEmpty) ...[
                    const SizedBox(height: 16),
                    Text(
                      '本次更新',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 6),
                    Text(update.releaseNotes),
                  ],
                  if (update.isMandatory) ...[
                    const SizedBox(height: 14),
                    const _NoticeBanner(
                      icon: Icons.shield_outlined,
                      text: '这次更新包含必要修复，请尽快完成。',
                    ),
                  ],
                  if (progress != null) ...[
                    const SizedBox(height: 18),
                    _UpdateProgressView(progress: progress!),
                  ],
                  if (actionError != null) ...[
                    const SizedBox(height: 14),
                    _ActionErrorNotice(message: actionError!),
                  ],
                  const SizedBox(height: 22),
                  FilledButton.icon(
                    onPressed: isInstalling || isBlocked ? null : onInstall,
                    icon: Icon(
                      isInstalling
                          ? Icons.hourglass_top_rounded
                          : isClientUpdate
                          ? Icons.download_rounded
                          : Icons.refresh_rounded,
                    ),
                    label: Text(
                      isInstalling
                          ? '正在准备更新'
                          : isBlocked
                          ? '暂时无法更新'
                          : (isClientUpdate ? '下载并安装' : '立即更新'),
                    ),
                  ),
                  const SizedBox(height: 10),
                  Center(
                    child: TextButton(
                      onPressed: isInstalling ? null : onCheckAgain,
                      child: const Text('重新检查更新'),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
        const SizedBox(height: 14),
        Text(
          '更新包会经过完整性校验后才使用。升级过程中请保持设备联网。',
          textAlign: TextAlign.center,
          style: Theme.of(context).textTheme.bodySmall,
        ),
      ],
    );
  }
}

class _UpdateProgressView extends StatelessWidget {
  const _UpdateProgressView({required this.progress});

  final AppUpdateProgress progress;

  @override
  Widget build(BuildContext context) {
    final label = switch (progress.stage) {
      AppUpdateStage.downloading => '正在下载更新包',
      AppUpdateStage.verifying => '正在校验更新包',
      AppUpdateStage.extracting => '正在准备新内容',
      AppUpdateStage.installing => '正在打开安装页面',
    };
    final percent = (progress.fraction * 100).round();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(child: Text(label)),
            if (progress.stage == AppUpdateStage.downloading) Text('$percent%'),
          ],
        ),
        const SizedBox(height: 8),
        ClipRRect(
          borderRadius: BorderRadius.circular(8),
          child: LinearProgressIndicator(
            value: progress.stage == AppUpdateStage.downloading
                ? progress.fraction
                : null,
            minHeight: 8,
          ),
        ),
      ],
    );
  }
}

class _ActionErrorNotice extends StatelessWidget {
  const _ActionErrorNotice({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.errorContainer,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        children: [
          Icon(
            Icons.error_outline_rounded,
            size: 20,
            color: Theme.of(context).colorScheme.onErrorContainer,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: TextStyle(
                color: Theme.of(context).colorScheme.onErrorContainer,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _NoticeBanner extends StatelessWidget {
  const _NoticeBanner({required this.icon, required this.text});

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.tertiaryContainer,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        children: [
          Icon(
            icon,
            size: 20,
            color: Theme.of(context).colorScheme.onTertiaryContainer,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              text,
              style: TextStyle(
                color: Theme.of(context).colorScheme.onTertiaryContainer,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _UpdateError extends StatelessWidget {
  const _UpdateError({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(28),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off_outlined, size: 46),
            const SizedBox(height: 14),
            Text(message, textAlign: TextAlign.center),
            const SizedBox(height: 16),
            FilledButton(onPressed: onRetry, child: const Text('重新检查更新')),
          ],
        ),
      ),
    );
  }
}

Future<bool?> _confirmInstall(BuildContext context, AppUpdateInfo update) {
  return showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text(update.isClientUpdate ? '准备安装新版本' : '准备更新内容'),
      content: Text(
        update.isClientUpdate
            ? '接下来会下载安装包，并在完成后打开系统安装页面。你的账号、设备绑定和资料都会保留。'
            : '接下来会下载新的资源内容，不会影响你的账号和设备信息。',
      ),
      actions: [
        if (!update.isMandatory)
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('稍后更新'),
          ),
        FilledButton(
          onPressed: () => Navigator.of(context).pop(true),
          child: Text(update.isClientUpdate ? '开始下载' : '开始更新'),
        ),
      ],
    ),
  );
}

String _errorMessage(Object? error) {
  if (error is AppException) {
    return error.message;
  }
  return '暂时无法检查更新，请稍后重试';
}
