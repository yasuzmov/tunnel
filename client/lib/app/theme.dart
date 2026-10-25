/// 全局主题：单一强调色 + 中性灰，纯色、无渐变。
/// 品牌色与管理后台 #3A6FF7 一致。
library;

import 'package:flutter/material.dart';

import 'transitions.dart';

/// 设计常量集中在此，组件内不要硬编码颜色。
class AppColors {
  AppColors._();

  static const brand = Color(0xFF3A6FF7);
  static const background = Color(0xFFF6F7F9);
  static const surface = Colors.white;
  static const border = Color(0xFFE6E8EE);
  static const text = Color(0xFF1F2329);
  static const text2 = Color(0xFF646A73);
  static const text3 = Color(0xFF8F959E);
  /// 侧栏选中项的底色，比侧栏本身深一档，保证选中态清晰。
  static const sidebarActive = Color(0xFFE5EAF4);
  /// 侧栏悬停底色，比选中态浅、比侧栏深。
  static const sidebarHover = Color(0xFFF0F3F8);
  static const sidebar = Color(0xFFFAFBFC);
  static const success = Color(0xFF16A34A);
  static const warning = Color(0xFFF59E0B);
  static const danger = Color(0xFFDC2626);
}

/// 构建应用主题。
///
/// 返回 Material 3 浅色主题；Windows 默认字体对中文渲染偏细，显式指定雅黑 UI。
ThemeData buildAppTheme() {
  final scheme = ColorScheme.fromSeed(
    seedColor: AppColors.brand,
    primary: AppColors.brand,
    surface: AppColors.surface,
  );
  return ThemeData(
    useMaterial3: true,
    colorScheme: scheme,
    fontFamily: 'Microsoft YaHei UI',
    scaffoldBackgroundColor: AppColors.background,
    pageTransitionsTheme: const PageTransitionsTheme(builders: {
      TargetPlatform.windows: SmoothSlidePageTransitionsBuilder(),
      TargetPlatform.linux: SmoothSlidePageTransitionsBuilder(),
      TargetPlatform.macOS: SmoothSlidePageTransitionsBuilder(),
    }),
    cardTheme: CardThemeData(
      color: AppColors.surface,
      elevation: 0,
      margin: EdgeInsets.zero,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(14),
        side: const BorderSide(color: AppColors.border),
      ),
    ),
    inputDecorationTheme: _inputTheme(),
    filledButtonTheme: FilledButtonThemeData(style: _buttonStyle()),
    outlinedButtonTheme: OutlinedButtonThemeData(style: _buttonStyle()),
    dividerTheme: const DividerThemeData(color: AppColors.border, space: 1),
    tooltipTheme: TooltipThemeData(
      decoration: BoxDecoration(color: AppColors.text, borderRadius: BorderRadius.circular(6)),
      textStyle: const TextStyle(color: Colors.white, fontSize: 12),
    ),
    snackBarTheme: SnackBarThemeData(
      behavior: SnackBarBehavior.floating,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
    ),
  );
}

InputDecorationTheme _inputTheme() {
  OutlineInputBorder border(Color c, [double w = 1]) => OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: c, width: w),
      );
  return InputDecorationTheme(
    filled: true,
    fillColor: const Color(0xFFF8F9FC),
    isDense: true,
    contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 14),
    border: border(AppColors.border),
    enabledBorder: border(AppColors.border),
    focusedBorder: border(AppColors.brand, 1.6),
    errorBorder: border(AppColors.danger),
    labelStyle: const TextStyle(color: AppColors.text2),
    hintStyle: const TextStyle(color: AppColors.text3),
  );
}

ButtonStyle _buttonStyle() {
  return ButtonStyle(
    padding: const WidgetStatePropertyAll(EdgeInsets.symmetric(horizontal: 18, vertical: 14)),
    shape: WidgetStatePropertyAll(RoundedRectangleBorder(borderRadius: BorderRadius.circular(10))),
  );
}
