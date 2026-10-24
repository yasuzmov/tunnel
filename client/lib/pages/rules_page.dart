/// 本地转发规则编辑。目标可以是本机端口或 hosts 中的域名。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../core/tunnel/probe.dart';
import '../models/profile.dart';
import '../models/rule.dart';
import '../providers/app_providers.dart';
import '../services/hosts_parser.dart';

class RulesPage extends ConsumerStatefulWidget {
  const RulesPage({super.key, required this.profile});
  final ConnectionProfile profile;

  @override
  ConsumerState<RulesPage> createState() => _RulesPageState();
}

class _RulesPageState extends ConsumerState<RulesPage> {
  @override
  Widget build(BuildContext context) {
    final profiles = ref.watch(profilesProvider);
    final profile = profiles.firstWhere(
      (e) => e.id == widget.profile.id,
      orElse: () => widget.profile,
    );
    return Scaffold(
      appBar: AppBar(title: Text('转发规则 - ${profile.name}')),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _edit(profile, ForwardRule(id: const Uuid().v4())),
        child: const Icon(Icons.add),
      ),
      body: ListView.builder(
        itemCount: profile.rules.length,
        itemBuilder: (context, i) {
          final rule = profile.rules[i];
          return ListTile(
            title: Text(rule.label),
            subtitle: Text(rule.enabled ? '已启用' : '已停用'),
            trailing: IconButton(
              icon: const Icon(Icons.delete_outline),
              onPressed: () async {
                profile.rules.removeAt(i);
                await ref.read(profilesProvider.notifier).upsert(profile);
              },
            ),
            onTap: () => _edit(profile, rule),
          );
        },
      ),
    );
  }

  Future<void> _edit(ConnectionProfile profile, ForwardRule rule) async {
    await Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => RuleEditorPage(profile: profile, rule: rule)),
    );
    setState(() {});
  }
}

class RuleEditorPage extends ConsumerStatefulWidget {
  const RuleEditorPage({super.key, required this.profile, required this.rule});
  final ConnectionProfile profile;
  final ForwardRule rule;

  @override
  ConsumerState<RuleEditorPage> createState() => _RuleEditorPageState();
}

class _RuleEditorPageState extends ConsumerState<RuleEditorPage> {
  late ForwardRule rule;
  List<HostsEntry> hosts = [];
  String probe = '';

  @override
  void initState() {
    super.initState();
    rule = widget.rule;
    readHostsFile().then((v) => setState(() => hosts = v));
  }

  Future<void> _save() async {
    final list = [...widget.profile.rules];
    final idx = list.indexWhere((e) => e.id == rule.id);
    if (idx >= 0) {
      list[idx] = rule;
    } else {
      list.add(rule);
    }
    widget.profile.rules = list;
    await ref.read(profilesProvider.notifier).upsert(widget.profile);
    if (mounted) {
      Navigator.pop(context);
    }
  }

  Future<void> _probe() async {
    try {
      final msg = await testLocalTarget(rule);
      setState(() => probe = msg);
    } catch (err) {
      setState(() => probe = '失败: $err');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('编辑规则')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          TextFormField(
            initialValue: rule.matchHost,
            decoration: const InputDecoration(labelText: '匹配域名（空=任意）'),
            onChanged: (v) => rule.matchHost = v.trim(),
          ),
          TextFormField(
            initialValue: rule.pathPrefix,
            decoration: const InputDecoration(labelText: '路径前缀（空=全部）'),
            onChanged: (v) => rule.pathPrefix = v.trim(),
          ),
          const SizedBox(height: 12),
          InputDecorator(
            decoration: const InputDecoration(labelText: '目标类型'),
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                isExpanded: true,
                value: rule.targetType,
                items: const [
                  DropdownMenuItem(value: 'port', child: Text('本机端口')),
                  DropdownMenuItem(value: 'hostsDomain', child: Text('hosts 域名')),
                ],
                onChanged: (v) => setState(() => rule.targetType = v ?? 'port'),
              ),
            ),
          ),
          TextFormField(
            initialValue: '${rule.port}',
            decoration: const InputDecoration(labelText: '端口'),
            keyboardType: TextInputType.number,
            onChanged: (v) => rule.port = int.tryParse(v) ?? rule.port,
          ),
          if (rule.targetType == 'hostsDomain')
            InputDecorator(
              decoration: const InputDecoration(labelText: 'hosts 中的本地域名'),
              child: DropdownButtonHideUnderline(
                child: DropdownButton<String>(
                  isExpanded: true,
                  hint: const Text('选择本机 hosts 记录'),
                  value: hosts.any((e) => e.hostname == rule.hostsDomain) ? rule.hostsDomain : null,
                  items: hosts
                      .map((e) => DropdownMenuItem(value: e.hostname, child: Text('${e.hostname} (${e.ip})')))
                      .toList(),
                  onChanged: (v) => setState(() => rule.hostsDomain = v ?? ''),
                ),
              ),
            ),
          SwitchListTile(
            title: const Text('本地 HTTPS'),
            value: rule.useHttps,
            onChanged: (v) => setState(() => rule.useHttps = v),
          ),
          SwitchListTile(
            title: const Text('忽略证书错误'),
            value: rule.ignoreTlsError,
            onChanged: (v) => setState(() => rule.ignoreTlsError = v),
          ),
          SwitchListTile(
            title: const Text('改写 Host 头'),
            value: rule.rewriteHost,
            onChanged: (v) => setState(() => rule.rewriteHost = v),
          ),
          if (rule.rewriteHost)
            TextFormField(
              initialValue: rule.rewriteHostValue,
              decoration: const InputDecoration(labelText: 'Host 改写值（空则用目标主机）'),
              onChanged: (v) => rule.rewriteHostValue = v.trim(),
            ),
          const SizedBox(height: 16),
          Row(
            children: [
              FilledButton(onPressed: _save, child: const Text('保存')),
              const SizedBox(width: 12),
              FilledButton.tonal(onPressed: _probe, child: const Text('测试本地目标')),
            ],
          ),
          const SizedBox(height: 12),
          Text(probe),
        ],
      ),
    );
  }
}
