import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:tunnel_client/core/forward/ws_forwarder.dart';
import 'package:tunnel_client/core/protocol/frame.dart';
import 'package:tunnel_client/core/tunnel/session_io.dart';
import 'package:tunnel_client/models/rule.dart';

void main() {
  test('WebSocket 文本按 UTF-8 透传并携带握手 Header', () async {
    final localServer = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    final localAuth = Completer<String?>();
    localServer.listen((req) async {
      localAuth.complete(req.headers.value('authorization'));
      final ws = await WebSocketTransformer.upgrade(req);
      ws.listen((data) {
        ws.add('世界');
      });
    });

    final ack = Completer<void>();
    final reply = Completer<Frame>();
    final tunnelServer = await HttpServer.bind(InternetAddress.loopbackIPv4, 0);
    tunnelServer.listen((req) async {
      final ws = await WebSocketTransformer.upgrade(req);
      ws.listen((data) {
        final frame = decodeFrame(Uint8List.fromList(data as List<int>));
        if (frame.type == FrameType.wsOpenAck) {
          ack.complete();
        } else if (frame.type == FrameType.wsMsg) {
          reply.complete(frame);
        }
      });
    });

    final tunnelWs = await WebSocket.connect(
      'ws://127.0.0.1:${tunnelServer.port}/',
    );
    final io = SessionIO(tunnelWs);
    final forwarder = WsForwarder(io);

    await forwarder.open(
      streamId: 1,
      header: {
        'path': '/',
        'query': '',
        'headers': {
          'Authorization': ['Bearer local'],
        },
      },
      rule: ForwardRule(id: 'r1', port: localServer.port),
    );
    await ack.future.timeout(const Duration(seconds: 5));

    await forwarder.onMsg(
      1,
      Frame(
        type: FrameType.wsMsg,
        streamId: 1,
        header: {'binary': false},
        payload: Uint8List.fromList(utf8.encode('你好')),
      ),
    );

    final frame = await reply.future.timeout(const Duration(seconds: 5));
    expect(utf8.decode(frame.payload), '世界');
    expect(await localAuth.future, 'Bearer local');

    await forwarder.onClose(1);
    await io.close();
    await tunnelServer.close(force: true);
    await localServer.close(force: true);
  });
}
