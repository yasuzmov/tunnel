/// 侧栏一级分区定义。枚举顺序即侧栏从上到下的顺序，也决定分区切换的滑动方向。
library;

import 'package:flutter/material.dart';

enum AppSection {
  connections('连接管理', Icons.hub_outlined, Icons.hub),
  logs('实时日志', Icons.receipt_long_outlined, Icons.receipt_long),
  about('关于', Icons.info_outline, Icons.info);

  const AppSection(this.label, this.icon, this.activeIcon);

  final String label;
  final IconData icon;
  final IconData activeIcon;
}
