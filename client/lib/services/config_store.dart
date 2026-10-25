/// 把连接配置存成 JSON 文件，避免引入数据库依赖。
library;

import 'dart:convert';
import 'dart:io';

import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

import '../models/profile.dart';

/// 配置仓库。
class ConfigStore {
  File? _file;

  Future<File> _ensureFile() async {
    if (_file != null) {
      return _file!;
    }
    final dir = await getApplicationSupportDirectory();
    _file = File(p.join(dir.path, 'profiles.json'));
    return _file!;
  }

  /// 配置文件绝对路径，供「关于」页展示，便于用户手动备份。
  Future<String> filePath() async => (await _ensureFile()).path;

  /// 读取全部配置，文件不存在时返回空列表。
  Future<List<ConnectionProfile>> load() async {
    final file = await _ensureFile();
    if (!await file.exists()) {
      return [];
    }
    final text = await file.readAsString();
    if (text.trim().isEmpty) {
      return [];
    }
    final raw = jsonDecode(text) as List<dynamic>;
    return raw
        .map((e) => ConnectionProfile.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// 覆盖写入。
  Future<void> save(List<ConnectionProfile> profiles) async {
    final file = await _ensureFile();
    await file.parent.create(recursive: true);
    await file.writeAsString(
      const JsonEncoder.withIndent('  ').convert(
        profiles.map((e) => e.toJson()).toList(),
      ),
    );
  }
}
