/// 隧道客户端：鉴权握手、心跳、指数退避重连，并把 HTTP/WS 分发给转发器。
library;

import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import '../../app/app_info.dart';
import '../../models/profile.dart';
import '../../models/rule.dart';
import '../forward/http_forwarder.dart';
import '../forward/matcher.dart';
import '../forward/ws_forwarder.dart';
import '../protocol/frame.dart';
import 'events.dart';
import 'probe.dart';
import 'session_io.dart';

enum TunnelStatus { idle, connecting, online, reconnecting, error }

/// 长连接隧道。
class TunnelClient {
  TunnelClient({
    required this.profile,
    required this.onStatus,
    required this.onLog,
    required this.onWelcome,
  });

  final ConnectionProfile profile;
  final void Function(TunnelStatus status, String message) onStatus;
  final void Function(RequestSnapshot snap) onLog;
  final void Function(WelcomeInfo info) onWelcome;

  SessionIO? _io;
  Timer? _ping;
  bool _stop = false;
  int _backoff = 1;
  WelcomeInfo? _welcome;
  final _bodies = <int, StreamController<Uint8List>>{};

  /// 停止自动重连并关闭。
  Future<void> stop() async {
    _stop = true;
    _ping?.cancel();
    await _io?.close();
  }

  /// 持续运行直到 [stop]。
  Future<void> run() async {
    _stop = false;
    while (!_stop) {
      try {
        onStatus(TunnelStatus.connecting, '正在连接...');
        await _serveOnce();
        _backoff = 1;
      } catch (err) {
        onStatus(TunnelStatus.error, err.toString());
      }
      if (_stop) {
        return;
      }
      onStatus(TunnelStatus.reconnecting, '${_backoff}s 后重连');
      await Future<void>.delayed(Duration(seconds: _backoff));
      _backoff = (_backoff * 2).clamp(1, 30);
    }
  }

  Future<void> _serveOnce() async {
    _welcome = null;
    final io = await _handshake();
    _io = io;
    final http = HttpForwarder(io);
    final ws = WsForwarder(io);
    _ping = Timer.periodic(const Duration(seconds: 15), (_) {
      io
          .send(
            Frame(
              type: FrameType.ping,
              streamId: 0,
              header: {'ts': DateTime.now().millisecondsSinceEpoch},
            ),
          )
          .ignore();
    });
    onStatus(TunnelStatus.online, '已联通');
    try {
      await for (final frame in io.frames) {
        await _dispatch(frame, http, ws);
      }
    } finally {
      _ping?.cancel();
    }
  }

  Future<void> _dispatch(
    Frame frame,
    HttpForwarder http,
    WsForwarder ws,
  ) async {
    switch (frame.type) {
      case FrameType.welcome:
        _welcome = _parseWelcome(frame, 0);
        onWelcome(_welcome!);
      case FrameType.configUpdate:
        _onConfigUpdate(frame);
      case FrameType.httpReq:
        await _onHttpReq(frame, http);
      case FrameType.bodyChunk:
        _onBody(frame);
      case FrameType.wsOpen:
        await _onWsOpen(frame, ws);
      case FrameType.wsMsg:
        await ws.onMsg(frame.streamId, frame);
      case FrameType.wsClose:
        await ws.onClose(frame.streamId);
      case FrameType.cancel:
        await _bodies[frame.streamId]?.close();
      case FrameType.pong:
        return;
      default:
        return;
    }
  }

  void _onConfigUpdate(Frame frame) {
    final current = _welcome;
    if (current == null) {
      return;
    }
    final domains = (frame.header['domains'] as List<dynamic>? ?? [])
        .map((e) => e.toString())
        .toList();
    _welcome = WelcomeInfo(
      tunnelId: current.tunnelId,
      name: current.name,
      domains: domains,
      pingMs: current.pingMs,
    );
    onWelcome(_welcome!);
  }

  WelcomeInfo _parseWelcome(Frame frame, int pingMs) {
    final domains = (frame.header['domains'] as List<dynamic>? ?? [])
        .map((e) => e.toString())
        .toList();
    return WelcomeInfo(
      tunnelId: (frame.header['tunnel_id'] as num?)?.toInt() ?? 0,
      name: (frame.header['name'] as String?) ?? '',
      domains: domains,
      pingMs: pingMs,
    );
  }

  Future<SessionIO> _handshake() async {
    final uri = wsUriFor(profile.serverUrl);
    final socket = await WebSocket.connect(
      uri,
      headers: {'Authorization': 'Bearer ${profile.key}'},
    );
    final io = SessionIO(socket);
    await io.send(
      Frame(
        type: FrameType.hello,
        streamId: 0,
        header: {
          'hostname': Platform.localHostname,
          'os': Platform.operatingSystem,
          'version': kClientVersion,
        },
      ),
    );
    return io;
  }

  Future<void> _onHttpReq(Frame frame, HttpForwarder http) async {
    final host = (frame.header['host'] as String?) ?? '';
    final path = (frame.header['path'] as String?) ?? '/';
    final rule = matchRule(profile.rules, host, path);
    if (rule == null) {
      await _io?.send(
        Frame(
          type: FrameType.streamError,
          streamId: frame.streamId,
          header: {'code': 'no_matching_rule', 'message': '没有匹配的本地转发规则'},
        ),
      );
      return;
    }
    final ctrl = StreamController<Uint8List>();
    _bodies[frame.streamId] = ctrl;
    unawaited(_runHttp(frame, http, rule, ctrl));
  }

  Future<void> _runHttp(
    Frame frame,
    HttpForwarder http,
    ForwardRule rule,
    StreamController<Uint8List> ctrl,
  ) async {
    try {
      await http.handle(
        streamId: frame.streamId,
        header: frame.header,
        body: ctrl.stream,
        rule: rule,
        onStatus: (status, target) {
          onLog(
            RequestSnapshot(
              time: DateTime.now(),
              method: (frame.header['method'] as String?) ?? '',
              host: (frame.header['host'] as String?) ?? '',
              path: (frame.header['path'] as String?) ?? '',
              status: status,
              target: target,
            ),
          );
        },
      );
    } catch (err) {
      await _io?.send(
        Frame(
          type: FrameType.streamError,
          streamId: frame.streamId,
          header: {'code': 'upstream_error', 'message': err.toString()},
        ),
      );
      onLog(
        RequestSnapshot(
          time: DateTime.now(),
          method: (frame.header['method'] as String?) ?? '',
          host: (frame.header['host'] as String?) ?? '',
          path: (frame.header['path'] as String?) ?? '',
          status: 0,
          target: rule.targetUri(),
          error: err.toString(),
        ),
      );
    } finally {
      _bodies.remove(frame.streamId);
    }
  }

  void _onBody(Frame frame) {
    final ctrl = _bodies[frame.streamId];
    if (ctrl == null) {
      return;
    }
    if (frame.payload.isNotEmpty) {
      ctrl.add(frame.payload);
    }
    if (frame.header['eof'] == true) {
      unawaited(ctrl.close());
    }
  }

  Future<void> _onWsOpen(Frame frame, WsForwarder ws) async {
    final host = (frame.header['host'] as String?) ?? '';
    final path = (frame.header['path'] as String?) ?? '/';
    final rule = matchRule(profile.rules, host, path);
    if (rule == null) {
      await _io?.send(
        Frame(
          type: FrameType.streamError,
          streamId: frame.streamId,
          header: {'code': 'no_matching_rule', 'message': '没有匹配的本地转发规则'},
        ),
      );
      return;
    }
    await ws.open(streamId: frame.streamId, header: frame.header, rule: rule);
  }
}
