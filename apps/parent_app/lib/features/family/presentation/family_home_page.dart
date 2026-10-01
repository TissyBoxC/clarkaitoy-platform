import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Landing page for the parent's family and device workspace.
class FamilyHomePage extends StatelessWidget {
  const FamilyHomePage({super.key});

  @override
  Widget build(BuildContext context) {
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
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(
            'Family workspace',
            style: Theme.of(context).textTheme.headlineSmall,
          ),
          const SizedBox(height: 12),
          const Card(
            child: ListTile(
              leading: Icon(Icons.family_restroom_outlined),
              title: Text('No family loaded'),
              subtitle: Text('Connect the parent API to load family data.'),
            ),
          ),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: () => context.go('/devices'),
            icon: const Icon(Icons.devices_outlined),
            label: const Text('View devices'),
          ),
        ],
      ),
    );
  }
}
