/// 隧道状态到展示文案 / 颜色 / 是否呼吸动画的映射，侧栏与连接页共用。
library;

import 'package:flutter/material.dart';

import '../app/theme.dart';
import '../core/tunnel/client.dart';

/// 状态展示样式。
class StatusStyle {
  const StatusStyle(this.label, this.color, {this.pulse = false});

  final String label;
  final Color color;

  /// 是否处于进行中或在线，需要呼吸动画提示。
  final bool pulse;
}

/// 根据 [status] 返回展示样式。
StatusStyle statusStyleOf(TunnelStatus status) {
  switch (status) {
    case TunnelStatus.online:
      return const StatusStyle('已连接', AppColors.success, pulse: true);
    case TunnelStatus.connecting:
      return const StatusStyle('连接中', AppColors.brand, pulse: true);
    case TunnelStatus.reconnecting:
      return const StatusStyle('重连中', AppColors.warning, pulse: true);
    case TunnelStatus.error:
      return const StatusStyle('连接失败', AppColors.danger);
    case TunnelStatus.idle:
      return const StatusStyle('未连接', AppColors.text3);
  }
}

/// 把时间格式化为 HH:mm:ss，日志列表使用。
String formatClock(DateTime t) {
  String two(int v) => v.toString().padLeft(2, '0');
  return '${two(t.hour)}:${two(t.minute)}:${two(t.second)}';
}
