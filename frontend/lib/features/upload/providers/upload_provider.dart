import 'dart:io' as io;
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:file_picker/file_picker.dart';
import '../models/upload_models.dart';
import '../../auth/providers/auth_provider.dart';
import '../../items/services/item_api_service.dart';
import '../../shares/providers/share_providers.dart';
import '../../share/providers/share_list_provider.dart';
import '../../../core/api_config_provider.dart';

/// 上传全生命周期 Notifier
class UploadNotifier extends Notifier<UploadState> {
  @override
  UploadState build() => const UploadState();

  /// 选择文件
  Future<void> selectFile() async {
    final result = await FilePicker.pickFiles(
      type: FileType.any,
      allowMultiple: false,
    );

    if (result == null || result.files.isEmpty) return;

    final file = result.files.first;

    // 使用 dart:io 读取系统文件真实属性（Windows 桌面端最可靠方案）
    final filePath = file.path;
    int realBytes;
    if (filePath != null) {
      realBytes = io.File(filePath).lengthSync();
    } else {
      realBytes = file.size;
    }
    final realSizeMB = realBytes / (1024 * 1024);

    state = state.copyWith(
      phase: UploadPhase.fileSelected,
      filePath: filePath,
      fileName: file.name,
      fileSizeMB: realSizeMB,
      fileSizeBytes: realBytes,
    );

    debugPrint('[INFO] 文件已选择: ${file.name} ($realBytes B)');
  }

  /// 设置 Recovery Code 长度（UI 预留，后端暂不支持）
  void setRecoveryCodeLength(int length) {
    state = state.copyWith(recoveryCodeLength: length);
  }

  /// 设置是否公开可见（UI 预留，后端暂不支持）
  void setPublic(bool value) {
    state = state.copyWith(isPublic: value);
  }

  /// 拖拽选择文件（由 upload_page 的 DropTarget 回调调用）
  void setDroppedFile({
    required String filePath,
    required String fileName,
    required double fileSizeMB,
    required int fileSizeBytes,
  }) {
    state = state.copyWith(
      phase: UploadPhase.fileSelected,
      filePath: filePath,
      fileName: fileName,
      fileSizeMB: fileSizeMB,
      fileSizeBytes: fileSizeBytes,
    );
  }

  /// 开始上传 — 对接真实 API `POST v1/protected/item/submit`
  Future<void> startUpload() async {
    if (state.filePath == null) return;

    // 获取认证令牌
    final authState = ref.read(authProvider);
    if (!authState.isLoggedIn || authState.token.isEmpty) {
      state = state.copyWith(
        phase: UploadPhase.error,
        errorMessage: '请先登录后再上传文件',
      );
      return;
    }

    state = state.copyWith(
      phase: UploadPhase.uploading,
      uploadProgress: 0.0,
      errorMessage: null,
    );

    // 模拟上传进度（API不支持真实进度回调）
    _mockProgress();

    try {
      final apiConfig = ref.read(apiConfigProvider);
      final service = ItemApiService(apiConfig.baseUrl);

      final apiResp = await service.submitItem(
        filePath: state.filePath!,
        token: authState.token,
      );

      if (apiResp.isSuccess && apiResp.data != null) {
        debugPrint('[INFO] 上传成功: ${apiResp.data!.name}');

        final result = apiResp.data!;

        // 上传成功后，自动将份额加密存入本地 Hive
        // 即使保存失败也不阻断上传成功流程，确保 Recovery Code 能展示给用户
        try {
          final shareService = ref.read(shareServiceProvider);
          await shareService.saveShare(
            itemId: result.itemId,
            originalFilename: state.fileName ?? 'unknown',
            serverFilename: result.name,
            share: result.share,
            recoveryCode: result.recoveryCode,
          );

          // 使份额列表缓存失效，下次进入页面时重新加载
          ref.invalidate(shareListProvider);
        } catch (e, stack) {
          debugPrint('[ERROR] 份额保存失败: $e');
          debugPrint('[ERROR] 堆栈: $stack');
        }

        state = state.copyWith(
          phase: UploadPhase.success,
          uploadProgress: 1.0,
          result: result,
        );
      } else {
        debugPrint('[ERROR] 上传失败: code=${apiResp.code}, message=${apiResp.message}');

        state = state.copyWith(
          phase: UploadPhase.error,
          errorMessage: apiResp.message.isNotEmpty ? apiResp.message : '上传失败，请稍后重试',
        );
      }
    } catch (e, stack) {
      debugPrint('[ERROR] UploadNotifier.startUpload 异常: $e');
      debugPrint('[ERROR] 堆栈: $stack');

      state = state.copyWith(
        phase: UploadPhase.error,
        errorMessage: '网络请求失败，请检查网络连接',
      );
    }
  }

  /// 模拟上传进度（10步，每步300ms）
  Future<void> _mockProgress() async {
    for (var i = 1; i <= 10; i++) {
      await Future.delayed(const Duration(milliseconds: 300));
      if (state.phase != UploadPhase.uploading) return;
      // 不覆盖真实成功/失败状态（phase可能已被真实响应改变）
      if (state.phase == UploadPhase.uploading) {
        state = state.copyWith(uploadProgress: i / 10);
      }
    }
  }

  /// 重置回到空闲
  void reset() {
    state = const UploadState();
  }
}

/// 上传 Provider（全局单例，保证状态跨页面保持）
final uploadProvider = NotifierProvider<UploadNotifier, UploadState>(
  UploadNotifier.new,
);
