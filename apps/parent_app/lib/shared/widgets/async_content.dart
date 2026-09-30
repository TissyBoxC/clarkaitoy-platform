import 'package:flutter/material.dart';

/// Shared widget for rendering a loading, error, or content state.
class AsyncContent extends StatelessWidget {
  const AsyncContent({
    required this.isLoading,
    required this.hasError,
    required this.child,
    this.onRetry,
    super.key,
  });

  final bool isLoading;
  final bool hasError;
  final Widget child;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (hasError) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('Unable to load this page.'),
            const SizedBox(height: 12),
            FilledButton(onPressed: onRetry, child: const Text('Retry')),
          ],
        ),
      );
    }
    return child;
  }
}
