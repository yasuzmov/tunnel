/// 规则列表中的一行：序号、拖拽把手、匹配条件 → 目标、启用开关与操作。
library;

import 'package:flutter/material.dart';

import '../../../app/theme.dart';
import '../../../models/rule.dart';
import '../../../widgets/badges.dart';

/// 规则行。[index] 既是显示序号也是 ReorderableDragStartListener 的下标。
class RuleTile extends StatelessWidget {
  const RuleTile({
    super.key,
    required this.index,
    required this.rule,
    required this.onToggle,
    required this.onEdit,
    required this.onDelete,
  });

  final int index;
  final ForwardRule rule;
  final ValueChanged<bool> onToggle;
  final VoidCallback onEdit;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Material(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(12),
        child: InkWell(
          borderRadius: BorderRadius.circular(12),
          onTap: onEdit,
          child: AnimatedOpacity(
            duration: const Duration(milliseconds: 250),
            opacity: rule.enabled ? 1 : 0.55,
            child: Container(
              padding: const EdgeInsets.fromLTRB(8, 12, 8, 12),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                border: Border.all(color: AppColors.border),
              ),
              child: Row(children: [_handle(), const SizedBox(width: 6), Expanded(child: _body()), ..._trailing()]),
            ),
          ),
        ),
      ),
    );
  }

  Widget _handle() {
    return ReorderableDragStartListener(
      index: index,
      child: MouseRegion(
        cursor: SystemMouseCursors.grab,
        child: Row(
          children: [
            const Icon(Icons.drag_indicator, color: AppColors.text3, size: 20),
            Container(
              width: 24,
              height: 24,
              alignment: Alignment.center,
              decoration: BoxDecoration(color: AppColors.brand.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(7)),
              child: Text('${index + 1}', style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: AppColors.brand)),
            ),
          ],
        ),
      ),
    );
  }

  Widget _body() {
    final host = rule.matchHost.isEmpty ? '任意域名' : rule.matchHost;
    final path = rule.pathPrefix.isEmpty ? '/' : rule.pathPrefix;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Flexible(
              child: Text('$host$path', maxLines: 1, overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontWeight: FontWeight.w600, color: AppColors.text)),
            ),
            const Padding(
              padding: EdgeInsets.symmetric(horizontal: 8),
              child: Icon(Icons.arrow_forward_rounded, size: 16, color: AppColors.text3),
            ),
            Flexible(
              child: Text(rule.targetUri(), maxLines: 1, overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontFamily: 'Consolas', fontSize: 13, color: AppColors.brand)),
            ),
          ],
        ),
        const SizedBox(height: 6),
        Wrap(spacing: 6, runSpacing: 4, children: _tags()),
      ],
    );
  }

  List<Widget> _tags() {
    return [
      Pill(text: rule.targetType == 'hostsDomain' ? 'hosts 域名' : '本机端口', color: AppColors.text2),
      if (rule.useHttps) const Pill(text: 'HTTPS', color: AppColors.success),
      if (rule.ignoreTlsError) const Pill(text: '忽略证书', color: AppColors.warning),
      if (rule.rewriteHost) Pill(text: 'Host→${rule.rewriteHostValue.isEmpty ? '目标' : rule.rewriteHostValue}', color: AppColors.brand),
    ];
  }

  List<Widget> _trailing() {
    return [
      Tooltip(
        message: rule.enabled ? '已启用' : '已停用',
        child: Switch(value: rule.enabled, onChanged: onToggle),
      ),
      IconButton(tooltip: '编辑', onPressed: onEdit, icon: const Icon(Icons.edit_outlined, size: 19)),
      IconButton(
        tooltip: '删除',
        onPressed: onDelete,
        icon: const Icon(Icons.delete_outline, size: 19, color: AppColors.danger),
      ),
    ];
  }
}
