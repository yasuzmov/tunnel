/// 连接页顶部状态横幅：当前隧道状态、记录名、延迟与已绑定域名。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../app/theme.dart';
import '../../../core/tunnel/client.dart';
import '../../../providers/app_providers.dart';
import '../../../utils/status_style.dart';
import '../../../widgets/status_dot.dart';

/// 状态横幅。状态切换时内容区淡入淡出并平滑改变高度。
class StatusHero extends ConsumerWidget {
  const StatusHero({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(tunnelProvider);
    final style = statusStyleOf(state.status);
    final idle = state.status == TunnelStatus.idle;
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppColors.border),
      ),
      child: AnimatedSize(
        duration: const Duration(milliseconds: 350),
        curve: Curves.easeOutCubic,
        alignment: Alignment.topCenter,
        child: AnimatedSwitcher(
          duration: const Duration(milliseconds: 300),
          child: KeyedSubtree(
            key: ValueKey('${state.status}-${state.welcome?.name}-${state.welcome?.domains.length}'),
            child: Row(
              children: [
                Expanded(child: _info(state, style)),
                if (!idle)
                  OutlinedButton.icon(
                    onPressed: () => ref.read(tunnelProvider.notifier).disconnect(),
                    icon: const Icon(Icons.power_settings_new, size: 18),
                    label: const Text('断开'),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _info(TunnelState state, StatusStyle style) {
    final welcome = state.welcome;
    final message = state.message.isEmpty ? '在下方选择一个连接配置并点击「连接」' : state.message;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            StatusDot(color: style.color, pulse: style.pulse),
            const SizedBox(width: 6),
            Text(style.label, style: const TextStyle(color: AppColors.text, fontSize: 18, fontWeight: FontWeight.w700)),
            if (welcome != null) ...[
              const SizedBox(width: 12),
              _chip(Icons.speed, '${welcome.pingMs} ms'),
            ],
          ],
        ),
        const SizedBox(height: 8),
        Text(
          welcome != null ? '记录：${welcome.name}' : message,
          style: const TextStyle(color: AppColors.text2, fontSize: 13),
        ),
        if (welcome != null && welcome.domains.isNotEmpty) ...[
          const SizedBox(height: 12),
          Wrap(spacing: 8, runSpacing: 8, children: [for (final d in welcome.domains) _chip(Icons.language, d)]),
        ],
      ],
    );
  }

  Widget _chip(IconData icon, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: AppColors.background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 13, color: AppColors.text2),
          const SizedBox(width: 4),
          Text(text, style: const TextStyle(color: AppColors.text2, fontSize: 12)),
        ],
      ),
    );
  }
}
