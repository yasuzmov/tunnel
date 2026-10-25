/// 分区切换器：所有分区常驻（保留滚动位置、表单输入与嵌套导航栈），
/// 切换时先让旧分区淡出，再让新分区沿侧栏方向纵向滑入，两页不重叠。
///
/// 为什么不用 AnimatedSwitcher：它会销毁旧子树，分区状态会丢失；
/// 为什么不用 PageView：跨多页跳转会依次滚过中间页，观感不连贯。
library;

import 'package:flutter/material.dart';

import '../app/transitions.dart';

/// 方向感知的分区切换容器。
///
/// [index] 当前分区下标；[children] 各分区内容，顺序需与侧栏一致。
/// 非可见分区通过 Offstage + TickerMode 暂停绘制与动画，不占用帧预算。
class SectionSwitcher extends StatefulWidget {
  const SectionSwitcher({super.key, required this.index, required this.children});

  final int index;
  final List<Widget> children;

  @override
  State<SectionSwitcher> createState() => _SectionSwitcherState();
}

class _SectionSwitcherState extends State<SectionSwitcher> with SingleTickerProviderStateMixin {
  static const _enterDistance = 0.075;
  static const _exitDistance = 0.045;

  /// 先让旧分区完全退场，再让新分区进场，避免两页同时抢占视线。
  static const _exitEnd = 0.42;
  static const _enterStart = 0.42;

  late final AnimationController _ctrl = AnimationController(
    vsync: this,
    duration: AppMotion.sectionSwitch,
    value: 1,
  )..addStatusListener(_onStatus);

  late int _current = widget.index;
  int? _previous;

  /// 1 = 向下（新分区从下方进入），-1 = 向上。
  int _direction = 1;

  @override
  void didUpdateWidget(SectionSwitcher oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.index == _current) return;
    _previous = _current;
    _direction = widget.index > _current ? 1 : -1;
    _current = widget.index;
    _ctrl.forward(from: 0);
  }

  void _onStatus(AnimationStatus status) {
    if (status == AnimationStatus.completed && _previous != null) {
      setState(() => _previous = null);
    }
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  /// 计算第 [i] 个分区在当前帧的纵向偏移（占自身高度比例）、不透明度与缩放。
  /// 时间轴前 42% 只负责旧页淡出，后 58% 才让新页进场，中间不重叠。
  (double, double, double) _frameFor(int i) {
    final t = _ctrl.value;
    if (i == _current) {
      final local = ((t - _enterStart) / (1 - _enterStart)).clamp(0.0, 1.0);
      final e = AppMotion.emphasized.transform(local);
      final fade = Curves.easeOut.transform((local / 0.8).clamp(0.0, 1.0));
      return (_direction * _enterDistance * (1 - e), fade, 1 - 0.01 * (1 - e));
    }
    if (i == _previous) {
      final local = (t / _exitEnd).clamp(0.0, 1.0);
      final e = Curves.easeInCubic.transform(local);
      final fade = 1 - Curves.easeIn.transform(local);
      return (-_direction * _exitDistance * e, fade, 1 - 0.006 * e);
    }
    return (0, 0, 1);
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      fit: StackFit.expand,
      children: [
        for (var i = 0; i < widget.children.length; i++) _layer(i),
      ],
    );
  }

  /// 每层的包装结构必须固定不变，否则 Flutter 会重建子树导致分区状态丢失。
  Widget _layer(int i) {
    final visible = i == _current || i == _previous;
    return Offstage(
      offstage: !visible,
      child: TickerMode(
        enabled: visible,
        child: IgnorePointer(
          ignoring: i != _current,
          child: AnimatedBuilder(
            animation: _ctrl,
            builder: (context, child) {
              final (dy, opacity, scale) = _frameFor(i);
              return FractionalTranslation(
                translation: Offset(0, dy),
                child: Transform.scale(
                  scale: scale,
                  child: Opacity(opacity: opacity, child: child),
                ),
              );
            },
            child: RepaintBoundary(child: widget.children[i]),
          ),
        ),
      ),
    );
  }
}
