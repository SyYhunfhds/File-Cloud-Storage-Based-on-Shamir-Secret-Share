import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/api_config_provider.dart';
import '../../auth/providers/auth_provider.dart';
import '../../shares/services/crypto_service.dart';
import '../../shares/services/share_refresh_service.dart';
import '../../shares/services/share_storage_service.dart';
import '../../shares/models/share_refresh_models.dart';
import '../../shares/models/share_record_data.dart';

/// 份额刷新进度状态
class ShareRefreshState {
  final bool isRefreshing;
  final int progress;
  final String currentMessage;
  final List<String> logLines;
  final String? errorMessage;
  final ShareRefreshRes? result;

  const ShareRefreshState({
    this.isRefreshing = false,
    this.progress = 0,
    this.currentMessage = '',
    this.logLines = const [],
    this.errorMessage,
    this.result,
  });

  ShareRefreshState copyWith({
    bool? isRefreshing,
    int? progress,
    String? currentMessage,
    List<String>? logLines,
    String? errorMessage,
    ShareRefreshRes? result,
    bool clearError = false,
  }) {
    return ShareRefreshState(
      isRefreshing: isRefreshing ?? this.isRefreshing,
      progress: progress ?? this.progress,
      currentMessage: currentMessage ?? this.currentMessage,
      logLines: logLines ?? this.logLines,
      errorMessage: clearError ? null : (errorMessage ?? this.errorMessage),
      result: result ?? this.result,
    );
  }
}

/// 份额刷新进度 Provider
final shareRefreshProvider =
    NotifierProvider<ShareRefreshNotifier, ShareRefreshState>(
  ShareRefreshNotifier.new,
);

class ShareRefreshNotifier extends Notifier<ShareRefreshState> {
  ShareRefreshService? _service;

  @override
  ShareRefreshState build() => const ShareRefreshState();

  /// 启动份额刷新
  Future<void> startRefresh({
    required int itemId,
    String? deviceShare,
    String? recoveryCode,
  }) async {
    final auth = ref.read(authProvider);
    if (!auth.isLoggedIn) {
      state = state.copyWith(
        errorMessage: '请先登录',
        isRefreshing: false,
      );
      return;
    }

    state = ShareRefreshState(isRefreshing: true);

    final config = ref.read(apiConfigProvider);
    _service = ShareRefreshService(config.baseUrl);

    await _service!.refresh(
      token: auth.token,
      itemId: itemId,
      deviceShare: deviceShare,
      recoveryCode: recoveryCode,
      onProgress: (msg) {
        state = state.copyWith(
          progress: msg.progress,
          currentMessage: msg.message,
          logLines: [...state.logLines, msg.message],
          result: msg.data,
        );
        if (msg.progress >= 100 || msg.data != null) {
          state = state.copyWith(isRefreshing: false);
          _saveRefreshedShare(itemId, msg.data, auth.userName);
        }
      },
      onError: (error) {
        state = state.copyWith(
          errorMessage: error,
          isRefreshing: false,
        );
      },
    );

    // 流自然结束但未收到 progress=100 → 异常
    if (state.isRefreshing) {
      state = state.copyWith(
        errorMessage: '份额刷新异常中断，请联系管理员',
        isRefreshing: false,
      );
    }
  }

  /// 保存刷新后的份额到 Hive
  Future<void> _saveRefreshedShare(
      int itemId, ShareRefreshRes? result, String userName) async {
    if (result == null) return;

    try {
      final shareStorage = ShareStorageService();
      final existing = await shareStorage.get(userName, itemId);
      if (existing == null) {
        debugPrint(
            '[ShareRefresh] itemId=$itemId 的旧份额不存在，跳过保存');
        return;
      }

      final encryptedShare =
          await ShareCryptoService.encrypt(result.deviceShare, userName);
      final encryptedRecovery = await ShareCryptoService.encrypt(
        result.recoveryCode,
        userName,
        type: CryptoKeyType.recovery,
      );

      final updated = ShareRecordData(
        itemId: existing.itemId,
        originalFilename: existing.originalFilename,
        serverFilename: existing.serverFilename,
        encryptedShare: encryptedShare,
        encryptedRecoveryCode: encryptedRecovery,
        createdAt: existing.createdAt,
      );

      await shareStorage.save(userName, updated);
      debugPrint('[ShareRefresh] itemId=$itemId 新份额已保存');
    } catch (e) {
      debugPrint('[ShareRefresh] 保存刷新后份额失败: $e');
    }
  }

  /// 取消刷新
  void cancel() {
    _service?.cancel();
    _service = null;
    state = ShareRefreshState(errorMessage: '已取消');
  }
}