/// 隧道运行时事件，供界面刷新状态与日志。
library;

/// 连通测试成功结果。
class WelcomeInfo {
  WelcomeInfo({
    required this.tunnelId,
    required this.name,
    required this.domains,
    required this.pingMs,
  });

  final int tunnelId;
  final String name;
  final List<String> domains;
  final int pingMs;
}

/// 最近一次转发摘要，仅保留在内存里。
class RequestSnapshot {
  RequestSnapshot({
    required this.time,
    required this.method,
    required this.host,
    required this.path,
    required this.status,
    required this.target,
    this.error = '',
  });

  final DateTime time;
  final String method;
  final String host;
  final String path;
  final int status;
  final String target;
  final String error;
}
