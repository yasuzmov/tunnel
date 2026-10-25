/// 连接管理首页（分区根页面）：状态横幅 + 连接配置卡片网格。
/// 子页面（编辑、规则）在分区内嵌 Navigator 中滑入。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/transitions.dart';
import '../../core/tunnel/client.dart';
import '../../models/profile.dart';
import '../../providers/app_providers.dart';
import '../../utils/status_style.dart';
import '../../widgets/empty_state.dart';
import '../../widgets/fade_slide_in.dart';
import '../../widgets/page_header.dart';
import '../profile_edit/profile_edit_page.dart';
import '../rules/rules_page.dart';
import 'widgets/profile_card.dart';
import 'widgets/status_hero.dart';

const double _cardMinWidth = 340;
const double _gap = 16;

class ConnectionsPage extends ConsumerWidget {
  const ConnectionsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profiles = ref.watch(profilesProvider);
    return ColoredBox(
      color: AppColors.background,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          PageHeader(
            title: '连接管理',
            subtitle: '配置服务端地址与接入 Key，连接后公网请求将转发到本机',
            actions: [
              FilledButton.icon(
                onPressed: () => _openEditor(context, null),
                icon: const Icon(Icons.add, size: 18),
                label: const Text('新建连接'),
              ),
            ],
          ),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.fromLTRB(28, 4, 28, 28),
              children: [
                const FadeSlideIn(child: StatusHero()),
                const SizedBox(height: 24),
                if (profiles.isEmpty) _empty(context) else _grid(context, ref, profiles),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _empty(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 40),
      child: EmptyState(
        icon: Icons.hub_outlined,
        title: '还没有连接配置',
        description: '在管理后台创建穿透记录后，把接入地址和 Key 填到这里',
        action: FilledButton.icon(
          onPressed: () => _openEditor(context, null),
          icon: const Icon(Icons.add, size: 18),
          label: const Text('新建连接'),
        ),
      ),
    );
  }

  /// 按可用宽度自适应列数，卡片等宽排列。
  Widget _grid(BuildContext context, WidgetRef ref, List<ConnectionProfile> profiles) {
    final tunnel = ref.watch(tunnelProvider);
    return LayoutBuilder(builder: (context, box) {
      final cols = ((box.maxWidth + _gap) / (_cardMinWidth + _gap)).floor().clamp(1, 4);
      final width = (box.maxWidth - _gap * (cols - 1)) / cols;
      return Wrap(
        spacing: _gap,
        runSpacing: _gap,
        children: [
          for (var i = 0; i < profiles.length; i++)
            SizedBox(
              key: ValueKey(profiles[i].id),
              width: width,
              child: FadeSlideIn(index: i, child: _card(context, ref, profiles[i], tunnel)),
            ),
        ],
      );
    });
  }

  Widget _card(BuildContext context, WidgetRef ref, ConnectionProfile p, TunnelState tunnel) {
    final isActive = tunnel.activeId == p.id && tunnel.status != TunnelStatus.idle;
    final busy = isActive && tunnel.status != TunnelStatus.error;
    final ctrl = ref.read(tunnelProvider.notifier);
    return ProfileCard(
      profile: p,
      style: isActive ? statusStyleOf(tunnel.status) : null,
      busy: busy,
      onToggle: () => busy ? ctrl.disconnect() : ctrl.connect(p),
      onRules: () => pushSlide(context, RulesPage(profileId: p.id)),
      onEdit: () => _openEditor(context, p),
      onDelete: () => _confirmDelete(context, ref, p),
    );
  }

  void _openEditor(BuildContext context, ConnectionProfile? profile) {
    pushSlide(context, ProfileEditPage(profile: profile));
  }

  Future<void> _confirmDelete(BuildContext context, WidgetRef ref, ConnectionProfile p) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('删除连接'),
        content: Text('确定删除「${p.name}」及其全部转发规则？'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('取消')),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: AppColors.danger),
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('删除'),
          ),
        ],
      ),
    );
    if (ok != true) return;
    if (ref.read(tunnelProvider).activeId == p.id) {
      await ref.read(tunnelProvider.notifier).disconnect();
    }
    await ref.read(profilesProvider.notifier).remove(p.id);
  }
}
