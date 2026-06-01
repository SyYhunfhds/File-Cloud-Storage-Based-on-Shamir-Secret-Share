import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/share.dart';

/// 份额状态
class ShareState {
  final List<Share> shares;
  final bool isLoading;
  final String? error;
  final int currentVersion;

  const ShareState({
    this.shares = const [],
    this.isLoading = false,
    this.error,
    this.currentVersion = 1,
  });

  ShareState copyWith({
    List<Share>? shares,
    bool? isLoading,
    String? error,
    int? currentVersion,
  }) {
    return ShareState(
      shares: shares ?? this.shares,
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
      currentVersion: currentVersion ?? this.currentVersion,
    );
  }
}

/// 份额状态管理
class ShareNotifier extends StateNotifier<ShareState> {
  ShareNotifier() : super(const ShareState());

  /// 设置加载状态
  void setLoading(bool loading) {
    state = state.copyWith(isLoading: loading);
  }

  /// 设置份额列表
  void setShares(List<Share> shares) {
    state = state.copyWith(shares: shares, isLoading: false, error: null);
  }

  /// 设置错误信息
  void setError(String error) {
    state = state.copyWith(error: error, isLoading: false);
  }

  /// 更新份额版本
  void updateVersion(int version) {
    state = state.copyWith(currentVersion: version);
  }

  /// 应用扰动更新
  void applyDelta(Map<int, String> delta) {
    final updatedShares = state.shares.map((share) {
      if (delta.containsKey(share.userId)) {
        // 模拟应用扰动
        return Share(
          id: share.id,
          userId: share.userId,
          value: delta[share.userId] ?? share.value,
          version: state.currentVersion + 1,
          createdAt: DateTime.now(),
        );
      }
      return share;
    }).toList();

    state = state.copyWith(
      shares: updatedShares,
      currentVersion: state.currentVersion + 1,
    );
  }
}

/// 份额Provider
final shareProvider = StateNotifierProvider<ShareNotifier, ShareState>((ref) {
  return ShareNotifier();
});
