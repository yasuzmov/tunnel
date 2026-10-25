/// 关于分区：版本、配置文件位置与工作原理简介。
library;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/app_info.dart';
import '../../app/theme.dart';
import '../../providers/app_providers.dart';
import '../../widgets/fade_slide_in.dart';
import '../../widgets/page_header.dart';
import '../../widgets/section_card.dart';

class AboutPage extends ConsumerWidget {
  const AboutPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return ColoredBox(
      color: AppColors.background,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const PageHeader(title: '关于', subtitle: '$kAppName v$kClientVersion'),
          Expanded(
            child: ListView(
              padding: const EdgeInsets.fromLTRB(28, 4, 28, 28),
              children: [
                FadeSlideIn(child: _configCard(context, ref)),
                const SizedBox(height: 16),
                const FadeSlideIn(index: 1, child: _FlowCard()),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _configCard(BuildContext context, WidgetRef ref) {
    final path = ref.watch(_configPathProvider).value;
    return SectionCard(
      title: '本地配置',
      icon: Icons.folder_outlined,
      child: Row(
        children: [
          Expanded(child: SelectableText(path ?? '读取中…', style: const TextStyle(fontFamily: 'Consolas', fontSize: 13))),
          IconButton(
            tooltip: '复制路径',
            icon: const Icon(Icons.copy_rounded, size: 18),
            onPressed: path == null
                ? null
                : () {
                    Clipboard.setData(ClipboardData(text: path));
                    ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('已复制')));
                  },
          ),
        ],
      ),
    );
  }
}

/// 配置文件路径只需解析一次，缓存在 provider 中避免每次重建都触发 IO。
final _configPathProvider = FutureProvider<String>((ref) => ref.read(configStoreProvider).filePath());

/// 请求链路示意：公网 → nginx → 服务端 → 隧道 → 本客户端 → 本地服务。
class _FlowCard extends StatelessWidget {
  const _FlowCard();

  static const _steps = [
    (Icons.public, '公网请求'),
    (Icons.dns_outlined, 'nginx'),
    (Icons.cloud_outlined, '服务端'),
    (Icons.hub_outlined, 'WebSocket 隧道'),
    (Icons.computer, '本客户端'),
    (Icons.storage_outlined, '本地服务'),
  ];

  @override
  Widget build(BuildContext context) {
    return SectionCard(
      title: '工作原理',
      icon: Icons.account_tree_outlined,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            crossAxisAlignment: WrapCrossAlignment.center,
            runSpacing: 10,
            children: [
              for (var i = 0; i < _steps.length; i++) ...[
                if (i > 0) const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 6),
                  child: Icon(Icons.arrow_forward_rounded, size: 16, color: AppColors.text3),
                ),
                _node(_steps[i].$1, _steps[i].$2),
              ],
            ],
          ),
          const SizedBox(height: 16),
          const Text(
            '客户端主动连接服务端并保持一条多路复用的 WebSocket 长连接，无需公网 IP 或端口映射。'
            '请求按转发规则自上而下匹配，命中后转给本机端口或 hosts 中配置的本地域名。',
            style: TextStyle(fontSize: 13, color: AppColors.text2, height: 1.7),
          ),
        ],
      ),
    );
  }

  Widget _node(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(color: const Color(0xFFF5F8FF), borderRadius: BorderRadius.circular(10)),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: AppColors.brand),
          const SizedBox(width: 6),
          Text(label, style: const TextStyle(fontSize: 13, color: AppColors.text)),
        ],
      ),
    );
  }
}
