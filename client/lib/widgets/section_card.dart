/// 带标题的白色分组卡片，用于表单与详情的分块展示。
library;

import 'package:flutter/material.dart';

import '../app/theme.dart';

/// 分组卡片。
///
/// [title] 为空时不渲染标题行；[trailing] 显示在标题右侧；
/// [padding] 默认 20，列表类内容可传 [EdgeInsets.zero] 让条目贴边。
class SectionCard extends StatelessWidget {
  const SectionCard({
    super.key,
    this.title,
    this.icon,
    this.trailing,
    this.padding = const EdgeInsets.all(20),
    required this.child,
  });

  final String? title;
  final IconData? icon;
  final Widget? trailing;
  final EdgeInsetsGeometry padding;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        mainAxisSize: MainAxisSize.min,
        children: [
          if (title != null) _header(),
          Padding(padding: padding, child: child),
        ],
      ),
    );
  }

  Widget _header() {
    return Container(
      padding: const EdgeInsets.fromLTRB(20, 16, 16, 12),
      decoration: const BoxDecoration(border: Border(bottom: BorderSide(color: AppColors.border))),
      child: Row(
        children: [
          if (icon != null) ...[
            Icon(icon, size: 18, color: AppColors.brand),
            const SizedBox(width: 8),
          ],
          Expanded(
            child: Text(title!, style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: AppColors.text)),
          ),
          ?trailing,
        ],
      ),
    );
  }
}
