/// 新建 / 编辑单条转发规则：匹配条件、转发目标、高级选项与本地目标测试。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../app/theme.dart';
import '../../core/tunnel/probe.dart';
import '../../models/rule.dart';
import '../../providers/app_providers.dart';
import '../../services/hosts_parser.dart';
import '../../widgets/form_page.dart';
import '../../widgets/probe_result.dart';
import '../../widgets/section_card.dart';
import 'widgets/target_section.dart';

/// 规则编辑页。[rule] 为 null 表示新增；编辑时基于副本修改，取消不影响原规则。
class RuleEditorPage extends ConsumerStatefulWidget {
  const RuleEditorPage({super.key, required this.profileId, this.rule});

  final String profileId;
  final ForwardRule? rule;

  @override
  ConsumerState<RuleEditorPage> createState() => _RuleEditorPageState();
}

class _RuleEditorPageState extends ConsumerState<RuleEditorPage> {
  final _form = GlobalKey<FormState>();
  late final ForwardRule _rule = widget.rule?.copy() ?? ForwardRule(id: const Uuid().v4());
  List<HostsEntry> _hosts = const [];
  ProbeResult _probe = const ProbeResult.idle();

  @override
  void initState() {
    super.initState();
    readHostsFile().then((v) {
      if (mounted) setState(() => _hosts = v);
    });
  }

  void _touch() => setState(() {});

  Future<void> _save() async {
    if (!_form.currentState!.validate()) return;
    if (_rule.targetType == 'hostsDomain' && _rule.hostsDomain.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('请选择 hosts 中的本地域名')));
      return;
    }
    final profiles = ref.read(profilesProvider);
    final profile = profiles.where((p) => p.id == widget.profileId).firstOrNull;
    if (profile == null) return;
    final rules = [...profile.rules];
    final idx = rules.indexWhere((r) => r.id == _rule.id);
    if (idx >= 0) {
      rules[idx] = _rule;
    } else {
      rules.add(_rule);
    }
    await ref.read(profilesProvider.notifier).updateRules(widget.profileId, rules);
    if (mounted) Navigator.of(context).pop();
  }

  Future<void> _test() async {
    setState(() => _probe = const ProbeResult.testing());
    try {
      final msg = await testLocalTarget(_rule);
      _setProbe(ProbeResult(ProbePhase.success, '本地服务可达', ['${_rule.targetUri()}/  →  $msg']));
    } catch (err) {
      _setProbe(ProbeResult(ProbePhase.failure, '本地服务不可达', ['$err']));
    }
  }

  void _setProbe(ProbeResult r) {
    if (mounted) setState(() => _probe = r);
  }

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _form,
      child: FormPage(
        title: widget.rule == null ? '新增规则' : '编辑规则',
        subtitle: '公网请求命中匹配条件后，会被转发到下方的本地目标',
        sections: [
          _matchSection(),
          TargetSection(rule: _rule, hosts: _hosts, onChanged: _touch),
          _advancedSection(),
          _testSection(),
        ],
        footer: [
          OutlinedButton(onPressed: () => Navigator.of(context).pop(), child: const Text('取消')),
          FilledButton.icon(onPressed: _save, icon: const Icon(Icons.check, size: 18), label: const Text('保存')),
        ],
      ),
    );
  }

  Widget _matchSection() {
    return SectionCard(
      title: '匹配条件',
      icon: Icons.filter_alt_outlined,
      child: Row(
        children: [
          Expanded(
            child: TextFormField(
              initialValue: _rule.matchHost,
              decoration: const InputDecoration(labelText: '域名', hintText: '留空匹配任意域名', prefixIcon: Icon(Icons.language)),
              onChanged: (v) => _rule.matchHost = v.trim(),
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: TextFormField(
              initialValue: _rule.pathPrefix,
              decoration: const InputDecoration(labelText: '路径前缀', hintText: '例如 /api，留空匹配全部', prefixIcon: Icon(Icons.route)),
              onChanged: (v) => _rule.pathPrefix = v.trim(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _advancedSection() {
    return SectionCard(
      title: '高级选项',
      icon: Icons.tune,
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        children: [
          _switch('本地使用 HTTPS', '本地服务只监听 https 时开启', _rule.useHttps, (v) => _rule.useHttps = v),
          _switch('忽略证书错误', '本地自签名证书时开启', _rule.ignoreTlsError, (v) => _rule.ignoreTlsError = v),
          _switch('改写 Host 头', '本地服务按 Host 区分站点时使用', _rule.rewriteHost, (v) => _rule.rewriteHost = v),
          AnimatedSize(
            duration: const Duration(milliseconds: 300),
            curve: Curves.easeOutCubic,
            child: _rule.rewriteHost
                ? Padding(
                    padding: const EdgeInsets.fromLTRB(16, 4, 16, 14),
                    child: TextFormField(
                      initialValue: _rule.rewriteHostValue,
                      decoration: const InputDecoration(labelText: 'Host 改写值', hintText: '留空则使用目标主机名'),
                      onChanged: (v) => _rule.rewriteHostValue = v.trim(),
                    ),
                  )
                : const SizedBox(width: double.infinity),
          ),
        ],
      ),
    );
  }

  Widget _switch(String title, String sub, bool value, ValueChanged<bool> apply) {
    return SwitchListTile(
      title: Text(title),
      subtitle: Text(sub, style: const TextStyle(fontSize: 12, color: AppColors.text3)),
      value: value,
      onChanged: (v) => setState(() => apply(v)),
    );
  }

  Widget _testSection() {
    return SectionCard(
      title: '测试本地目标',
      icon: Icons.network_check,
      trailing: FilledButton.tonalIcon(
        onPressed: _probe.phase == ProbePhase.testing ? null : _test,
        icon: const Icon(Icons.bolt, size: 18),
        label: const Text('发送 GET /'),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Text('直接从本机请求目标地址，确认服务已启动且规则配置正确。',
              style: TextStyle(fontSize: 12.5, color: AppColors.text3)),
          ProbeResultView(result: _probe),
        ],
      ),
    );
  }
}
