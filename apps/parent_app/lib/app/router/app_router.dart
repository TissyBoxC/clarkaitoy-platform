import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../features/device/presentation/device_list_page.dart';
import '../../features/family/presentation/family_home_page.dart';

/// Creates the application router.
///
/// Feature routes are registered here so removing a feature only requires
/// removing its route entry and package directory.
GoRouter createAppRouter() {
  return GoRouter(
    initialLocation: '/family',
    routes: [
      GoRoute(
        path: '/family',
        builder: (context, state) => const FamilyHomePage(),
      ),
      GoRoute(
        path: '/devices',
        builder: (context, state) => const DeviceListPage(),
      ),
    ],
    errorBuilder: (context, state) => const _RouteNotFoundPage(),
  );
}

class _RouteNotFoundPage extends StatelessWidget {
  const _RouteNotFoundPage();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Page not found')),
      body: const Center(child: Text('This page is not available.')),
    );
  }
}
