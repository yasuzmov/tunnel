/// 测试结果展示：空闲 / 测试中 / 成功 / 失败四种状态，切换时淡入并平滑改变高度。
/// 连通性测试与本地目标测试共用。
library;

import 'package:flutter/material.dart';

import '../app/theme.dart';

/// 测试阶段。
enum ProbePhase { idle, testing, success, failure }

/// 一次测试的结果。
class ProbeResult {
  const ProbeResult(this.phase, [this.title = '', this.lines = const []]);

  const ProbeResult.idle() : this(ProbePhase.idle);
  const ProbeResult.testing() : this(ProbePhase.testing, '测试中…');

  final ProbePhase phase;
  final String title;

  /// 详情行，如「记录：xxx」「延迟：12 ms」。
  final List<String> lines;
}

/// 结果面板；idle 时不占空间。
class ProbeResultView extends StatelessWidget {
  const ProbeResultView({super.key, required this.result});

  final ProbeResult result;

  @override
  Widget build(BuildContext context) {
    return AnimatedSize(
      duration: const Duration(milliseconds: 320),
      curve: Curves.easeOutCubic,
      alignment: Alignment.topCenter,
      child: AnimatedSwitcher(
        duration: const Duration(milliseconds: 260),
        transitionBuilder: (child, anim) => FadeTransition(
          opacity: anim,
          child: SlideTransition(
            position: Tween(begin: const Offset(0, 0.08), end: Offset.zero).animate(anim),
            child: child,
          ),
        ),
        child: result.phase == ProbePhase.idle
            ? const SizedBox(width: double.infinity, key: ValueKey('idle'))
            : _panel(),
      ),
    );
  }

  Widget _panel() {
    final (color, icon) = switch (result.phase) {
      ProbePhase.success => (AppColors.success, Icons.check_circle_rounded),
      ProbePhase.failure => (AppColors.danger, Icons.error_rounded),
      _ => (AppColors.brand, Icons.hourglass_top_rounded),
    };
    return Container(
      key: ValueKey('${result.phase}-${result.title}'),
      width: double.infinity,
      margin: const EdgeInsets.only(top: 16),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.07),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: color.withValues(alpha: 0.25)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          result.phase == ProbePhase.testing
              ? SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: color))
              : Icon(icon, size: 18, color: color),
          const SizedBox(width: 10),
          Expanded(child: _texts(color)),
        ],
      ),
    );
  }

  Widget _texts(Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(result.title, style: TextStyle(color: color, fontWeight: FontWeight.w600)),
        for (final line in result.lines) ...[
          const SizedBox(height: 4),
          SelectableText(line, style: const TextStyle(fontSize: 12.5, color: AppColors.text2)),
        ],
      ],
    );
  }
}
