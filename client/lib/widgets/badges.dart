/// 小型标签：HTTP 方法、状态码与通用胶囊标签。配色与管理后台日志列表保持一致。
library;

import 'package:flutter/material.dart';

import '../app/theme.dart';

/// 通用胶囊标签：浅色底 + 同色文字。
class Pill extends StatelessWidget {
  const Pill({super.key, required this.text, required this.color, this.mono = false});

  final String text;
  final Color color;

  /// 是否使用等宽字体（方法、状态码等对齐展示）。
  final bool mono;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 11.5,
          fontWeight: FontWeight.w600,
          color: color,
          fontFamily: mono ? 'Consolas' : null,
        ),
      ),
    );
  }
}

/// HTTP 方法对应的颜色；未知方法为灰色。
Color methodColor(String method) {
  switch (method.toUpperCase()) {
    case 'GET':
      return AppColors.success;
    case 'POST':
      return AppColors.brand;
    case 'PUT':
    case 'PATCH':
      return AppColors.warning;
    case 'DELETE':
      return AppColors.danger;
    default:
      return AppColors.text2;
  }
}

/// 状态码颜色：0（未拿到响应）与 5xx 红、4xx 橙、3xx 灰、2xx 绿。
Color statusColor(int status) {
  if (status == 0 || status >= 500) return AppColors.danger;
  if (status >= 400) return AppColors.warning;
  if (status >= 300) return AppColors.text2;
  return AppColors.success;
}
