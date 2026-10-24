import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:tunnel_client/core/protocol/frame.dart';

void main() {
  test('frame round trip', () {
    final encoded = encodeFrame(Frame(
      type: FrameType.httpReq,
      streamId: 7,
      header: {'method': 'GET', 'host': 'a.com', 'path': '/'},
      payload: Uint8List.fromList('hi'.codeUnits),
    ));
    final decoded = decodeFrame(encoded);
    expect(decoded.type, FrameType.httpReq);
    expect(decoded.streamId, 7);
    expect(decoded.header['host'], 'a.com');
    expect(String.fromCharCodes(decoded.payload), 'hi');
  });
}
