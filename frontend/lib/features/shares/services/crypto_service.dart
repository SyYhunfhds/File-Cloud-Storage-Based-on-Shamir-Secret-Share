import 'dart:convert';
import 'dart:math';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:pointycastle/export.dart';

/// 密钥类型：share 和 recoveryCode 使用独立的 AES-256-GCM 密钥
enum CryptoKeyType { share, recovery }

/// AES-256-GCM 加密/解密服务
///
/// - 密钥存储在 [flutter_secure_storage] 中，按 userId 隔离（不同用户不同密钥）
/// - share 和 recoveryCode 使用不同的 AES 密钥（api.md L293 要求）
/// - 首次调用自动生成 256-bit 随机密钥并持久化
/// - 密文格式: Base64(IV[12] || ciphertext || tag[16])
class ShareCryptoService {
  static const _shareKeyPrefix = 'share_aes_key_';
  static const _recoveryKeyPrefix = 'recovery_aes_key_';
  static const _ivLength = 12;
  static const _tagLength = 16;

  static final _secureStorage = FlutterSecureStorage();

  // ===========================================================================
  // 密钥管理
  // ===========================================================================

  /// 读取或生成 AES-256 密钥（按 userId + keyType 隔离）
  static Future<Uint8List> getOrCreateKey(
    String userId,
    CryptoKeyType type,
  ) async {
    final prefix =
        type == CryptoKeyType.share ? _shareKeyPrefix : _recoveryKeyPrefix;
    final keyName = '$prefix$userId';
    String? storedKey = await _secureStorage.read(key: keyName);

    if (storedKey != null && storedKey.length >= 32) {
      return base64Decode(storedKey);
    }

    final key = Uint8List(32);
    final rng = Random.secure();
    for (int i = 0; i < 32; i++) {
      key[i] = rng.nextInt(256);
    }

    final keyBase64 = base64Encode(key);
    await _secureStorage.write(key: keyName, value: keyBase64);

    final label = type == CryptoKeyType.share ? 'share' : 'recoveryCode';
    debugPrint('[Crypto] 已为用户 $userId 生成新的 AES-256 密钥 ($label)');

    return key;
  }

  // ===========================================================================
  // 加密
  // ===========================================================================

  /// 加密明文，[type] 指定使用 share 密钥还是 recoveryCode 密钥
  static Future<String> encrypt(
    String plaintext,
    String userId, {
    CryptoKeyType type = CryptoKeyType.share,
  }) async {
    final key = await getOrCreateKey(userId, type);
    final plainBytes = Uint8List.fromList(utf8.encode(plaintext));

    final iv = _randomBytes(_ivLength);

    final cipher = GCMBlockCipher(AESEngine())
      ..init(true,
          AEADParameters(KeyParameter(key), _tagLength * 8, iv, Uint8List(0)));

    final encrypted = cipher.process(plainBytes);

    final result = Uint8List(iv.length + encrypted.length);
    result.setAll(0, iv);
    result.setAll(iv.length, encrypted);

    return base64Encode(result);
  }

  // ===========================================================================
  // 解密
  // ===========================================================================

  /// 解密密文，[type] 必须与加密时一致
  static Future<String> decrypt(
    String ciphertextBase64,
    String userId, {
    CryptoKeyType type = CryptoKeyType.share,
  }) async {
    final key = await getOrCreateKey(userId, type);
    final combined = base64Decode(ciphertextBase64);

    if (combined.length < _ivLength + _tagLength) {
      throw ArgumentError('密文数据不完整');
    }

    final iv = combined.sublist(0, _ivLength);
    final ct = combined.sublist(_ivLength);

    final cipher = GCMBlockCipher(AESEngine())
      ..init(false,
          AEADParameters(KeyParameter(key), _tagLength * 8, iv, Uint8List(0)));

    final decrypted = cipher.process(ct);
    return utf8.decode(decrypted);
  }

  // ===========================================================================
  // 辅助
  // ===========================================================================

  static Uint8List _randomBytes(int length) {
    final rng = Random.secure();
    final bytes = Uint8List(length);
    for (int i = 0; i < length; i++) {
      bytes[i] = rng.nextInt(256);
    }
    return bytes;
  }
}
