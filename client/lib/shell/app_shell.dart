/// 应用外壳：左侧深色侧栏 + 右侧（自定义标题栏 + 分区内容）。
/// 也负责启动时按配置自动连接。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/profile.dart';
import '../pages/about/about_page.dart';
import '../pages/connections/connections_page.dart';
import '../pages/logs/logs_page.dart';
import '../providers/app_providers.dart';
import 'app_section.dart';
import 'section_navigator.dart';
import 'section_switcher.dart';
import 'sidebar.dart';
import 'title_bar.dart';

/// 顶层布局。
class AppShell extends ConsumerStatefulWidget {
  const AppShell({super.key});

  @override
  ConsumerState<AppShell> createState() => _AppShellState();
}

class _AppShellState extends ConsumerState<AppShell> {
  final _connectionsNav = GlobalKey<NavigatorState>();
  AppSection _section = AppSection.connections;
  bool _autoConnectDone = false;

  @override
  void initState() {
    super.initState();
    ref.listenManual<List<ConnectionProfile>>(profilesProvider, (_, next) => _maybeAutoConnect(next));
  }

  /// 首次状态变化即配置文件加载完成（哪怕为空列表，StateNotifier 也按实例比较会通知），
  /// 只在这一次连接第一个勾选了自动连接的配置，之后新增配置不触发。
  void _maybeAutoConnect(List<ConnectionProfile> profiles) {
    if (_autoConnectDone) return;
    _autoConnectDone = true;
    for (final p in profiles) {
      if (p.autoConnect) {
        ref.read(tunnelProvider.notifier).connect(p);
        return;
      }
    }
  }

  /// 重复点击「连接管理」时回到列表根页，符合常见桌面应用习惯。
  void _select(AppSection s) {
    if (s == _section) {
      if (s == AppSection.connections) {
        _connectionsNav.currentState?.popUntil((r) => r.isFirst);
      }
      return;
    }
    setState(() => _section = s);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          Sidebar(current: _section, onSelect: _select),
          Expanded(
            child: Column(
              children: [
                TitleBar(title: _section.label),
                Expanded(child: ClipRect(child: _content())),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _content() {
    return SectionSwitcher(
      index: _section.index,
      children: [
        SectionNavigator(navigatorKey: _connectionsNav, root: const ConnectionsPage()),
        const LogsPage(),
        const AboutPage(),
      ],
    );
  }
}
