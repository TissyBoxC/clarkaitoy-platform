import 'package:flutter/material.dart';

import '../../core/theme/app_motion.dart';

/// Reveals a section with a short vertical lift and fade.
///
/// Staggering should stay small enough that lists remain readable on entry.
class AppReveal extends StatelessWidget {
  const AppReveal({
    required this.child,
    this.delay = Duration.zero,
    this.duration = AppMotion.entrance,
    this.offset = 14,
    super.key,
  });

  final Widget child;
  final Duration delay;
  final Duration duration;
  final double offset;

  @override
  Widget build(BuildContext context) {
    final mediaQuery = MediaQuery.maybeOf(context);
    if (mediaQuery?.disableAnimations ?? false) {
      return child;
    }

    final totalDuration = duration + delay;
    final begin = totalDuration.inMicroseconds == 0
        ? 0.0
        : delay.inMicroseconds / totalDuration.inMicroseconds;
    return TweenAnimationBuilder<double>(
      tween: Tween<double>(begin: 0, end: 1),
      duration: totalDuration,
      curve: Interval(begin, 1, curve: AppMotion.enterCurve),
      child: child,
      builder: (context, value, child) {
        return Opacity(
          opacity: value,
          child: Transform.translate(
            offset: Offset(0, offset * (1 - value)),
            child: child,
          ),
        );
      },
    );
  }
}
