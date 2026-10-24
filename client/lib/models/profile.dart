/// 一条可保存的服务器连接配置。
library;

import 'rule.dart';

/// 客户端侧的穿透配置：接入地址、key 与本地转发规则。
class ConnectionProfile {
  ConnectionProfile({
    required this.id,
    required this.name,
    required this.serverUrl,
    required this.key,
    this.autoConnect = false,
    List<ForwardRule>? rules,
  }) : rules = rules ?? [];

  String id;
  String name;
  String serverUrl;
  String key;
  bool autoConnect;
  List<ForwardRule> rules;

  factory ConnectionProfile.fromJson(Map<String, dynamic> json) {
    final rawRules = json['rules'] as List<dynamic>? ?? [];
    return ConnectionProfile(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      serverUrl: json['serverUrl'] as String? ?? '',
      key: json['key'] as String? ?? '',
      autoConnect: json['autoConnect'] as bool? ?? false,
      rules: rawRules
          .map((e) => ForwardRule.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'serverUrl': serverUrl,
        'key': key,
        'autoConnect': autoConnect,
        'rules': rules.map((e) => e.toJson()).toList(),
      };
}
