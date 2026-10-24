/// 把隧道上的 HTTP 请求转发到本机端口或 hosts 域名。
/// 关闭自动跳转与自动解压，响应原样回传。
library;

import 'dart:io';
import 'dart:typed_data';

import '../../models/rule.dart';
import '../protocol/frame.dart';
import '../tunnel/session_io.dart';

/// 执行一次本地 HTTP 转发并按块回写响应。
class HttpForwarder {
  HttpForwarder(this.io);

  final SessionIO io;

  /// [onStatus] 用于写入客户端实时日志。
  Future<void> handle({
    required int streamId,
    required Map<String, dynamic> header,
    required Stream<Uint8List> body,
    required ForwardRule rule,
    required void Function(int status, String target) onStatus,
  }) async {
    final target = _targetUri(rule, header);
    final client = HttpClient()..autoUncompress = false;
    if (rule.ignoreTlsError) {
      client.badCertificateCallback = (cert, host, port) => true;
    }
    try {
      final req = await client.openUrl(
        (header['method'] as String?) ?? 'GET',
        target,
      );
      req.followRedirects = false;
      _copyHeaders(req, header, rule);
      await for (final chunk in body) {
        req.add(chunk);
      }
      final resp = await req.close();
      onStatus(resp.statusCode, rule.targetUri());
      await _writeResponse(streamId, resp, rule);
    } finally {
      client.close(force: true);
    }
  }

  Uri _targetUri(ForwardRule rule, Map<String, dynamic> header) {
    final path = (header['path'] as String?) ?? '/';
    final query = (header['query'] as String?) ?? '';
    final base = rule.targetUri();
    return Uri.parse('$base$path${query.isEmpty ? '' : '?$query'}');
  }

  void _copyHeaders(HttpClientRequest req, Map<String, dynamic> header, ForwardRule rule) {
    final raw = header['headers'];
    if (raw is Map) {
      raw.forEach((key, value) {
        if (_skip.contains(key.toString().toLowerCase())) {
          return;
        }
        if (value is List) {
          req.headers.set(key.toString(), value.map((e) => e.toString()).toList());
        }
      });
    }
    if (rule.rewriteHost) {
      final host = rule.rewriteHostValue.isNotEmpty
          ? rule.rewriteHostValue
          : (rule.targetType == 'hostsDomain' ? rule.hostsDomain : '127.0.0.1');
      req.headers.host = host;
    }
  }

  Future<void> _writeResponse(int streamId, HttpClientResponse resp, ForwardRule rule) async {
    final headers = <String, List<String>>{};
    resp.headers.forEach((name, values) {
      headers[name] = values;
    });
    await io.send(Frame(
      type: FrameType.httpResp,
      streamId: streamId,
      header: {
        'status': resp.statusCode,
        'headers': headers,
        'matched_rule': rule.label,
        'local_target': rule.targetUri(),
      },
    ));
    await for (final chunk in resp) {
      for (final part in _split(Uint8List.fromList(chunk))) {
        await io.send(Frame(
          type: FrameType.bodyChunk,
          streamId: streamId,
          header: {'eof': false},
          payload: part,
        ));
      }
    }
    await io.send(Frame(
      type: FrameType.bodyChunk,
      streamId: streamId,
      header: {'eof': true},
    ));
  }
}

final _skip = {
  'connection',
  'keep-alive',
  'proxy-authenticate',
  'proxy-authorization',
  'te',
  'trailer',
  'transfer-encoding',
  'upgrade',
  'content-length',
};

List<Uint8List> _split(Uint8List data) {
  if (data.length <= chunkSize) {
    return [data];
  }
  final out = <Uint8List>[];
  for (var i = 0; i < data.length; i += chunkSize) {
    final end = i + chunkSize > data.length ? data.length : i + chunkSize;
    out.add(Uint8List.sublistView(data, i, end));
  }
  return out;
}
