/// 解析 Windows hosts，供规则下拉框选择本机虚拟域名。
library;

import 'dart:io';

/// 一条 hosts 记录。
class HostsEntry {
  HostsEntry(this.ip, this.hostname);
  final String ip;
  final String hostname;
}

/// 读取系统 hosts。解析失败时返回空列表，不抛给 UI。
Future<List<HostsEntry>> readHostsFile() async {
  const path = r'C:\Windows\System32\drivers\etc\hosts';
  final file = File(path);
  if (!await file.exists()) {
    return [];
  }
  final lines = await file.readAsLines();
  final out = <HostsEntry>[];
  final seen = <String>{};
  for (final raw in lines) {
    final line = raw.trim();
    if (line.isEmpty || line.startsWith('#')) {
      continue;
    }
    final parts = line.split(RegExp(r'\s+'));
    if (parts.length < 2) {
      continue;
    }
    final ip = parts.first;
    for (final host in parts.skip(1)) {
      if (host.startsWith('#') || host.isEmpty) {
        break;
      }
      final key = host.toLowerCase();
      if (seen.add(key)) {
        out.add(HostsEntry(ip, host));
      }
    }
  }
  return out;
}
