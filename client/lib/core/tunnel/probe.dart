/// 一次性连通性测试：握手、welcome、ping/pong 后断开，不开启转发。
library;

import 'dart:async';
import 'dart:io';

import '../../app/app_info.dart';
import '../../models/profile.dart';
import '../../models/rule.dart';
import '../protocol/frame.dart';
import 'events.dart';
import 'session_io.dart';

/// 测试服务端 key 是否有效，并测量一次心跳往返。
Future<WelcomeInfo> testConnectivity(ConnectionProfile profile) async {
  final uri = wsUriFor(profile.serverUrl);
  final socket = await WebSocket.connect(
    uri,
    headers: {'Authorization': 'Bearer ${profile.key}'},
  );
  final io = SessionIO(socket);
  try {
    await io.send(Frame(
      type: FrameType.hello,
      streamId: 0,
      header: {
        'hostname': Platform.localHostname,
        'os': Platform.operatingSystem,
        'version': kClientVersion,
      },
    ));
    WelcomeInfo? info;
    final sentAt = DateTime.now().millisecondsSinceEpoch;
    await for (final frame in io.frames.timeout(const Duration(seconds: 10))) {
      if (frame.type == FrameType.welcome && info == null) {
        final domains = (frame.header['domains'] as List<dynamic>? ?? [])
            .map((e) => e.toString())
            .toList();
        info = WelcomeInfo(
          tunnelId: (frame.header['tunnel_id'] as num?)?.toInt() ?? 0,
          name: (frame.header['name'] as String?) ?? '',
          domains: domains,
          pingMs: 0,
        );
        await io.send(Frame(type: FrameType.ping, streamId: 0, header: {'ts': sentAt}));
        continue;
      }
      if (frame.type == FrameType.pong && info != null) {
        return WelcomeInfo(
          tunnelId: info.tunnelId,
          name: info.name,
          domains: info.domains,
          pingMs: DateTime.now().millisecondsSinceEpoch - sentAt,
        );
      }
    }
    throw const FormatException('未收到 welcome/pong');
  } finally {
    await io.close();
  }
}

/// 直接请求本地目标，确认规则能打到服务。
Future<String> testLocalTarget(ForwardRule rule) async {
  final client = HttpClient();
  if (rule.ignoreTlsError) {
    client.badCertificateCallback = (cert, host, port) => true;
  }
  try {
    final req = await client.getUrl(Uri.parse('${rule.targetUri()}/'));
    req.followRedirects = false;
    final resp = await req.close().timeout(const Duration(seconds: 5));
    await resp.drain<void>();
    return 'HTTP ${resp.statusCode}';
  } finally {
    client.close(force: true);
  }
}

/// 与 [TunnelClient] 使用同一套 URL 规则。
String wsUriFor(String serverUrl) {
  final trimmed = serverUrl.trim();
  if (trimmed.startsWith('ws://') || trimmed.startsWith('wss://')) {
    return trimmed.contains('/tunnel/ws') ? trimmed : '$trimmed/tunnel/ws';
  }
  if (trimmed.startsWith('https://')) {
    return 'wss://${trimmed.substring(8).replaceAll(RegExp(r'/$'), '')}/tunnel/ws';
  }
  if (trimmed.startsWith('http://')) {
    return 'ws://${trimmed.substring(7).replaceAll(RegExp(r'/$'), '')}/tunnel/ws';
  }
  return 'ws://$trimmed/tunnel/ws';
}
