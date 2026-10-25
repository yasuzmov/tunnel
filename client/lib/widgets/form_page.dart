/// 表单类子页面骨架：页头（带返回）+ 居中限宽的滚动内容 + 底部固定操作栏。
library;

import 'package:flutter/material.dart';

import '../app/theme.dart';
import 'fade_slide_in.dart';
import 'page_header.dart';

/// 表单页。
///
/// [sections] 依次错峰入场；[footer] 为底部操作栏右侧按钮，通常是取消 / 保存。
class FormPage extends StatelessWidget {
  const FormPage({
    super.key,
    required this.title,
    this.subtitle,
    required this.sections,
    required this.footer,
  });

  final String title;
  final String? subtitle;
  final List<Widget> sections;
  final List<Widget> footer;

  static const double maxWidth = 760;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: AppColors.background,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          PageHeader(title: title, subtitle: subtitle, showBack: true),
          Expanded(
            child: SingleChildScrollView(
              padding: const EdgeInsets.fromLTRB(28, 4, 28, 28),
              child: Center(
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxWidth: maxWidth),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      for (var i = 0; i < sections.length; i++) ...[
                        if (i > 0) const SizedBox(height: 16),
                        FadeSlideIn(index: i, child: sections[i]),
                      ],
                    ],
                  ),
                ),
              ),
            ),
          ),
          _footerBar(),
        ],
      ),
    );
  }

  Widget _footerBar() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 14),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        border: Border(top: BorderSide(color: AppColors.border)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.end,
        children: footer.expand((w) => [const SizedBox(width: 10), w]).skip(1).toList(),
      ),
    );
  }
}
