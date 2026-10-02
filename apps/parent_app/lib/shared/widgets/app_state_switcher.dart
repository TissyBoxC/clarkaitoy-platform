import 'package:flutter/material.dart';

import '../../core/theme/app_motion.dart';

/// Cross-fades and lifts a page state when its identity changes.
///
/// Use a stable [stateKey] per loading, empty, error, and content state so the
/// switcher does not replay animation for unrelated rebuilds.
class AppStateSwitcher extends StatelessWidget {
  const AppStateSwitcher({
    required this.stateKey,
    required this.child,
    this.alignment = Alignment.topCenter,
    super.key,
  });

  final Object stateKey;
  final Widget child;
  final Alignment alignment;

  @override
  Widget build(BuildContext context) {
    return AnimatedSwitcher(
      duration: AppMotion.standard,
      reverseDuration: AppMotion.fast,
      switchInCurve: AppMotion.enterCurve,
      switchOutCurve: AppMotion.exitCurve,
      layoutBuilder: (currentChild, previousChildren) {
        return Stack(
          alignment: alignment,
          children: <Widget>[...previousChildren, ?currentChild],
        );
      },
      transitionBuilder: (child, animation) {
        final slideAnimation = Tween<Offset>(
          begin: const Offset(0, 0.018),
          end: Offset.zero,
        ).animate(animation);
        return FadeTransition(
          opacity: animation,
          child: SlideTransition(position: slideAnimation, child: child),
        );
      },
      child: KeyedSubtree(key: ValueKey<Object>(stateKey), child: child),
    );
  }
}
