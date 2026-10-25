/// 本地转发规则。按列表顺序匹配，命中即停。
library;

/// 一条把公网请求转到本机的规则。
class ForwardRule {
  ForwardRule({
    required this.id,
    this.matchHost = '',
    this.pathPrefix = '',
    this.targetType = 'port',
    this.port = 80,
    this.hostsDomain = '',
    this.useHttps = false,
    this.ignoreTlsError = false,
    this.rewriteHost = false,
    this.rewriteHostValue = '',
    this.enabled = true,
  });

  String id;
  String matchHost;
  String pathPrefix;
  String targetType;
  int port;
  String hostsDomain;
  bool useHttps;
  bool ignoreTlsError;
  bool rewriteHost;
  String rewriteHostValue;
  bool enabled;

  factory ForwardRule.fromJson(Map<String, dynamic> json) {
    return ForwardRule(
      id: json['id'] as String? ?? '',
      matchHost: json['matchHost'] as String? ?? '',
      pathPrefix: json['pathPrefix'] as String? ?? '',
      targetType: json['targetType'] as String? ?? 'port',
      port: json['port'] as int? ?? 80,
      hostsDomain: json['hostsDomain'] as String? ?? '',
      useHttps: json['useHttps'] as bool? ?? false,
      ignoreTlsError: json['ignoreTlsError'] as bool? ?? false,
      rewriteHost: json['rewriteHost'] as bool? ?? false,
      rewriteHostValue: json['rewriteHostValue'] as String? ?? '',
      enabled: json['enabled'] as bool? ?? true,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'matchHost': matchHost,
        'pathPrefix': pathPrefix,
        'targetType': targetType,
        'port': port,
        'hostsDomain': hostsDomain,
        'useHttps': useHttps,
        'ignoreTlsError': ignoreTlsError,
        'rewriteHost': rewriteHost,
        'rewriteHostValue': rewriteHostValue,
        'enabled': enabled,
      };

  /// 深拷贝。编辑页基于副本修改，点击保存才写回，取消不会污染原配置。
  ForwardRule copy() => ForwardRule.fromJson(toJson());

  /// 用于日志的规则简述。
  String get label {
    final host = matchHost.isEmpty ? '*' : matchHost;
    final path = pathPrefix.isEmpty ? '/' : pathPrefix;
    return '$host$path -> ${targetUri()}';
  }

  /// 拼出本地目标 URL（不含 path）。
  String targetUri() {
    final scheme = useHttps ? 'https' : 'http';
    if (targetType == 'hostsDomain') {
      return '$scheme://$hostsDomain:$port';
    }
    return '$scheme://127.0.0.1:$port';
  }
}
