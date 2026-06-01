import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';
import '../../items/services/item_api_service.dart';
import '../../../core/api_config_provider.dart';

/// 下载状态
class DownloadState {
  final bool isDownloading;
  final String? errorMessage;
  final String? savedPath;

  const DownloadState({
    this.isDownloading = false,
    this.errorMessage,
    this.savedPath,
  });

  DownloadState copyWith({
    bool? isDownloading,
    String? errorMessage,
    String? savedPath,
  }) {
    return DownloadState(
      isDownloading: isDownloading ?? this.isDownloading,
      errorMessage: errorMessage,
      savedPath: savedPath ?? this.savedPath,
    );
  }
}

/// 下载 Notifier
class DownloadNotifier extends Notifier<DownloadState> {
  @override
  DownloadState build() => const DownloadState();

  /// 下载文件
  ///
  /// 返回 true 表示下载成功，false 表示失败。
  /// 成功时文件数据通过 [onFileReceived] 回调传递给 UI 层处理保存。
  Future<bool> download({
    required int itemId,
    required String share,
    String defaultFileName = '',
    void Function(List<int> data, String defaultFileName)? onFileReceived,
  }) async {
    state = state.copyWith(isDownloading: true, errorMessage: null, savedPath: null);

    try {
      final authState = ref.read(authProvider);
      if (!authState.isLoggedIn || authState.token.isEmpty) {
        state = state.copyWith(
          isDownloading: false,
          errorMessage: '请先登录',
        );
        return false;
      }

      final apiConfig = ref.read(apiConfigProvider);
      final service = ItemApiService(apiConfig.baseUrl);

      final result = await service.downloadItem(
        itemId: itemId,
        share: share,
        token: authState.token,
      );

      if (result.data != null) {
        debugPrint('[INFO] 下载成功: ${result.data!.length} 字节');
        final fileName =
            defaultFileName.isNotEmpty ? defaultFileName : 'item_$itemId';
        onFileReceived?.call(result.data!, fileName);

        state = state.copyWith(isDownloading: false);
        return true;
      } else {
        debugPrint('[ERROR] 下载失败: ${result.errorMessage}');

        state = state.copyWith(
          isDownloading: false,
          errorMessage: result.errorMessage ?? '下载失败',
        );
        return false;
      }
    } catch (e, stack) {
      debugPrint('[ERROR] DownloadNotifier.download 异常: $e');
      debugPrint('[ERROR] 堆栈: $stack');

      state = state.copyWith(
        isDownloading: false,
        errorMessage: '网络请求失败，请检查网络连接',
      );
      return false;
    }
  }

  /// 清除状态
  void reset() {
    state = const DownloadState();
  }
}

/// 下载 Provider
final downloadProvider = NotifierProvider<DownloadNotifier, DownloadState>(
  DownloadNotifier.new,
);
