/// 配置列表首页：连接、编辑、删除。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/tunnel/client.dart';
import '../models/profile.dart';
import '../providers/app_providers.dart';
import 'edit_profile_page.dart';
import 'logs_page.dart';
import 'rules_page.dart';

class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profiles = ref.watch(profilesProvider);
    final tunnel = ref.watch(tunnelProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('内网穿透客户端'),
        actions: [
          IconButton(
            tooltip: '实时日志',
            onPressed: () => Navigator.push(context, MaterialPageRoute(builder: (_) => const LogsPage())),
            icon: const Icon(Icons.article_outlined),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _edit(context, ConnectionProfile(id: '', name: '新连接', serverUrl: 'http://127.0.0.1:8080', key: '')),
        child: const Icon(Icons.add),
      ),
      body: Column(
        children: [
          _StatusBar(state: tunnel),
          Expanded(
            child: profiles.isEmpty
                ? const Center(child: Text('还没有连接配置，点击右下角添加'))
                : ListView.separated(
                    itemCount: profiles.length,
                    separatorBuilder: (_, _) => const Divider(height: 1),
                    itemBuilder: (context, i) {
                      final p = profiles[i];
                      final active = tunnel.activeId == p.id && tunnel.status == TunnelStatus.online;
                      return ListTile(
                        title: Text(p.name),
                        subtitle: Text(p.serverUrl),
                        leading: Icon(Icons.circle, size: 12, color: active ? Colors.green : Colors.grey),
                        trailing: Wrap(
                          spacing: 8,
                          children: [
                            TextButton(
                              onPressed: () {
                                if (active) {
                                  ref.read(tunnelProvider.notifier).disconnect();
                                } else {
                                  ref.read(tunnelProvider.notifier).connect(p);
                                }
                              },
                              child: Text(active ? '断开' : '连接'),
                            ),
                            TextButton(
                              onPressed: () => Navigator.push(
                                context,
                                MaterialPageRoute(builder: (_) => RulesPage(profile: p)),
                              ),
                              child: const Text('规则'),
                            ),
                            IconButton(
                              onPressed: () => _edit(context, p),
                              icon: const Icon(Icons.edit_outlined),
                            ),
                            IconButton(
                              onPressed: () => ref.read(profilesProvider.notifier).remove(p.id),
                              icon: const Icon(Icons.delete_outline),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  void _edit(BuildContext context, ConnectionProfile profile) {
    Navigator.push(context, MaterialPageRoute(builder: (_) => EditProfilePage(profile: profile)));
  }
}

class _StatusBar extends StatelessWidget {
  const _StatusBar({required this.state});
  final TunnelState state;

  @override
  Widget build(BuildContext context) {
    final welcome = state.welcome;
    return Material(
      color: const Color(0xFF1D1E2C),
      child: ListTile(
        title: Text(state.message.isEmpty ? '未连接' : state.message, style: const TextStyle(color: Colors.white)),
        subtitle: welcome == null
            ? null
            : Text('记录: ${welcome.name}  延迟: ${welcome.pingMs}ms  域名: ${welcome.domains.join(', ')}',
                style: const TextStyle(color: Colors.white70)),
      ),
    );
  }
}
