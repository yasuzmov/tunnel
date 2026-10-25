/// 首次出现时上移淡入，[index] 用于列表项错峰出场（每项延迟 40ms，最多 8 项）。
library;

import 'dart:async';

import 'package:flutter/material.dart';

import '../app/transitions.dart';

/// 入场动画包装。只在 State 创建时播放一次，列表刷新不会重复触发。
class FadeSlideIn extends StatefulWidget {
  const FadeSlideIn({super.key, required this.child, this.index = 0, this.offset = 16});

  final Widget child;

  /// 错峰序号；超过 8 的统一按 8 计，避免长列表末尾等待过久。
  final int index;

  /// 起始下移距离（逻辑像素）。
  final double offset;

  @override
  State<FadeSlideIn> createState() => _FadeSlideInState();
}

class _FadeSlideInState extends State<FadeSlideIn> with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl = AnimationController(
    vsync: this,
    duration: const Duration(milliseconds: 480),
  );
  late final Animation<double> _curve = CurvedAnimation(parent: _ctrl, curve: AppMotion.smooth);
  Timer? _delay;

  @override
  void initState() {
    super.initState();
    final wait = Duration(milliseconds: 40 * widget.index.clamp(0, 8));
    if (wait == Duration.zero) {
      _ctrl.forward();
    } else {
      _delay = Timer(wait, () {
        if (mounted) _ctrl.forward();
      });
    }
  }

  @override
  void dispose() {
    _delay?.cancel();
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _curve,
      builder: (context, child) => Opacity(
        opacity: _curve.value,
        child: Transform.translate(offset: Offset(0, widget.offset * (1 - _curve.value)), child: child),
      ),
      child: widget.child,
    );
  }
}
