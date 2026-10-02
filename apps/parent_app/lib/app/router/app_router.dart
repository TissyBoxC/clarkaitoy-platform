import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/application/auth_controller.dart';
import '../../features/auth/presentation/bind_email_page.dart';
import '../../features/auth/presentation/login_page.dart';
import '../../features/auth/presentation/register_page.dart';
import '../../features/device/presentation/device_list_page.dart';
import '../../features/device/domain/device_payload.dart';
import '../../features/device/presentation/device_provisioning_page.dart';
import '../../features/device/presentation/device_qr_scan_page.dart';
import '../../features/family/presentation/family_home_page.dart';

/// Creates the application router with authentication-aware redirects.
GoRouter createAppRouter(ProviderContainer container) {
  return GoRouter(
    refreshListenable: _AuthRefreshListenable(container),
    initialLocation: '/family',
    redirect: (context, state) {
      final authState = container.read(authControllerProvider);
      if (authState.isLoading) {
        return null;
      }
      final isSignedIn = authState.value?.isSignedIn ?? false;
      final location = state.matchedLocation;
      final isPublicRoute = location == '/login' || location == '/register';
      if (!isSignedIn && !isPublicRoute) {
        return '/login';
      }
      if (isSignedIn && isPublicRoute) {
        return '/family';
      }
      return null;
    },
    routes: [
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(
        path: '/register',
        builder: (context, state) => const RegisterPage(),
      ),
      GoRoute(
        path: '/family',
        builder: (context, state) => const FamilyHomePage(),
      ),
      GoRoute(
        path: '/account/email',
        builder: (context, state) => const BindEmailPage(),
      ),
      GoRoute(
        path: '/devices',
        builder: (context, state) => const DeviceListPage(),
      ),
      GoRoute(
        path: '/devices/scan',
        builder: (context, state) => const DeviceQrScanPage(),
      ),
      GoRoute(
        path: '/devices/provision',
        builder: (context, state) {
          final setup = state.extra;
          if (setup is! DeviceSetupPayload) {
            return const _RouteNotFoundPage();
          }
          return DeviceProvisioningPage(setup: setup);
        },
      ),
    ],
    errorBuilder: (context, state) => const _RouteNotFoundPage(),
  );
}

class _AuthRefreshListenable extends ChangeNotifier {
  _AuthRefreshListenable(ProviderContainer container) {
    _subscription = container.listen(
      authControllerProvider,
      (_, _) => notifyListeners(),
      fireImmediately: true,
    );
  }

  late final ProviderSubscription<AsyncValue<AuthState>> _subscription;

  @override
  void dispose() {
    _subscription.close();
    super.dispose();
  }
}

class _RouteNotFoundPage extends StatelessWidget {
  const _RouteNotFoundPage();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('页面走丢了')),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Text('这个页面已经移动或不再存在。'),
              const SizedBox(height: 16),
              FilledButton(
                onPressed: () => context.go('/family'),
                child: const Text('返回首页'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
