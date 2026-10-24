/// 把公网 WebSocket 中继到本机服务。先完成本地握手再 ack，避免过早升级公网连接。
library;

import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import '../../models/rule.dart';
import '../protocol/frame.dart';
import '../tunnel/session_io.dart';

/// 本地 WebSocket 转发器。
class WsForwarder {
  WsForwarder(this.io);

  final SessionIO io;
  final _locals = <int, WebSocket>{};
  final _clients = <int, HttpClient>{};

  Future<void> open({
    required int streamId,
    required Map<String, dynamic> header,
    required ForwardRule rule,
  }) async {
    final scheme = rule.useHttps ? 'wss' : 'ws';
    final host = rule.targetType == 'hostsDomain'
        ? rule.hostsDomain
        : '127.0.0.1';
    final path = (header['path'] as String?) ?? '/';
    final query = (header['query'] as String?) ?? '';
    final uri = Uri.parse(
      '$scheme://$host:${rule.port}$path${query.isEmpty ? '' : '?$query'}',
    );
    HttpClient? client;
    if (rule.ignoreTlsError) {
      client = HttpClient()
        ..badCertificateCallback = (cert, host, port) => true;
    }
    try {
      final ws = await WebSocket.connect(
        uri.toString(),
        headers: _forwardHeaders(header),
        customClient: client,
      );
      _locals[streamId] = ws;
      if (client != null) {
        _clients[streamId] = client;
      }
      await io.send(
        Frame(
          type: FrameType.wsOpenAck,
          streamId: streamId,
          header: {
            'ok': true,
            'status': 101,
            'headers': <String, List<String>>{},
            'matched_rule': rule.label,
            'local_target': rule.targetUri(),
          },
        ),
      );
      ws.listen(
        (data) => _fromLocal(streamId, data),
        onDone: () {
          _locals.remove(streamId);
          _clients.remove(streamId)?.close(force: true);
          io
              .send(
                Frame(
                  type: FrameType.wsClose,
                  streamId: streamId,
                  header: {'code': 1000, 'reason': 'local_closed'},
                ),
              )
              .ignore();
        },
      );
    } catch (err) {
      _locals.remove(streamId)?.close().ignore();
      _clients.remove(streamId)?.close(force: true);
      client?.close(force: true);
      await io.send(
        Frame(
          type: FrameType.wsOpenAck,
          streamId: streamId,
          header: {'ok': false, 'status': 502, 'message': err.toString()},
        ),
      );
    }
  }

  Map<String, String> _forwardHeaders(Map<String, dynamic> header) {
    final raw = header['headers'];
    final out = <String, String>{};
    if (raw is Map) {
      raw.forEach((key, value) {
        final name = key.toString();
        if (_wsSkip.contains(name.toLowerCase())) {
          return;
        }
        if (value is List) {
          out[name] = value.map((e) => e.toString()).join(', ');
        } else if (value != null) {
          out[name] = value.toString();
        }
      });
    }
    return out;
  }

  Future<void> onMsg(int streamId, Frame frame) async {
    final ws = _locals[streamId];
    if (ws == null) {
      return;
    }
    final binary = frame.header['binary'] == true;
    if (binary) {
      ws.add(frame.payload);
    } else {
      ws.add(utf8.decode(frame.payload));
    }
  }

  Future<void> onClose(int streamId) async {
    final ws = _locals.remove(streamId);
    await ws?.close();
    _clients.remove(streamId)?.close(force: true);
  }

  Future<void> _fromLocal(int streamId, dynamic data) async {
    var binary = true;
    late Uint8List payload;
    if (data is String) {
      binary = false;
      payload = Uint8List.fromList(utf8.encode(data));
    } else if (data is List<int>) {
      payload = Uint8List.fromList(data);
    } else {
      return;
    }
    await io.send(
      Frame(
        type: FrameType.wsMsg,
        streamId: streamId,
        header: {'binary': binary},
        payload: payload,
      ),
    );
  }
}

final _wsSkip = {
  'connection',
  'keep-alive',
  'proxy-authenticate',
  'proxy-authorization',
  'te',
  'trailer',
  'transfer-encoding',
  'upgrade',
  'content-length',
  'host',
  'sec-websocket-key',
  'sec-websocket-version',
  'sec-websocket-extensions',
  'sec-websocket-protocol',
};
