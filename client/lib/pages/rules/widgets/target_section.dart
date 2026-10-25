/// 规则编辑页的「转发目标」分组：目标类型切换、端口、hosts 域名选择。
library;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../app/theme.dart';
import '../../../models/rule.dart';
import '../../../services/hosts_parser.dart';
import '../../../widgets/section_card.dart';

/// 转发目标分组。直接修改 [rule]（编辑页传入的是副本），变化后调用 [onChanged] 触发重建。
class TargetSection extends StatelessWidget {
  const TargetSection({super.key, required this.rule, required this.hosts, required this.onChanged});

  final ForwardRule rule;

  /// 本机 hosts 文件解析结果；为空时提示用户先编辑 hosts。
  final List<HostsEntry> hosts;
  final VoidCallback onChanged;

  @override
  Widget build(BuildContext context) {
    final isHosts = rule.targetType == 'hostsDomain';
    return SectionCard(
      title: '转发目标',
      icon: Icons.output_rounded,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          SegmentedButton<String>(
            segments: const [
              ButtonSegment(value: 'port', icon: Icon(Icons.settings_ethernet), label: Text('本机端口')),
              ButtonSegment(value: 'hostsDomain', icon: Icon(Icons.dns_outlined), label: Text('hosts 域名')),
            ],
            selected: {rule.targetType},
            showSelectedIcon: false,
            onSelectionChanged: (v) {
              rule.targetType = v.first;
              onChanged();
            },
          ),
          AnimatedSize(
            duration: const Duration(milliseconds: 320),
            curve: Curves.easeOutCubic,
            alignment: Alignment.topCenter,
            child: isHosts ? _hostsPicker() : const SizedBox(width: double.infinity),
          ),
          const SizedBox(height: 14),
          TextFormField(
            initialValue: '${rule.port}',
            decoration: const InputDecoration(labelText: '端口', prefixIcon: Icon(Icons.numbers)),
            keyboardType: TextInputType.number,
            inputFormatters: [FilteringTextInputFormatter.digitsOnly, LengthLimitingTextInputFormatter(5)],
            validator: _validatePort,
            onChanged: (v) {
              rule.port = int.tryParse(v) ?? rule.port;
              onChanged();
            },
          ),
          const SizedBox(height: 10),
          _preview(),
        ],
      ),
    );
  }

  String? _validatePort(String? v) {
    final n = int.tryParse(v ?? '');
    return n == null || n < 1 || n > 65535 ? '端口范围 1~65535' : null;
  }

  Widget _hostsPicker() {
    final selected = hosts.any((e) => e.hostname == rule.hostsDomain) ? rule.hostsDomain : null;
    return Padding(
      padding: const EdgeInsets.only(top: 14),
      child: InputDecorator(
        decoration: InputDecoration(
          labelText: 'hosts 中的本地域名',
          prefixIcon: const Icon(Icons.language),
          helperText: hosts.isEmpty ? '未读取到 hosts 记录，请先在 C:\\Windows\\System32\\drivers\\etc\\hosts 中添加' : null,
        ),
        child: DropdownButtonHideUnderline(
          child: DropdownButton<String>(
            isExpanded: true,
            isDense: true,
            borderRadius: BorderRadius.circular(10),
            hint: const Text('选择本机 hosts 记录'),
            value: selected,
            items: [
              for (final e in hosts)
                DropdownMenuItem(value: e.hostname, child: Text('${e.hostname}  (${e.ip})')),
            ],
            onChanged: (v) {
              rule.hostsDomain = v ?? '';
              onChanged();
            },
          ),
        ),
      ),
    );
  }

  /// 实时预览最终目标地址，让用户确认协议、主机与端口组合是否正确。
  Widget _preview() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(color: const Color(0xFFF5F8FF), borderRadius: BorderRadius.circular(8)),
      child: Row(
        children: [
          const Text('目标地址', style: TextStyle(fontSize: 12, color: AppColors.text3)),
          const SizedBox(width: 12),
          Expanded(
            child: Text(rule.targetUri(), style: const TextStyle(fontFamily: 'Consolas', color: AppColors.brand)),
          ),
        ],
      ),
    );
  }
}
