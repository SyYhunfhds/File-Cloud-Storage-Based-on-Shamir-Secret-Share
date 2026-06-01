import 'package:flutter/foundation.dart';
import '../models/share_record_data.dart';
import 'crypto_service.dart';
import 'share_storage_service.dart';

/// 份额存储服务（协调层）
///
/// 封装 [ShareStorageService] + [ShareCryptoService]，提供：
/// - saveShare: 上传后加密存储
/// - getShareForDownload: 下载前解密返回明文
/// - listAll: 分页查询
/// - getDecryptedShare: 开发者模式解密
class ShareService {
  final ShareStorageService _storage;
  final String _userId;

  ShareService(this._storage, {required String userId}) : _userId = userId;

  // ===========================================================================
  // 保存份额（上传成功后调用）
  // ===========================================================================

  Future<void> saveShare({
    required int itemId,
    required String originalFilename,
    required String serverFilename,
    required String share,
    required String recoveryCode,
  }) async {
    debugPrint('[ShareService] 开始加密存储份额: itemId=$itemId');

    try {
      final encryptedShare =
          await ShareCryptoService.encrypt(share, _userId);
      final encryptedRecovery = await ShareCryptoService.encrypt(
        recoveryCode,
        _userId,
        type: CryptoKeyType.recovery,
      );

      final record = ShareRecordData(
        itemId: itemId,
        originalFilename: originalFilename,
        serverFilename: serverFilename,
        encryptedShare: encryptedShare,
        encryptedRecoveryCode: encryptedRecovery,
        createdAt: DateTime.now(),
      );

      await _storage.save(_userId, record);

      debugPrint('[ShareService] 份额已加密存储: itemId=$itemId');
    } catch (e) {
      debugPrint('[ShareService] 保存份额失败: $e');
      rethrow;
    }
  }

  // ===========================================================================
  // 获取份额（下载前调用）
  // ===========================================================================

  Future<String?> getShareForDownload(int itemId) async {
    try {
      final record = await _storage.get(_userId, itemId);
      if (record == null) {
        debugPrint('[ShareService] 未找到 itemId=$itemId 的份额记录');
        return null;
      }

      final plaintext =
          await ShareCryptoService.decrypt(record.encryptedShare, _userId);
      debugPrint('[ShareService] 已解密 itemId=$itemId 的份额');
      return plaintext;
    } catch (e) {
      debugPrint('[ShareService] 解密份额失败: $e');
      return null;
    }
  }

  // ===========================================================================
  // 获取原始份额记录（不含解密）
  // ===========================================================================

  Future<ShareRecordData?> getShareRecord(int itemId) async {
    return _storage.get(_userId, itemId);
  }

  // ===========================================================================
  // 列表查询
  // ===========================================================================

  Future<List<ShareRecordData>> listAll() => _storage.listAll(_userId);

  Future<int> count() => _storage.count(_userId);

  // ===========================================================================
  // 开发者模式：解密份额明文（用于复制到剪贴板）
  // ===========================================================================

  Future<String?> getDecryptedShare(int itemId) async {
    final record = await _storage.get(_userId, itemId);
    if (record == null) {
      debugPrint('[ShareService] getDecryptedShare: itemId=$itemId 未找到');
      return null;
    }
    return ShareCryptoService.decrypt(record.encryptedShare, _userId);
  }
}
