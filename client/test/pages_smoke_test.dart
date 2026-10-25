/// 页面冒烟测试：在桌面常见尺寸下渲染各页面，确保没有布局溢出与构建异常。
/// AppShell 依赖 window_manager 原生通道，不在此覆盖。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:tunnel_client/app/theme.dart';
import 'package:tunnel_client/models/profile.dart';
import 'package:tunnel_client/models/rule.dart';
import 'package:tunnel_client/pages/about/about_page.dart';
import 'package:tunnel_client/pages/connections/connections_page.dart';
import 'package:tunnel_client/pages/logs/logs_page.dart';
import 'package:tunnel_client/pages/profile_edit/profile_edit_page.dart';
import 'package:tunnel_client/pages/rules/rule_editor_page.dart';
import 'package:tunnel_client/pages/rules/rules_page.dart';
import 'package:tunnel_client/providers/app_providers.dart';
import 'package:tunnel_client/services/config_store.dart';

/// 内存版配置仓库，避免测试触达 path_provider 平台通道与真实文件。
class _MemStore extends ConfigStore {
  _MemStore(this.items);
  final List<ConnectionProfile> items;

  @override
  Future<List<ConnectionProfile>> load() async => items;

  @override
  Future<void> save(List<ConnectionProfile> profiles) async {}

  @override
  Future<String> filePath() async => r'C:\fake\profiles.json';
}

ConnectionProfile _sample() {
  return ConnectionProfile(
    id: 'p1',
    name: '办公室 NAS',
    serverUrl: 'https://tunnel.example.com',
    key: 'abcdef1234567890',
    rules: [
      ForwardRule(id: 'r1', matchHost: 'nas.example.com', pathPrefix: '/api', port: 5000, useHttps: true),
      ForwardRule(id: 'r2', targetType: 'hostsDomain', hostsDomain: 'dev.local', port: 8080, rewriteHost: true),
    ],
  );
}

Future<void> _pump(WidgetTester tester, Widget page, Size size) async {
  tester.view.physicalSize = size;
  tester.view.devicePixelRatio = 1;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(ProviderScope(
    overrides: [configStoreProvider.overrideWithValue(_MemStore([_sample()]))],
    child: MaterialApp(theme: buildAppTheme(), home: Scaffold(body: page)),
  ));
  await tester.pumpAndSettle();
}

void main() {
  final pages = <String, Widget>{
    'connections': const ConnectionsPage(),
    'profile edit': ProfileEditPage(profile: _sample()),
    'rules': const RulesPage(profileId: 'p1'),
    'rule editor': RuleEditorPage(profileId: 'p1', rule: _sample().rules[1]),
    'logs': const LogsPage(),
    'about': const AboutPage(),
  };
  // 宽度对应最小窗口 960 与默认窗口 1160 减去 220 的侧栏。
  for (final size in const [Size(740, 600), Size(940, 720)]) {
    for (final entry in pages.entries) {
      testWidgets('${entry.key} renders at ${size.width.toInt()}x${size.height.toInt()}', (tester) async {
        await _pump(tester, entry.value, size);
        expect(tester.takeException(), isNull);
      });
    }
  }
}
