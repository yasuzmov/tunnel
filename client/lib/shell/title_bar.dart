/// 自定义窗口标题栏：替换系统标题栏，使侧栏深色背景延伸到窗口顶部。
/// 关闭按钮触发 windowManager.close()，由 TunnelApp.onWindowClose 改为隐藏到托盘。
library;

import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

import '../app/theme.dart';

/// 标题栏高度，侧栏顶部留出同样高度以对齐。
const double kTitleBarHeight = 40;

/// 标题栏右侧区域（窗口按钮 + 可拖拽空白）。左侧品牌区由侧栏自行绘制。
class TitleBar extends StatefulWidget {
  const TitleBar({super.key, this.title = ''});

  /// 显示在拖拽区的当前页面标题。
  final String title;

  @override
  State<TitleBar> createState() => _TitleBarState();
}

class _TitleBarState extends State<TitleBar> with WindowListener {
  bool _maximized = false;

  @override
  void initState() {
    super.initState();
    windowManager.addListener(this);
    windowManager.isMaximized().then((v) {
      if (mounted) setState(() => _maximized = v);
    });
  }

  @override
  void dispose() {
    windowManager.removeListener(this);
    super.dispose();
  }

  @override
  void onWindowMaximize() => setState(() => _maximized = true);

  @override
  void onWindowUnmaximize() => setState(() => _maximized = false);

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: kTitleBarHeight,
      child: Row(
        children: [
          Expanded(
            child: DragToMoveArea(
              child: Container(
                alignment: Alignment.centerLeft,
                padding: const EdgeInsets.only(left: 20),
                child: AnimatedSwitcher(
                  duration: const Duration(milliseconds: 250),
                  child: Text(
                    widget.title,
                    key: ValueKey(widget.title),
                    style: const TextStyle(fontSize: 12, color: AppColors.text3),
                  ),
                ),
              ),
            ),
          ),
          ..._buttons(),
        ],
      ),
    );
  }

  List<Widget> _buttons() {
    return [
      WindowCaptionButton.minimize(brightness: Brightness.light, onPressed: windowManager.minimize),
      _maximized
          ? WindowCaptionButton.unmaximize(brightness: Brightness.light, onPressed: windowManager.unmaximize)
          : WindowCaptionButton.maximize(brightness: Brightness.light, onPressed: windowManager.maximize),
      WindowCaptionButton.close(brightness: Brightness.light, onPressed: windowManager.close),
    ];
  }
}
