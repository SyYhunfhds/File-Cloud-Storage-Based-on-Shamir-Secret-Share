import 'package:flutter/material.dart';

/// Material Design 3 主题配置
class AppTheme {
  static const _seedColor = Color(0xFF1E6FFF); // 企业蓝色

  /// 亮色主题
  static ThemeData get light => ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: _seedColor,
          brightness: Brightness.light,
        ),
        visualDensity: VisualDensity.compact,
      );

  /// 暗色主题
  static ThemeData get dark => ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: _seedColor,
          brightness: Brightness.dark,
        ),
        visualDensity: VisualDensity.compact,
      );
}
