/// 编辑接入地址与 Key，并提供连通性测试。
library;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/tunnel/probe.dart';
import '../models/profile.dart';
import '../providers/app_providers.dart';

class EditProfilePage extends ConsumerStatefulWidget {
  const EditProfilePage({super.key, required this.profile});
  final ConnectionProfile profile;

  @override
  ConsumerState<EditProfilePage> createState() => _EditProfilePageState();
}

class _EditProfilePageState extends ConsumerState<EditProfilePage> {
  late final TextEditingController name;
  late final TextEditingController url;
  late final TextEditingController key;
  String result = '';
  bool testing = false;

  @override
  void initState() {
    super.initState();
    name = TextEditingController(text: widget.profile.name);
    url = TextEditingController(text: widget.profile.serverUrl);
    key = TextEditingController(text: widget.profile.key);
  }

  @override
  void dispose() {
    name.dispose();
    url.dispose();
    key.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    widget.profile
      ..name = name.text.trim()
      ..serverUrl = url.text.trim()
      ..key = key.text.trim();
    await ref.read(profilesProvider.notifier).upsert(widget.profile);
    if (mounted) {
      Navigator.pop(context);
    }
  }

  Future<void> _test() async {
    setState(() {
      testing = true;
      result = '测试中...';
    });
    try {
      final info = await testConnectivity(ConnectionProfile(
        id: widget.profile.id,
        name: name.text,
        serverUrl: url.text,
        key: key.text,
      ));
      setState(() {
        result =
            '联通成功\n记录: ${info.name}\n延迟: ${info.pingMs} ms\n域名: ${info.domains.isEmpty ? '(尚未绑定)' : info.domains.join(', ')}';
      });
    } catch (err) {
      setState(() => result = '联通失败: $err');
    } finally {
      setState(() => testing = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('连接配置')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          TextField(controller: name, decoration: const InputDecoration(labelText: '名称')),
          const SizedBox(height: 12),
          TextField(controller: url, decoration: const InputDecoration(labelText: '服务端地址', hintText: 'http://127.0.0.1:8080')),
          const SizedBox(height: 12),
          TextField(controller: key, decoration: const InputDecoration(labelText: '接入 Key')),
          const SizedBox(height: 20),
          Row(
            children: [
              FilledButton(onPressed: testing ? null : _test, child: const Text('测试连通性')),
              const SizedBox(width: 12),
              FilledButton.tonal(onPressed: _save, child: const Text('保存')),
            ],
          ),
          const SizedBox(height: 16),
          Text(result),
        ],
      ),
    );
  }
}
