import 'package:flutter/material.dart';

/// Shared timing and route motion for the parent application.
///
/// Keeping animation values in one place prevents pages from developing
/// different pacing while still allowing local emphasis where needed.
abstract final class AppMotion {
  static const Duration fast = Duration(milliseconds: 140);
  static const Duration standard = Duration(milliseconds: 240);
  static const Duration slow = Duration(milliseconds: 360);
  static const Duration entrance = Duration(milliseconds: 440);

  static const Curve enterCurve = Cubic(0.22, 1, 0.36, 1);
  static const Curve exitCurve = Curves.easeInCubic;
  static const Curve emphasizedCurve = Curves.easeOutBack;

  static const PageTransitionsTheme pageTransitionsTheme = PageTransitionsTheme(
    builders: <TargetPlatform, PageTransitionsBuilder>{
      TargetPlatform.android: _SproutPageTransitionsBuilder(),
      TargetPlatform.iOS: _SproutPageTransitionsBuilder(),
      TargetPlatform.macOS: _SproutPageTransitionsBuilder(),
      TargetPlatform.windows: _SproutPageTransitionsBuilder(),
      TargetPlatform.linux: _SproutPageTransitionsBuilder(),
      TargetPlatform.fuchsia: _SproutPageTransitionsBuilder(),
    },
  );
}

class _SproutPageTransitionsBuilder extends PageTransitionsBuilder {
  const _SproutPageTransitionsBuilder();

  @override
  Widget buildTransitions<T>(
    PageRoute<T> route,
    BuildContext context,
    Animation<double> animation,
    Animation<double> secondaryAnimation,
    Widget child,
  ) {
    final curvedAnimation = CurvedAnimation(
      parent: animation,
      curve: AppMotion.enterCurve,
      reverseCurve: AppMotion.exitCurve,
    );
    return FadeTransition(
      opacity: curvedAnimation,
      child: SlideTransition(
        position: Tween<Offset>(
          begin: const Offset(0.025, 0.018),
          end: Offset.zero,
        ).animate(curvedAnimation),
        child: child,
      ),
    );
  }
}
