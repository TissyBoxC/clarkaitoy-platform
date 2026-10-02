import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../auth/application/auth_controller.dart';
import '../../auth/data/auth_api.dart';
import '../../device/application/device_binding_controller.dart';
import '../../../shared/widgets/app_reveal.dart';
import '../../../shared/widgets/app_state_switcher.dart';

/// Parent workspace with account, AI usage, and device entry points.
class FamilyHomePage extends ConsumerWidget {
  const FamilyHomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final auth = ref.watch(authControllerProvider);
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            ClipOval(
              child: Image.asset(
                'assets/brand/sprout/brand_avatar.png',
                width: 32,
                height: 32,
                semanticLabel: '如此萌屋',
              ),
            ),
            const SizedBox(width: 10),
            const Text('如此萌屋'),
          ],
        ),
        actions: [
          IconButton(
            tooltip: '退出登录',
            onPressed: () =>
                ref.read(authControllerProvider.notifier).logout().then((_) {
                  if (context.mounted) {
                    context.go('/login');
                  }
                }),
            icon: const Icon(Icons.logout),
          ),
        ],
      ),
      body: AppStateSwitcher(
        stateKey: auth.when(
          data: (value) => value.account == null ? 'signed-out' : 'content',
          error: (_, _) => 'error',
          loading: () => 'loading',
        ),
        child: auth.when(
          loading: () => const Center(
            key: ValueKey<String>('loading'),
            child: CircularProgressIndicator(),
          ),
          error: (error, _) => _ErrorState(
            key: const ValueKey<String>('error'),
            message: authErrorMessage(error),
            onRetry: () =>
                ref.read(authControllerProvider.notifier).refreshAccount(),
          ),
          data: (state) {
            final account = state.account;
            if (account == null) {
              return const _SignInPrompt(key: ValueKey<String>('signed-out'));
            }
            final devices = ref.watch(deviceBindingControllerProvider);
            return RefreshIndicator(
              key: const ValueKey<String>('content'),
              onRefresh: () async {
                await ref
                    .read(authControllerProvider.notifier)
                    .refreshAccount();
                await ref
                    .read(deviceBindingControllerProvider.notifier)
                    .refresh();
              },
              child: ListView(
                padding: const EdgeInsets.all(16),
                children: [
                  AppReveal(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '你好，${account.displayName}',
                          style: Theme.of(context).textTheme.headlineSmall,
                        ),
                        const SizedBox(height: 6),
                        const Text('孩子今天想聊些什么？'),
                      ],
                    ),
                  ),
                  const SizedBox(height: 20),
                  AppReveal(
                    delay: const Duration(milliseconds: 70),
                    child: _AiAccountCard(
                      aiAccount: state.aiAccount,
                      onRetry: () => ref
                          .read(authControllerProvider.notifier)
                          .retryAIService(),
                      onModelsChanged: (models) => ref
                          .read(authControllerProvider.notifier)
                          .updateSelectedModels(models),
                    ),
                  ),
                  const SizedBox(height: 20),
                  AppReveal(
                    delay: const Duration(milliseconds: 130),
                    child: _AccountSecurityCard(account: account),
                  ),
                  const SizedBox(height: 20),
                  AppReveal(
                    delay: const Duration(milliseconds: 180),
                    child: Row(
                      children: [
                        Text(
                          '我的设备',
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        const Spacer(),
                        TextButton.icon(
                          onPressed: () => context.go('/devices'),
                          icon: const Icon(Icons.add_circle_outline),
                          label: const Text('添加设备'),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 8),
                  AppStateSwitcher(
                    stateKey: devices.when(
                      data: (deviceState) => deviceState.bindings.isEmpty
                          ? 'devices-empty'
                          : 'devices-content',
                      error: (_, _) => 'devices-error',
                      loading: () => 'devices-loading',
                    ),
                    child: devices.when(
                      loading: () => const Card(
                        key: ValueKey<String>('devices-loading'),
                        child: ListTile(
                          leading: SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                          title: Text('正在读取设备…'),
                        ),
                      ),
                      error: (error, _) => _InlineError(
                        key: const ValueKey<String>('devices-error'),
                        message: authErrorMessage(error),
                        onRetry: () => ref
                            .read(deviceBindingControllerProvider.notifier)
                            .refresh(),
                      ),
                      data: (deviceState) {
                        final bindings = deviceState.bindings;
                        if (deviceState.errorMessage != null) {
                          return _InlineError(
                            key: const ValueKey<String>('devices-error'),
                            message: deviceState.errorMessage!,
                            onRetry: () => ref
                                .read(deviceBindingControllerProvider.notifier)
                                .refresh(),
                          );
                        }
                        if (bindings.isEmpty) {
                          return const Card(
                            key: ValueKey<String>('devices-empty'),
                            child: ListTile(
                              leading: Icon(Icons.toys_outlined),
                              title: Text('还没有绑定设备'),
                              subtitle: Text('打开初芽的配网页，用手机扫描二维码或在附近设备中添加。'),
                            ),
                          );
                        }
                        return Column(
                          key: const ValueKey<String>('devices-content'),
                          children: [
                            for (
                              var index = 0;
                              index < bindings.length;
                              index++
                            )
                              AppReveal(
                                delay: Duration(
                                  milliseconds: 70 * index.clamp(0, 5),
                                ),
                                child: Card(
                                  child: ListTile(
                                    leading: const Icon(Icons.toys_outlined),
                                    title: Text(bindings[index].deviceName),
                                    subtitle: Text(
                                      bindings[index].boundAt
                                          .toLocal()
                                          .toString()
                                          .split(' ')
                                          .first,
                                    ),
                                    trailing: const Icon(Icons.chevron_right),
                                    onTap: () => context.go('/devices'),
                                  ),
                                ),
                              ),
                          ],
                        );
                      },
                    ),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }
}

class _AccountSecurityCard extends StatelessWidget {
  const _AccountSecurityCard({required this.account});

  final ParentAccount account;

  @override
  Widget build(BuildContext context) {
    final hasEmail = account.email.isNotEmpty;
    return Card(
      child: ListTile(
        leading: const Icon(Icons.mark_email_read_outlined),
        title: Text(hasEmail ? account.email : '未绑定登录邮箱'),
        subtitle: Text(hasEmail ? '可以使用手机号或邮箱登录' : '绑定后可以用邮箱登录，也方便找回账号'),
        trailing: TextButton(
          onPressed: () => context.go('/account/email'),
          child: Text(hasEmail ? '更换邮箱' : '绑定邮箱'),
        ),
      ),
    );
  }
}

class _AiAccountCard extends StatelessWidget {
  const _AiAccountCard({
    required this.aiAccount,
    required this.onRetry,
    required this.onModelsChanged,
  });

  final AiAccount? aiAccount;
  final Future<void> Function() onRetry;
  final Future<void> Function(List<String>) onModelsChanged;

  @override
  Widget build(BuildContext context) {
    final account = aiAccount;
    if (account == null) {
      return Card(
        child: ListTile(
          leading: const Icon(Icons.cloud_off_outlined),
          title: const Text('AI 服务正在准备'),
          subtitle: const Text('完成前不会产生对话费用，也不需要重新注册。'),
          trailing: TextButton(onPressed: onRetry, child: const Text('重新准备')),
        ),
      );
    }
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.auto_awesome_outlined),
                const SizedBox(width: 10),
                Text('AI 陪伴额度', style: Theme.of(context).textTheme.titleMedium),
              ],
            ),
            const SizedBox(height: 16),
            Text('可用余额', style: Theme.of(context).textTheme.bodySmall),
            const SizedBox(height: 4),
            Text(
              '\$${account.balanceUsd.toStringAsFixed(2)}',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            const SizedBox(height: 12),
            Text('同时对话：${account.concurrencyLimit} 台'),
            const SizedBox(height: 4),
            Text(
              account.selectedModels.isEmpty
                  ? '模型：由家长端统一安排'
                  : '模型：${account.selectedModels.join('、')}',
            ),
            const SizedBox(height: 10),
            Align(
              alignment: Alignment.centerLeft,
              child: TextButton.icon(
                onPressed: () => _showModelPicker(context, account),
                icon: const Icon(Icons.tune),
                label: const Text('选择可用模型'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _showModelPicker(BuildContext context, AiAccount account) async {
    // An empty saved selection means "all available", so the picker opens with
    // every approved model checked instead of showing a misleading empty state.
    final selectedModels = <String>{
      if (account.selectedModels.isEmpty)
        ...account.availableModels
      else
        ...account.selectedModels,
    };
    final saved = await showModalBottomSheet<bool>(
      context: context,
      showDragHandle: true,
      builder: (context) => StatefulBuilder(
        builder: (context, setModalState) => Padding(
          padding: const EdgeInsets.fromLTRB(20, 0, 20, 24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Text('选择对话模型', style: Theme.of(context).textTheme.titleLarge),
              const SizedBox(height: 6),
              const Text('只勾选允许孩子使用的模型，至少要保留一个。'),
              const SizedBox(height: 12),
              if (account.availableModels.isEmpty)
                const ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: Icon(Icons.info_outline),
                  title: Text('当前没有可选模型'),
                  subtitle: Text('请稍后重新准备 AI 服务。'),
                )
              else
                ...account.availableModels.map(
                  (model) => CheckboxListTile(
                    value: selectedModels.contains(model),
                    contentPadding: EdgeInsets.zero,
                    title: Text(model),
                    onChanged: (isSelected) => setModalState(() {
                      if (isSelected == true) {
                        selectedModels.add(model);
                      } else {
                        selectedModels.remove(model);
                      }
                    }),
                  ),
                ),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: selectedModels.isEmpty
                    ? null
                    : () => Navigator.of(context).pop(true),
                child: const Text('保存模型选择'),
              ),
            ],
          ),
        ),
      ),
    );
    if (saved == true) {
      await onModelsChanged(selectedModels.toList(growable: false));
    }
  }
}

class _SignInPrompt extends StatelessWidget {
  const _SignInPrompt({super.key});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.lock_outline, size: 40),
            const SizedBox(height: 12),
            const Text('请先登录家长账号'),
            const SizedBox(height: 16),
            FilledButton(
              onPressed: () => context.go('/login'),
              child: const Text('去登录'),
            ),
          ],
        ),
      ),
    );
  }
}

class _InlineError extends StatelessWidget {
  const _InlineError({required this.message, required this.onRetry, super.key});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: const Icon(Icons.error_outline),
        title: Text(message),
        trailing: TextButton(onPressed: onRetry, child: const Text('重试')),
      ),
    );
  }
}

class _ErrorState extends StatelessWidget {
  const _ErrorState({required this.message, required this.onRetry, super.key});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(message),
            const SizedBox(height: 12),
            FilledButton(onPressed: onRetry, child: const Text('重试')),
          ],
        ),
      ),
    );
  }
}
