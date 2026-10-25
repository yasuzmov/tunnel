/// 某个连接的转发规则列表：按顺序匹配、命中即停，支持拖拽调整优先级。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/transitions.dart';
import '../../models/profile.dart';
import '../../models/rule.dart';
import '../../providers/app_providers.dart';
import '../../widgets/empty_state.dart';
import '../../widgets/fade_slide_in.dart';
import '../../widgets/page_header.dart';
import 'rule_editor_page.dart';
import 'widgets/rule_tile.dart';

/// 规则列表页。按 [profileId] 从 provider 读取，保证与其他页面的修改同步。
class RulesPage extends ConsumerWidget {
  const RulesPage({super.key, required this.profileId});

  final String profileId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profile = ref.watch(profilesProvider.select((list) {
      for (final p in list) {
        if (p.id == profileId) return p;
      }
      return null;
    }));
    return ColoredBox(
      color: AppColors.background,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          PageHeader(
            title: '转发规则',
            subtitle: profile == null ? '配置已删除' : '${profile.name} · 自上而下匹配，命中即停，拖动左侧把手调整优先级',
            showBack: true,
            actions: [
              if (profile != null)
                FilledButton.icon(
                  onPressed: () => pushSlide(context, RuleEditorPage(profileId: profileId)),
                  icon: const Icon(Icons.add, size: 18),
                  label: const Text('新增规则'),
                ),
            ],
          ),
          Expanded(child: profile == null || profile.rules.isEmpty ? _empty(context, profile) : _list(context, ref, profile)),
        ],
      ),
    );
  }

  Widget _empty(BuildContext context, ConnectionProfile? profile) {
    return EmptyState(
      icon: Icons.alt_route,
      title: profile == null ? '找不到该连接配置' : '还没有转发规则',
      description: profile == null ? null : '没有规则时所有请求都会返回 502，添加一条把请求转到本地服务',
      action: profile == null
          ? null
          : FilledButton.icon(
              onPressed: () => pushSlide(context, RuleEditorPage(profileId: profileId)),
              icon: const Icon(Icons.add, size: 18),
              label: const Text('新增规则'),
            ),
    );
  }

  Widget _list(BuildContext context, WidgetRef ref, ConnectionProfile profile) {
    final notifier = ref.read(profilesProvider.notifier);
    final rules = profile.rules;
    return ReorderableListView.builder(
      padding: const EdgeInsets.fromLTRB(28, 4, 28, 28),
      buildDefaultDragHandles: false,
      proxyDecorator: _proxyDecorator,
      itemCount: rules.length,
      onReorderItem: (from, to) {
        final next = [...rules];
        next.insert(to, next.removeAt(from));
        notifier.updateRules(profileId, next);
      },
      itemBuilder: (context, i) {
        final rule = rules[i];
        return FadeSlideIn(
          key: ValueKey(rule.id),
          index: i,
          child: RuleTile(
            index: i,
            rule: rule,
            onToggle: (v) => notifier.updateRules(profileId, _replace(rules, i, rule.copy()..enabled = v)),
            onEdit: () => pushSlide(context, RuleEditorPage(profileId: profileId, rule: rule)),
            onDelete: () => _confirmDelete(context, notifier, rules, i),
          ),
        );
      },
    );
  }

  List<ForwardRule> _replace(List<ForwardRule> rules, int i, ForwardRule rule) => [...rules]..[i] = rule;

  /// 拖拽中的条目放大并加阴影，给出「被拿起」的反馈。
  Widget _proxyDecorator(Widget child, int index, Animation<double> anim) {
    return AnimatedBuilder(
      animation: anim,
      builder: (context, child) {
        final t = Curves.easeOutCubic.transform(anim.value);
        return Transform.scale(
          scale: 1 + 0.02 * t,
          child: Material(
            color: Colors.transparent,
            elevation: 12 * t,
            shadowColor: Colors.black26,
            borderRadius: BorderRadius.circular(12),
            child: child,
          ),
        );
      },
      child: child,
    );
  }

  Future<void> _confirmDelete(BuildContext context, ProfilesNotifier notifier, List<ForwardRule> rules, int i) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('删除规则'),
        content: Text('确定删除规则「${rules[i].label}」？'),
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
    if (ok == true) await notifier.updateRules(profileId, [...rules]..removeAt(i));
  }
}
