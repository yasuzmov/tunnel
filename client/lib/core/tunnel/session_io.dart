/// WebSocket 写入串行化。Dart 单线程仍可能交错多个 async write。
library;

import 'dart:async';
import 'dart:io';
import 'dart:typed_data';

import '../protocol/frame.dart';

/// 对一条 WebSocket 的收发封装。
class SessionIO {
  SessionIO(this._ws);

  final WebSocket _ws;
  Future<void> _chain = Future.value();

  /// 串行发送一帧。
  Future<void> send(Frame frame) {
    final bytes = encodeFrame(frame);
    _chain = _chain.then((_) async {
      _ws.add(bytes);
    });
    return _chain;
  }

  /// 监听二进制帧。
  Stream<Frame> get frames async* {
    await for (final raw in _ws) {
      if (raw is List<int>) {
        yield decodeFrame(Uint8List.fromList(raw));
      }
    }
  }

  Future<void> close() => _ws.close();
}
