/// 分区内嵌导航：子页面（编辑配置、转发规则等）只在内容区内滑动，侧栏保持不动。
library;

import 'package:flutter/material.dart';

/// 包一层独立 Navigator。
///
/// [navigatorKey] 供外层在重复点击侧栏时 popUntil 回到根页；
/// [root] 为该分区的首页。外层需 ClipRect，避免转场时页面绘制到侧栏上。
class SectionNavigator extends StatelessWidget {
  const SectionNavigator({super.key, required this.navigatorKey, required this.root});

  final GlobalKey<NavigatorState> navigatorKey;
  final Widget root;

  @override
  Widget build(BuildContext context) {
    return Navigator(
      key: navigatorKey,
      onGenerateRoute: (settings) => MaterialPageRoute(settings: settings, builder: (_) => root),
    );
  }
}
