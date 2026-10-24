/// 按顺序匹配转发规则。空 matchHost 表示匹配任意公网域名。
library;

import '../../models/rule.dart';

/// 返回第一条同时满足 Host 与路径前缀的启用规则。
ForwardRule? matchRule(List<ForwardRule> rules, String host, String path) {
  final h = host.toLowerCase();
  for (final rule in rules) {
    if (!rule.enabled) {
      continue;
    }
    if (rule.matchHost.isNotEmpty && rule.matchHost.toLowerCase() != h) {
      continue;
    }
    if (rule.pathPrefix.isNotEmpty && !path.startsWith(rule.pathPrefix)) {
      continue;
    }
    return rule;
  }
  return null;
}
