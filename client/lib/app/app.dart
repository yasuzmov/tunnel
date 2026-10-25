/// 根组件：MaterialApp + 窗口/托盘事件。
/// 关闭窗口时隐藏到托盘而不是退出，避免隧道被误关；托盘菜单「退出程序」才真正退出。
library;

import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';
import 'package:tray_manager/tray_manager.dart';
import 'package:window_manager/window_manager.dart';

import '../shell/app_shell.dart';
import 'app_info.dart';
import 'theme.dart';

class TunnelApp extends StatefulWidget {
  const TunnelApp({super.key});

  @override
  State<TunnelApp> createState() => _TunnelAppState();
}

class _TunnelAppState extends State<TunnelApp> with WindowListener, TrayListener {
  /// 托盘初始化成功后才允许「关闭窗口 = 隐藏到托盘」。
  /// 否则窗口一关就再也找不回来，此时关闭窗口直接退出更安全。
  bool _trayReady = false;

  @override
  void initState() {
    super.initState();
    windowManager.addListener(this);
    trayManager.addListener(this);
    unawaited(_initTray());
  }

  Future<void> _initTray() async {
    try {
      final iconPath = await _materializeTrayIcon();
      await trayManager.setIcon(iconPath);
      await trayManager.setToolTip(kAppName);
      await trayManager.setContextMenu(Menu(items: [
        MenuItem(key: 'show', label: '显示主界面'),
        MenuItem.separator(),
        MenuItem(key: 'exit', label: '退出程序'),
      ]));
      _trayReady = true;
    } catch (e) {
      debugPrint('初始化托盘失败: $e');
    }
  }

  /// tray_manager 在 Windows 上按文件路径加载图标，不能直接读 Flutter asset，
  /// 因此先把打包进 assets 的图标写到临时目录。
  Future<String> _materializeTrayIcon() async {
    final data = await rootBundle.load('assets/tray_icon.ico');
    final dir = await getTemporaryDirectory();
    final file = File(p.join(dir.path, 'tunnel_client_tray.ico'));
    await file.writeAsBytes(data.buffer.asUint8List(), flush: true);
    return file.path;
  }

  @override
  void dispose() {
    windowManager.removeListener(this);
    trayManager.removeListener(this);
    super.dispose();
  }

  Future<void> _showWindow() async {
    await windowManager.show();
    await windowManager.focus();
  }

  /// 退出前先移除托盘图标，避免系统通知区残留幽灵图标。
  Future<void> _exitApp() async {
    try {
      await trayManager.destroy();
    } catch (_) {
      // 托盘已被移除时忽略。
    }
    await windowManager.destroy();
  }

  @override
  void onWindowClose() {
    if (_trayReady) {
      windowManager.hide();
    } else {
      windowManager.destroy();
    }
  }

  @override
  void onTrayIconMouseDown() {
    unawaited(_showWindow());
  }

  @override
  void onTrayIconRightMouseDown() {
    unawaited(trayManager.popUpContextMenu());
  }

  @override
  void onTrayMenuItemClick(MenuItem item) {
    if (item.key == 'show') {
      unawaited(_showWindow());
    } else if (item.key == 'exit') {
      unawaited(_exitApp());
    }
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: kAppName,
      debugShowCheckedModeBanner: false,
      theme: buildAppTheme(),
      // 路由切换时旧页会先淡出；给 Navigator 背后垫一层应用底色，避免透出窗口或桌面。
      builder: (context, child) => ColoredBox(
        color: AppColors.background,
        child: child ?? const SizedBox.shrink(),
      ),
      home: const AppShell(),
    );
  }
}
