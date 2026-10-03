import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../core/theme/app_theme.dart';
import '../features/ota/presentation/mandatory_update_gate.dart';
import 'router/app_router.dart';

/// Root widget that composes application-wide services and navigation.
class ParentApp extends ConsumerStatefulWidget {
  const ParentApp({super.key});

  @override
  ConsumerState<ParentApp> createState() => _ParentAppState();
}

class _ParentAppState extends ConsumerState<ParentApp> {
  GoRouter? _router;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _router ??= createAppRouter(ProviderScope.containerOf(context));
  }

  @override
  Widget build(BuildContext context) {
    final router = _router!;
    return MaterialApp.router(
      title: '如此萌屋',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      routerConfig: router,
      // The gate wraps the routed content so a required client update blocks
      // every route, including sign-in, without replacing the router.
      builder: (context, child) =>
          MandatoryUpdateGate(child: child ?? const SizedBox.shrink()),
    );
  }
}
