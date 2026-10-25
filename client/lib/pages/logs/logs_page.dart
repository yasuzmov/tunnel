/// 实时日志分区：展示内存中最近 200 条转发记录，支持关键字过滤与清空。
/// 完整的 Header / Body 日志在服务端管理后台查看。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../core/tunnel/events.dart';
import '../../providers/app_providers.dart';
import '../../widgets/empty_state.dart';
import '../../widgets/fade_slide_in.dart';
import '../../widgets/page_header.dart';
import 'widgets/log_row.dart';

class LogsPage extends ConsumerStatefulWidget {
  const LogsPage({super.key});

  @override
  ConsumerState<LogsPage> createState() => _LogsPageState();
}

class _LogsPageState extends ConsumerState<LogsPage> {
  String _keyword = '';

  /// 关键字同时匹配方法、Host、路径、状态码与本地目标。
  List<RequestSnapshot> _filter(List<RequestSnapshot> logs) {
    final k = _keyword.trim().toLowerCase();
    if (k.isEmpty) return logs;
    return logs.where((e) {
      final text = '${e.method} ${e.host}${e.path} ${e.status} ${e.target}'.toLowerCase();
      return text.contains(k);
    }).toList();
  }

  @override
  Widget build(BuildContext context) {
    final all = ref.watch(tunnelProvider.select((s) => s.logs));
    final logs = _filter(all);
    return ColoredBox(
      color: AppColors.background,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          PageHeader(
            title: '实时日志',
            subtitle: '本机最近 ${all.length} 条转发记录（最多保留 200 条，仅存于内存）',
            actions: [_searchBox(), _clearButton(all.isEmpty)],
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(28, 4, 28, 28),
              child: Card(child: logs.isEmpty ? _empty(all.isEmpty) : _list(logs)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _searchBox() {
    return SizedBox(
      width: 240,
      child: TextField(
        decoration: const InputDecoration(hintText: '过滤：方法 / 路径 / 状态码', prefixIcon: Icon(Icons.search, size: 18)),
        onChanged: (v) => setState(() => _keyword = v),
      ),
    );
  }

  Widget _clearButton(bool disabled) {
    return OutlinedButton.icon(
      onPressed: disabled ? null : () => ref.read(tunnelProvider.notifier).clearLogs(),
      icon: const Icon(Icons.delete_sweep_outlined, size: 18),
      label: const Text('清空'),
    );
  }

  Widget _empty(bool noData) {
    return EmptyState(
      icon: Icons.receipt_long_outlined,
      title: noData ? '暂无请求' : '没有匹配的记录',
      description: noData ? '连接成功后，经隧道转发的请求会实时出现在这里' : '换个关键字试试',
    );
  }

  /// 以对象身份作 key：新日志插到顶部时只有它创建新 State 并播放入场动画。
  Widget _list(List<RequestSnapshot> logs) {
    return ListView.builder(
      itemCount: logs.length,
      itemBuilder: (context, i) {
        final e = logs[i];
        return FadeSlideIn(key: ObjectKey(e), offset: 10, child: LogRow(entry: e));
      },
    );
  }
}
