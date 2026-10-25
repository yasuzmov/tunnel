/// 新建 / 编辑连接配置：填写地址与 Key、测试连通性、设置启动自动连接。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../core/tunnel/probe.dart';
import '../../models/profile.dart';
import '../../providers/app_providers.dart';
import '../../widgets/form_page.dart';
import '../../widgets/probe_result.dart';
import '../../widgets/section_card.dart';

/// 配置编辑页。[profile] 为 null 表示新建。
class ProfileEditPage extends ConsumerStatefulWidget {
  const ProfileEditPage({super.key, this.profile});

  final ConnectionProfile? profile;

  @override
  ConsumerState<ProfileEditPage> createState() => _ProfileEditPageState();
}

class _ProfileEditPageState extends ConsumerState<ProfileEditPage> {
  final _form = GlobalKey<FormState>();
  late final _name = TextEditingController(text: widget.profile?.name ?? '新连接');
  late final _url = TextEditingController(text: widget.profile?.serverUrl ?? 'http://127.0.0.1:8080');
  late final _key = TextEditingController(text: widget.profile?.key ?? '');
  late bool _autoConnect = widget.profile?.autoConnect ?? false;
  bool _showKey = false;
  ProbeResult _probe = const ProbeResult.idle();

  @override
  void dispose() {
    _name.dispose();
    _url.dispose();
    _key.dispose();
    super.dispose();
  }

  /// 基于表单值构造新实例，保存前不修改 provider 中的原对象。
  ConnectionProfile _draft() {
    return ConnectionProfile(
      id: widget.profile?.id ?? '',
      name: _name.text.trim(),
      serverUrl: _url.text.trim(),
      key: _key.text.trim(),
      autoConnect: _autoConnect,
      rules: widget.profile?.rules,
    );
  }

  Future<void> _save() async {
    if (!_form.currentState!.validate()) return;
    await ref.read(profilesProvider.notifier).upsert(_draft());
    if (!mounted) return;
    final editingActive = widget.profile != null && ref.read(tunnelProvider).activeId == widget.profile!.id;
    final tip = editingActive ? '已保存，当前连接需断开重连后生效' : '已保存';
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(tip)));
    Navigator.of(context).pop();
  }

  Future<void> _test() async {
    if (!_form.currentState!.validate()) return;
    setState(() => _probe = const ProbeResult.testing());
    try {
      final info = await testConnectivity(_draft());
      _setProbe(ProbeResult(ProbePhase.success, '联通成功', [
        '记录：${info.name}',
        '心跳往返：${info.pingMs} ms',
        '已绑定域名：${info.domains.isEmpty ? '暂无' : info.domains.join('、')}',
      ]));
    } catch (err) {
      _setProbe(ProbeResult(ProbePhase.failure, '联通失败', [_friendlyError(err)]));
    }
  }

  void _setProbe(ProbeResult r) {
    if (mounted) setState(() => _probe = r);
  }

  /// 把常见异常翻译成可操作的提示，原始信息附在后面便于排查。
  String _friendlyError(Object err) {
    final raw = err.toString();
    if (raw.contains('401')) return 'Key 无效或已被重置（$raw）';
    if (raw.contains('SocketException')) return '无法连接服务端，请检查地址与网络（$raw）';
    if (raw.contains('TimeoutException')) return '等待服务端响应超时（$raw）';
    return raw;
  }

  @override
  Widget build(BuildContext context) {
    final isNew = widget.profile == null;
    return Form(
      key: _form,
      child: FormPage(
        title: isNew ? '新建连接' : '编辑连接',
        subtitle: '地址与 Key 可在管理后台的穿透记录详情页复制',
        sections: [_basicSection(), _testSection(), _optionSection()],
        footer: [
          OutlinedButton(onPressed: () => Navigator.of(context).pop(), child: const Text('取消')),
          FilledButton.icon(onPressed: _save, icon: const Icon(Icons.check, size: 18), label: const Text('保存')),
        ],
      ),
    );
  }

  Widget _basicSection() {
    return SectionCard(
      title: '基本信息',
      icon: Icons.badge_outlined,
      child: Column(
        children: [
          TextFormField(
            controller: _name,
            decoration: const InputDecoration(labelText: '名称', prefixIcon: Icon(Icons.label_outline)),
            validator: (v) => (v ?? '').trim().isEmpty ? '请填写名称' : null,
          ),
          const SizedBox(height: 14),
          TextFormField(
            controller: _url,
            decoration: const InputDecoration(
              labelText: '服务端地址',
              hintText: 'https://tunnel.example.com',
              prefixIcon: Icon(Icons.dns_outlined),
            ),
            validator: _validateUrl,
          ),
          const SizedBox(height: 14),
          TextFormField(
            controller: _key,
            obscureText: !_showKey,
            style: const TextStyle(fontFamily: 'Consolas'),
            decoration: InputDecoration(
              labelText: '接入 Key',
              prefixIcon: const Icon(Icons.key_outlined),
              suffixIcon: IconButton(
                tooltip: _showKey ? '隐藏' : '显示',
                icon: Icon(_showKey ? Icons.visibility_off_outlined : Icons.visibility_outlined),
                onPressed: () => setState(() => _showKey = !_showKey),
              ),
            ),
            validator: (v) => (v ?? '').trim().isEmpty ? '请填写接入 Key' : null,
          ),
        ],
      ),
    );
  }

  String? _validateUrl(String? v) {
    final s = (v ?? '').trim();
    if (s.isEmpty) return '请填写服务端地址';
    final ok = RegExp(r'^(https?|wss?)://[^\s/]+').hasMatch(s);
    return ok ? null : '需以 http:// 或 https:// 开头';
  }

  Widget _testSection() {
    return SectionCard(
      title: '连通性测试',
      icon: Icons.network_check,
      trailing: FilledButton.tonalIcon(
        onPressed: _probe.phase == ProbePhase.testing ? null : _test,
        icon: const Icon(Icons.bolt, size: 18),
        label: const Text('开始测试'),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Text(
            '会完成一次完整握手与心跳往返，但不会开启转发；测试成功后管理后台的联通记录会实时出现本机。',
            style: TextStyle(fontSize: 12.5, color: AppColors.text3, height: 1.5),
          ),
          ProbeResultView(result: _probe),
        ],
      ),
    );
  }

  Widget _optionSection() {
    return SectionCard(
      title: '选项',
      icon: Icons.tune,
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: SwitchListTile(
        value: _autoConnect,
        onChanged: (v) => setState(() => _autoConnect = v),
        title: const Text('启动时自动连接'),
        subtitle: const Text('多个配置勾选时只连接列表中的第一个'),
      ),
    );
  }
}
