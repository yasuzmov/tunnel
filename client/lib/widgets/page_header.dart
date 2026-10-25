/// 页面顶部标题区：可选返回按钮、标题、副标题与右侧操作。
library;

import 'package:flutter/material.dart';

import '../app/theme.dart';

/// 统一的页头。
///
/// [showBack] 为 true 时显示返回按钮，点击 pop 当前 Navigator；
/// [actions] 放在右侧，通常是主操作按钮。
class PageHeader extends StatelessWidget {
  const PageHeader({
    super.key,
    required this.title,
    this.subtitle,
    this.showBack = false,
    this.actions = const [],
  });

  final String title;
  final String? subtitle;
  final bool showBack;
  final List<Widget> actions;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(28, 24, 28, 16),
      child: Row(
        children: [
          if (showBack) ...[
            _BackButton(onTap: () => Navigator.of(context).maybePop()),
            const SizedBox(width: 14),
          ],
          Expanded(child: _titles()),
          ...actions.expand((w) => [const SizedBox(width: 10), w]),
        ],
      ),
    );
  }

  Widget _titles() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(title, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w700, color: AppColors.text)),
        if (subtitle != null) ...[
          const SizedBox(height: 4),
          Text(subtitle!, style: const TextStyle(fontSize: 13, color: AppColors.text3)),
        ],
      ],
    );
  }
}

/// 带悬停位移反馈的返回按钮。
class _BackButton extends StatefulWidget {
  const _BackButton({required this.onTap});
  final VoidCallback onTap;

  @override
  State<_BackButton> createState() => _BackButtonState();
}

class _BackButtonState extends State<_BackButton> {
  bool _hover = false;

  @override
  Widget build(BuildContext context) {
    final color = _hover ? AppColors.brand : AppColors.text2;
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: GestureDetector(
        onTap: widget.onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOutCubic,
          width: 38,
          height: 38,
          transform: Matrix4.translationValues(_hover ? -2 : 0, 0, 0),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _hover ? AppColors.brand : AppColors.border),
          ),
          child: Icon(Icons.arrow_back_rounded, size: 18, color: color),
        ),
      ),
    );
  }
}
