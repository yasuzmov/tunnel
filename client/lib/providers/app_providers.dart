/// Riverpod 状态：配置列表、当前隧道与实时日志。
library;

import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../core/tunnel/client.dart';
import '../core/tunnel/events.dart';
import '../models/profile.dart';
import '../services/config_store.dart';

final configStoreProvider = Provider<ConfigStore>((ref) => ConfigStore());

final profilesProvider =
    StateNotifierProvider<ProfilesNotifier, List<ConnectionProfile>>(
  (ref) => ProfilesNotifier(ref.read(configStoreProvider)),
);

/// 维护多套连接配置。
class ProfilesNotifier extends StateNotifier<List<ConnectionProfile>> {
  ProfilesNotifier(this._store) : super(const []) {
    unawaited(reload());
  }

  final ConfigStore _store;
  static const _uuid = Uuid();

  Future<void> reload() async {
    state = await _store.load();
  }

  Future<void> upsert(ConnectionProfile profile) async {
    if (profile.id.isEmpty) {
      profile.id = _uuid.v4();
    }
    final next = [...state];
    final idx = next.indexWhere((e) => e.id == profile.id);
    if (idx >= 0) {
      next[idx] = profile;
    } else {
      next.add(profile);
    }
    state = next;
    await _store.save(state);
  }

  Future<void> remove(String id) async {
    state = state.where((e) => e.id != id).toList();
    await _store.save(state);
  }
}

class TunnelState {
  const TunnelState({
    this.status = TunnelStatus.idle,
    this.message = '',
    this.activeId,
    this.welcome,
    this.logs = const [],
  });

  final TunnelStatus status;
  final String message;
  final String? activeId;
  final WelcomeInfo? welcome;
  final List<RequestSnapshot> logs;

  TunnelState copyWith({
    TunnelStatus? status,
    String? message,
    String? activeId,
    WelcomeInfo? welcome,
    List<RequestSnapshot>? logs,
  }) {
    return TunnelState(
      status: status ?? this.status,
      message: message ?? this.message,
      activeId: activeId ?? this.activeId,
      welcome: welcome ?? this.welcome,
      logs: logs ?? this.logs,
    );
  }
}

final tunnelProvider = StateNotifierProvider<TunnelController, TunnelState>(
  (ref) => TunnelController(),
);

/// 同一时间只保持一条长连接。
class TunnelController extends StateNotifier<TunnelState> {
  TunnelController() : super(const TunnelState());

  TunnelClient? _client;

  Future<void> connect(ConnectionProfile profile) async {
    await disconnect();
    final client = TunnelClient(
      profile: profile,
      onStatus: (status, message) {
        state = state.copyWith(status: status, message: message, activeId: profile.id);
      },
      onWelcome: (info) {
        state = state.copyWith(welcome: info);
      },
      onLog: (snap) {
        final next = [snap, ...state.logs];
        if (next.length > 200) {
          next.removeRange(200, next.length);
        }
        state = state.copyWith(logs: next);
      },
    );
    _client = client;
    unawaited(client.run());
  }

  Future<void> disconnect() async {
    await _client?.stop();
    _client = null;
    state = const TunnelState(status: TunnelStatus.idle, message: '已断开');
  }
}
