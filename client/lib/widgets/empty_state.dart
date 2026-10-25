/// 空状态占位：图标 + 标题 + 说明 + 可选操作按钮。
library;

import 'package:flutter/material.dart';

import '../app/theme.dart';

/// 列表为空时的占位组件。
class EmptyState extends StatelessWidget {
  const EmptyState({
    super.key,
    required this.icon,
    required this.title,
    this.description,
    this.action,
  });

  final IconData icon;
  final String title;
  final String? description;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 72,
              height: 72,
              decoration: BoxDecoration(
                color: AppColors.brand.withValues(alpha: 0.08),
                borderRadius: BorderRadius.circular(22),
              ),
              child: Icon(icon, size: 34, color: AppColors.brand),
            ),
            const SizedBox(height: 18),
            Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: AppColors.text)),
            if (description != null) ...[
              const SizedBox(height: 6),
              Text(description!, textAlign: TextAlign.center, style: const TextStyle(fontSize: 13, color: AppColors.text3)),
            ],
            if (action != null) ...[const SizedBox(height: 20), action!],
          ],
        ),
      ),
    );
  }
}
