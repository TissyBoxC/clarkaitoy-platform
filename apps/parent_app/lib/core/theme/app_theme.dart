import 'package:flutter/material.dart';

/// Central Material 3 theme for the parent application.
abstract final class AppTheme {
  static const _seedColor = Color(0xFF2F6B5F);

  static ThemeData get light {
    final colorScheme = ColorScheme.fromSeed(seedColor: _seedColor);
    return ThemeData(
      colorScheme: colorScheme,
      useMaterial3: true,
      visualDensity: VisualDensity.standard,
      cardTheme: const CardThemeData(
        margin: EdgeInsets.zero,
        clipBehavior: Clip.antiAlias,
      ),
      inputDecorationTheme: const InputDecorationTheme(
        border: OutlineInputBorder(),
      ),
    );
  }
}
