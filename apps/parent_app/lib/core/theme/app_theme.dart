import 'package:flutter/material.dart';

import 'app_motion.dart';

/// Central Material 3 theme for the 如此萌屋 parent application.
///
/// Pink-white surfaces stay readable through deep plum text, restrained
/// yellow/blue accents, and consistent 16 to 28 pixel radii.
abstract final class AppTheme {
  static const _primary = Color(0xFFF7A8BF);
  static const _primaryStrong = Color(0xFFD94F83);
  static const _surface = Color(0xFFFFFBFC);
  static const _surfaceMuted = Color(0xFFFFF2F5);
  static const _text = Color(0xFF4A2E3B);
  static const _textMuted = Color(0xFF6B4F5A);
  static const _outline = Color(0xFFF0BDCB);

  static ThemeData get light {
    final colorScheme = ColorScheme.fromSeed(
      seedColor: _primary,
      brightness: Brightness.light,
      primary: _primaryStrong,
      onPrimary: Colors.white,
      surface: _surface,
      onSurface: _text,
      outline: _outline,
      error: const Color(0xFFB3261E),
    );
    return ThemeData(
      colorScheme: colorScheme,
      useMaterial3: true,
      scaffoldBackgroundColor: _surface,
      pageTransitionsTheme: AppMotion.pageTransitionsTheme,
      splashFactory: InkSparkle.splashFactory,
      visualDensity: VisualDensity.standard,
      appBarTheme: const AppBarTheme(
        backgroundColor: _surface,
        foregroundColor: _text,
        elevation: 0,
        scrolledUnderElevation: 0,
        centerTitle: false,
      ),
      cardTheme: CardThemeData(
        margin: EdgeInsets.zero,
        color: Colors.white,
        elevation: 0,
        clipBehavior: Clip.antiAlias,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
          side: const BorderSide(color: _outline),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: _surfaceMuted,
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(18),
          borderSide: const BorderSide(color: _outline),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(18),
          borderSide: const BorderSide(color: _outline),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(18),
          borderSide: const BorderSide(color: _primaryStrong, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(18),
          borderSide: BorderSide(color: colorScheme.error),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 18,
          vertical: 16,
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: _primaryStrong,
          foregroundColor: Colors.white,
          minimumSize: const Size.fromHeight(52),
          animationDuration: AppMotion.fast,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(26),
          ),
          textStyle: const TextStyle(fontWeight: FontWeight.w600),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: _primaryStrong,
          minimumSize: const Size.fromHeight(52),
          animationDuration: AppMotion.fast,
          side: const BorderSide(color: _outline),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(26),
          ),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: _primaryStrong,
          animationDuration: AppMotion.fast,
        ),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: Colors.white,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(26),
          side: const BorderSide(color: _outline),
        ),
        titleTextStyle: const TextStyle(
          color: _text,
          fontSize: 20,
          fontWeight: FontWeight.w700,
        ),
        contentTextStyle: const TextStyle(color: _textMuted, height: 1.45),
      ),
      bottomSheetTheme: const BottomSheetThemeData(
        backgroundColor: Colors.white,
        surfaceTintColor: Colors.transparent,
        showDragHandle: true,
        dragHandleColor: _outline,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
        ),
      ),
      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        backgroundColor: _text,
        contentTextStyle: const TextStyle(color: Colors.white),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(18)),
      ),
      listTileTheme: const ListTileThemeData(
        iconColor: _primaryStrong,
        textColor: _text,
      ),
      checkboxTheme: CheckboxThemeData(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
        fillColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return _primaryStrong;
          }
          return Colors.transparent;
        }),
      ),
      progressIndicatorTheme: const ProgressIndicatorThemeData(
        color: _primaryStrong,
      ),
      textTheme: const TextTheme(
        bodyMedium: TextStyle(color: _textMuted, height: 1.45),
        bodySmall: TextStyle(color: _textMuted),
      ),
    );
  }
}
