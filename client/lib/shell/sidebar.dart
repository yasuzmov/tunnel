/// 浅色侧栏：品牌、分区导航（选中背景块平滑滑动）、底部连接状态卡片。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:window_manager/window_manager.dart';

import '../app/theme.dart';
import '../app/transitions.dart';
import '../providers/app_providers.dart';
import '../utils/status_style.dart';
import '../widgets/status_dot.dart';
import 'app_section.dart';

const double _itemHeight = 44;
const double _itemGap = 4;

/// 侧栏。
///
/// [current] 当前分区；[onSelect] 点击分区回调（重复点击当前分区也会回调，
/// 由外层决定是否回到该分区根页面）。
class Sidebar extends StatelessWidget {
  const Sidebar({super.key, required this.current, required this.onSelect});

  final AppSection current;
  final ValueChanged<AppSection> onSelect;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 220,
      decoration: const BoxDecoration(
        color: AppColors.sidebar,
        border: Border(right: BorderSide(color: AppColors.border)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const DragToMoveArea(child: _Brand()),
          const SizedBox(height: 12),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: _NavList(current: current, onSelect: onSelect),
          ),
          const Spacer(),
          const Padding(padding: EdgeInsets.all(12), child: _StatusCard()),
        ],
      ),
    );
  }
}

class _Brand extends StatelessWidget {
  const _Brand();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 72,
      padding: const EdgeInsets.fromLTRB(20, 16, 12, 0),
      alignment: Alignment.centerLeft,
      child: Row(
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(10),
            child: Image.asset('assets/app_icon.png', width: 32, height: 32),
          ),
          const SizedBox(width: 10),
          const Text('内网穿透', style: TextStyle(color: AppColors.text, fontSize: 16, fontWeight: FontWeight.w700)),
        ],
      ),
    );
  }
}

/// 导航列表：选中背景块是独立图层，用 AnimatedPositioned 在条目间滑动。
class _NavList extends StatelessWidget {
  const _NavList({required this.current, required this.onSelect});

  final AppSection current;
  final ValueChanged<AppSection> onSelect;

  @override
  Widget build(BuildContext context) {
    const values = AppSection.values;
    return SizedBox(
      height: values.length * (_itemHeight + _itemGap),
      child: Stack(
        children: [
          AnimatedPositioned(
            duration: AppMotion.sectionSwitch,
            curve: AppMotion.emphasized,
            top: current.index * (_itemHeight + _itemGap),
            left: 0,
            right: 0,
            height: _itemHeight,
            child: const _Indicator(),
          ),
          for (final s in values)
            Positioned(
              top: s.index * (_itemHeight + _itemGap),
              left: 0,
              right: 0,
              height: _itemHeight,
              child: _NavItem(section: s, active: s == current, onTap: () => onSelect(s)),
            ),
        ],
      ),
    );
  }
}

class _Indicator extends StatelessWidget {
  const _Indicator();

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(10),
        color: AppColors.sidebarActive,
      ),
      child: Align(
        alignment: Alignment.centerLeft,
        child: Container(
          margin: const EdgeInsets.only(left: 2),
          width: 3,
          height: 18,
          decoration: BoxDecoration(
            color: AppColors.brand,
            borderRadius: BorderRadius.circular(2),
          ),
        ),
      ),
    );
  }
}

class _NavItem extends StatefulWidget {
  const _NavItem({required this.section, required this.active, required this.onTap});

  final AppSection section;
  final bool active;
  final VoidCallback onTap;

  @override
  State<_NavItem> createState() => _NavItemState();
}

class _NavItemState extends State<_NavItem> {
  bool _hover = false;

  @override
  Widget build(BuildContext context) {
    final active = widget.active;
    final textColor = active ? AppColors.text : (_hover ? AppColors.text : AppColors.text2);
    final iconColor = active ? AppColors.brand : textColor;
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: GestureDetector(
        onTap: widget.onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 220),
          curve: AppMotion.smooth,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(10),
            color: _hover && !active ? AppColors.sidebarHover : Colors.transparent,
          ),
          child: Row(
            children: [
              TweenAnimationBuilder<Color?>(
                tween: ColorTween(end: iconColor),
                duration: const Duration(milliseconds: 220),
                curve: AppMotion.smooth,
                builder: (context, iconColor, _) => Icon(
                  active ? widget.section.activeIcon : widget.section.icon,
                  size: 19,
                  color: iconColor,
                ),
              ),
              const SizedBox(width: 12),
              AnimatedDefaultTextStyle(
                duration: const Duration(milliseconds: 220),
                curve: AppMotion.smooth,
                style: DefaultTextStyle.of(context).style.copyWith(
                      color: textColor,
                      fontSize: 14,
                      fontWeight: active ? FontWeight.w600 : FontWeight.w400,
                    ),
                child: Text(widget.section.label),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// 底部当前隧道状态，状态变化时内容淡入淡出切换。
class _StatusCard extends ConsumerWidget {
  const _StatusCard();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(tunnelProvider);
    final style = statusStyleOf(state.status);
    final welcome = state.welcome;
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.border),
      ),
      child: AnimatedSwitcher(
        duration: const Duration(milliseconds: 300),
        child: Row(
          key: ValueKey('${state.status}-${welcome?.name}'),
          children: [
            StatusDot(color: style.color, pulse: style.pulse, size: 8),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(style.label, style: const TextStyle(color: AppColors.text, fontSize: 13, fontWeight: FontWeight.w600)),
                  const SizedBox(height: 2),
                  Text(
                    welcome != null ? '${welcome.name} · ${welcome.pingMs}ms' : '选择一个连接开始',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(color: AppColors.text3, fontSize: 11.5),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
