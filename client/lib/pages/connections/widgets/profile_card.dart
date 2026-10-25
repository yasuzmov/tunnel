/// 单个连接配置卡片：名称、地址、规则数、当前状态与操作按钮。
library;

import 'package:flutter/material.dart';

import '../../../app/theme.dart';
import '../../../models/profile.dart';
import '../../../utils/status_style.dart';
import '../../../widgets/status_dot.dart';

/// 连接配置卡片。
///
/// [style] 为 null 表示该配置不是当前活动连接；
/// [busy] 为 true（连接中 / 重连中 / 已连接）时主按钮显示「断开」。
class ProfileCard extends StatefulWidget {
  const ProfileCard({
    super.key,
    required this.profile,
    required this.style,
    required this.busy,
    required this.onToggle,
    required this.onRules,
    required this.onEdit,
    required this.onDelete,
  });

  final ConnectionProfile profile;
  final StatusStyle? style;
  final bool busy;
  final VoidCallback onToggle;
  final VoidCallback onRules;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  @override
  State<ProfileCard> createState() => _ProfileCardState();
}

class _ProfileCardState extends State<ProfileCard> {
  bool _hover = false;

  @override
  Widget build(BuildContext context) {
    final active = widget.style != null;
    return MouseRegion(
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 260),
        curve: Curves.easeOutCubic,
        transform: Matrix4.translationValues(0, _hover ? -3 : 0, 0),
        padding: const EdgeInsets.all(18),
        decoration: BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: active ? AppColors.brand.withValues(alpha: 0.5) : AppColors.border, width: active ? 1.5 : 1),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFF101828).withValues(alpha: _hover ? 0.1 : 0.04),
              blurRadius: _hover ? 24 : 10,
              offset: Offset(0, _hover ? 10 : 3),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _head(),
            const SizedBox(height: 14),
            _meta(),
            const SizedBox(height: 16),
            _actions(),
          ],
        ),
      ),
    );
  }

  Widget _head() {
    final p = widget.profile;
    final style = widget.style;
    return Row(
      children: [
        Container(
          width: 40,
          height: 40,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: AppColors.brand.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Text(
            p.name.isEmpty ? '?' : p.name.characters.first,
            style: const TextStyle(color: AppColors.brand, fontWeight: FontWeight.w700, fontSize: 16),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(p.name, maxLines: 1, overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: AppColors.text)),
              const SizedBox(height: 2),
              Text(p.serverUrl, maxLines: 1, overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 12, color: AppColors.text3)),
            ],
          ),
        ),
        if (style != null) ...[
          StatusDot(color: style.color, pulse: style.pulse, size: 8),
          Text(style.label, style: TextStyle(fontSize: 12, color: style.color, fontWeight: FontWeight.w600)),
        ],
        _menu(),
      ],
    );
  }

  Widget _meta() {
    final p = widget.profile;
    final enabled = p.rules.where((r) => r.enabled).length;
    return Wrap(
      spacing: 16,
      runSpacing: 6,
      children: [
        _metaItem(Icons.alt_route, '规则 $enabled/${p.rules.length}'),
        _metaItem(Icons.key_outlined, p.key.isEmpty ? '未填写 Key' : 'Key ····${_tail(p.key)}'),
        if (p.autoConnect) _metaItem(Icons.bolt_outlined, '启动自动连接'),
      ],
    );
  }

  String _tail(String key) => key.length <= 4 ? key : key.substring(key.length - 4);

  Widget _metaItem(IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: AppColors.text3),
        const SizedBox(width: 4),
        Text(text, style: const TextStyle(fontSize: 12, color: AppColors.text2)),
      ],
    );
  }

  Widget _actions() {
    return Row(
      children: [
        Expanded(
          child: widget.busy
              ? FilledButton.tonalIcon(
                  onPressed: widget.onToggle,
                  icon: const Icon(Icons.stop_circle_outlined, size: 18),
                  label: const Text('断开'),
                )
              : FilledButton.icon(
                  onPressed: widget.onToggle,
                  icon: const Icon(Icons.play_arrow_rounded, size: 18),
                  label: const Text('连接'),
                ),
        ),
        const SizedBox(width: 10),
        OutlinedButton.icon(
          onPressed: widget.onRules,
          icon: const Icon(Icons.alt_route, size: 16),
          label: const Text('转发规则'),
        ),
      ],
    );
  }

  Widget _menu() {
    return PopupMenuButton<String>(
      tooltip: '更多',
      icon: const Icon(Icons.more_horiz, color: AppColors.text3),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
      onSelected: (v) => v == 'edit' ? widget.onEdit() : widget.onDelete(),
      itemBuilder: (_) => const [
        PopupMenuItem(value: 'edit', child: ListTile(leading: Icon(Icons.edit_outlined), title: Text('编辑'), dense: true)),
        PopupMenuItem(
          value: 'delete',
          child: ListTile(
            leading: Icon(Icons.delete_outline, color: AppColors.danger),
            title: Text('删除', style: TextStyle(color: AppColors.danger)),
            dense: true,
          ),
        ),
      ],
    );
  }
}
