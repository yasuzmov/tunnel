/// 隧道二进制帧编解码。
/// 格式与服务端 Go 实现一致：1B 类型 + 4B streamId + 4B 头长度 + JSON 头 + 载荷，大端。
library;

import 'dart:convert';
import 'dart:typed_data';

/// 帧类型常量，数值不可改。
class FrameType {
  static const int hello = 0;
  static const int welcome = 1;
  static const int ping = 2;
  static const int pong = 3;
  static const int configUpdate = 4;
  static const int httpReq = 5;
  static const int httpResp = 6;
  static const int bodyChunk = 7;
  static const int streamError = 8;
  static const int cancel = 9;
  static const int wsOpen = 10;
  static const int wsOpenAck = 11;
  static const int wsMsg = 12;
  static const int wsClose = 13;
}

/// 与服务端对齐的分块大小。
const int chunkSize = 32 * 1024;
const int maxHeaderSize = 256 * 1024;
const int maxPayloadSize = 64 * 1024;

/// 一条隧道消息。
class Frame {
  Frame({
    required this.type,
    required this.streamId,
    Map<String, dynamic>? header,
    Uint8List? payload,
  })  : header = header ?? <String, dynamic>{},
        payload = payload ?? Uint8List(0);

  final int type;
  final int streamId;
  final Map<String, dynamic> header;
  final Uint8List payload;
}

/// 编码为 WebSocket 二进制消息。
Uint8List encodeFrame(Frame frame) {
  final headerBytes = utf8.encode(jsonEncode(frame.header));
  if (headerBytes.length > maxHeaderSize) {
    throw StateError('帧头过大');
  }
  if (frame.payload.length > maxPayloadSize) {
    throw StateError('帧载荷过大');
  }
  final out = Uint8List(9 + headerBytes.length + frame.payload.length);
  final bd = ByteData.sublistView(out);
  bd.setUint8(0, frame.type);
  bd.setUint32(1, frame.streamId);
  bd.setUint32(5, headerBytes.length);
  out.setRange(9, 9 + headerBytes.length, headerBytes);
  out.setRange(9 + headerBytes.length, out.length, frame.payload);
  return out;
}

/// 解析二进制消息。
Frame decodeFrame(Uint8List data) {
  if (data.length < 9) {
    throw const FormatException('帧过短');
  }
  final bd = ByteData.sublistView(data);
  final headerLen = bd.getUint32(5);
  if (headerLen > maxHeaderSize || data.length < 9 + headerLen) {
    throw const FormatException('帧头长度非法');
  }
  final headerRaw = utf8.decode(data.sublist(9, 9 + headerLen));
  final header = jsonDecode(headerRaw) as Map<String, dynamic>;
  final payload = data.sublist(9 + headerLen);
  return Frame(
    type: bd.getUint8(0),
    streamId: bd.getUint32(1),
    header: header,
    payload: Uint8List.fromList(payload),
  );
}
