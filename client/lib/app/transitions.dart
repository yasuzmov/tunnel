/// 页面转场：所有 MaterialPageRoute 统一使用水平滑动。
///
/// 进入页从右侧 20% 处滑入并淡入；被覆盖的页同时轻微左移，并在新页到位前淡出，
/// 避免旧页以半透明状态停在新页下方。曲线用强调型快出慢停，返回时用其反向，
/// 手感上「起步干脆、落点柔和」。只动 transform / opacity，不触发重新布局。
library;

import 'package:flutter/material.dart';

/// 统一动效参数：曲线偏「快出慢停」，长距离位移收尾更柔和。
class AppMotion {
  AppMotion._();

  /// 通用丝滑曲线，用于卡片、颜色与侧栏指示块。
  static const smooth = Cubic(0.22, 1, 0.36, 1);

  /// 强调曲线，用于分区与页面切换，起步更果断。
  static const emphasized = Cubic(0.16, 1, 0.3, 1);

  static const sectionSwitch = Duration(milliseconds: 500);
  static const routeEnter = Duration(milliseconds: 460);
  static const routeExit = Duration(milliseconds: 360);
}

/// 平滑滑动转场构建器，注册在 [ThemeData.pageTransitionsTheme] 中。
class SmoothSlidePageTransitionsBuilder extends PageTransitionsBuilder {
  const SmoothSlidePageTransitionsBuilder();

  @override
  Duration get transitionDuration => AppMotion.routeEnter;

  @override
  Duration get reverseTransitionDuration => AppMotion.routeExit;

  @override
  Widget buildTransitions<T>(
    PageRoute<T> route,
    BuildContext context,
    Animation<double> animation,
    Animation<double> secondaryAnimation,
    Widget child,
  ) {
    return SlideParallax(
      animation: animation,
      secondaryAnimation: secondaryAnimation,
      child: child,
    );
  }
}

/// 进入/覆盖两段动画的组合，供路由转场与分区切换复用。
class SlideParallax extends StatelessWidget {
  const SlideParallax({
    super.key,
    required this.animation,
    required this.secondaryAnimation,
    required this.child,
    this.axis = Axis.horizontal,
  });

  /// 本页进入进度 0→1。
  final Animation<double> animation;

  /// 本页被上层覆盖的进度 0→1。
  final Animation<double> secondaryAnimation;

  /// 滑动方向；分区切换使用纵向以对应侧栏的上下排列。
  final Axis axis;
  final Widget child;

  Offset _offset(double v) => axis == Axis.horizontal ? Offset(v, 0) : Offset(0, v);

  @override
  Widget build(BuildContext context) {
    // 先让被覆盖页退场，再让新页进场；两段不重叠，避免旧页停在新页下方。
    const switchPoint = 0.35;
    final enterSlide = CurvedAnimation(
      parent: animation,
      curve: const Interval(switchPoint, 1.0, curve: AppMotion.emphasized),
      reverseCurve: const Interval(switchPoint, 1.0, curve: Curves.easeInCubic),
    );
    final enterFade = CurvedAnimation(
      parent: animation,
      curve: const Interval(switchPoint, 0.85, curve: Curves.easeOut),
      reverseCurve: const Interval(switchPoint, 1.0, curve: Curves.easeIn),
    );
    final coverFade = CurvedAnimation(
      parent: secondaryAnimation,
      curve: const Interval(0.0, switchPoint, curve: Curves.easeIn),
      reverseCurve: const Interval(1 - switchPoint, 1.0, curve: Curves.easeOut),
    );
    final coverSlide = CurvedAnimation(
      parent: secondaryAnimation,
      curve: const Interval(0.0, switchPoint, curve: Curves.easeInCubic),
      reverseCurve: const Interval(1 - switchPoint, 1.0, curve: Curves.easeOut),
    );
    return SlideTransition(
      position: Tween(begin: Offset.zero, end: _offset(-0.02)).animate(coverSlide),
      child: FadeTransition(
        opacity: Tween(begin: 1.0, end: 0.0).animate(coverFade),
        child: SlideTransition(
          position: Tween(begin: _offset(0.20), end: Offset.zero).animate(enterSlide),
          child: FadeTransition(
            opacity: Tween(begin: 0.0, end: 1.0).animate(enterFade),
            child: child,
          ),
        ),
      ),
    );
  }
}

/// 以统一转场推入新页面。
///
/// [context] 用于定位最近的 Navigator（在分区内即为分区自己的嵌套导航）。
/// 返回新页面 pop 时携带的结果。
Future<T?> pushSlide<T>(BuildContext context, Widget page) {
  return Navigator.of(context).push<T>(MaterialPageRoute(builder: (_) => page));
}
