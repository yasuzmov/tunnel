/// 实时日志中的一行：时间、方法、状态码、请求地址、本地目标与错误信息。
library;

import 'package:flutter/material.dart';

import '../../../app/theme.dart';
import '../../../core/tunnel/events.dart';
import '../../../utils/status_style.dart';
import '../../../widgets/badges.dart';

/// 日志行。悬停时浅色高亮，错误信息以红色副行展示。
class LogRow extends StatefulWidget {
  const LogRow({super.key, required this.entry});

  final RequestSnapshot entry;

  @override
  State<LogRow> createState() => _LogRowState();
}

class _LogRowState extends State<LogRow> {
  bool _hover = false;

  @override
  Widget build(BuildContext context) {
    final e = widget.entry;
    return MouseRegion(
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
        decoration: BoxDecoration(
          color: _hover ? const Color(0xFFF5F8FF) : Colors.transparent,
          border: const Border(bottom: BorderSide(color: AppColors.border)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 70,
              child: Text(formatClock(e.time), style: const TextStyle(fontFamily: 'Consolas', fontSize: 12.5, color: AppColors.text3)),
            ),
            SizedBox(width: 64, child: Align(alignment: Alignment.centerLeft, child: Pill(text: e.method, color: methodColor(e.method), mono: true))),
            SizedBox(
              width: 52,
              child: Align(
                alignment: Alignment.centerLeft,
                child: Pill(text: e.status == 0 ? 'ERR' : '${e.status}', color: statusColor(e.status), mono: true),
              ),
            ),
            Expanded(child: _detail(e)),
          ],
        ),
      ),
    );
  }

  Widget _detail(RequestSnapshot e) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('${e.host}${e.path}', maxLines: 1, overflow: TextOverflow.ellipsis,
            style: const TextStyle(fontSize: 13, color: AppColors.text, fontWeight: FontWeight.w500)),
        const SizedBox(height: 3),
        Row(
          children: [
            const Icon(Icons.subdirectory_arrow_right, size: 13, color: AppColors.text3),
            const SizedBox(width: 4),
            Flexible(
              child: Text(e.target.isEmpty ? '未匹配规则' : e.target, maxLines: 1, overflow: TextOverflow.ellipsis,
                  style: const TextStyle(fontSize: 12, color: AppColors.text3, fontFamily: 'Consolas')),
            ),
          ],
        ),
        if (e.error.isNotEmpty) ...[
          const SizedBox(height: 3),
          Text(e.error, style: const TextStyle(fontSize: 12, color: AppColors.danger)),
        ],
      ],
    );
  }
}
