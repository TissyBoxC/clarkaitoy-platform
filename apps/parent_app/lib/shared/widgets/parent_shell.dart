import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../core/theme/app_motion.dart';

/// Persistent bottom navigation shared by the signed-in guardian pages.
class ParentShell extends StatelessWidget {
  const ParentShell({required this.navigationShell, super.key});

  final StatefulNavigationShell navigationShell;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: DecoratedBox(
        decoration: BoxDecoration(
          color: Colors.white,
          border: Border(
            top: BorderSide(
              color: Theme.of(context).colorScheme.outlineVariant,
            ),
          ),
          boxShadow: const [
            BoxShadow(
              color: Color(0x1AF7A8BF),
              blurRadius: 24,
              offset: Offset(0, -6),
            ),
          ],
        ),
        child: SafeArea(
          top: false,
          child: NavigationBar(
            height: 68,
            animationDuration: AppMotion.standard,
            selectedIndex: navigationShell.currentIndex,
            onDestinationSelected: (index) {
              if (index == navigationShell.currentIndex) {
                navigationShell.goBranch(index);
                return;
              }
              navigationShell.goBranch(index);
            },
            destinations: const [
              NavigationDestination(
                icon: Icon(Icons.home_outlined),
                selectedIcon: Icon(Icons.home_rounded),
                label: '首页',
              ),
              NavigationDestination(
                icon: Icon(Icons.toys_outlined),
                selectedIcon: Icon(Icons.toys_rounded),
                label: '设备',
              ),
              NavigationDestination(
                icon: Icon(Icons.person_outline_rounded),
                selectedIcon: Icon(Icons.person_rounded),
                label: '我的',
              ),
            ],
          ),
        ),
      ),
    );
  }
}
