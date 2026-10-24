/// 客户端内存中的最近 200 条转发记录。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/app_providers.dart';

class LogsPage extends ConsumerWidget {
  const LogsPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final logs = ref.watch(tunnelProvider).logs;
    return Scaffold(
      appBar: AppBar(title: const Text('实时请求日志')),
      body: logs.isEmpty
          ? const Center(child: Text('暂无请求'))
          : ListView.separated(
              itemCount: logs.length,
              separatorBuilder: (_, _) => const Divider(height: 1),
              itemBuilder: (context, i) {
                final e = logs[i];
                return ListTile(
                  dense: true,
                  title: Text('${e.method} ${e.host}${e.path}  -> ${e.status}'),
                  subtitle: Text('${e.target}  ${e.error}'),
                  trailing: Text(
                    '${e.time.hour.toString().padLeft(2, '0')}:${e.time.minute.toString().padLeft(2, '0')}:${e.time.second.toString().padLeft(2, '0')}',
                  ),
                );
              },
            ),
    );
  }
}
