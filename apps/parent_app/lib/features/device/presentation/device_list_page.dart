import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Shows devices bound to the active family.
class DeviceListPage extends StatelessWidget {
  const DeviceListPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Devices'),
        leading: IconButton(
          tooltip: 'Back',
          onPressed: () => context.go('/family'),
          icon: const Icon(Icons.arrow_back),
        ),
      ),
      body: const Center(child: Text('No devices are bound yet.')),
    );
  }
}
