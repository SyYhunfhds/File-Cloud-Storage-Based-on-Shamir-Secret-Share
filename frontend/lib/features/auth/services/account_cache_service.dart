import 'dart:convert';
import 'dart:math';

import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:pointycastle/export.dart';

import '../models/account_cache_model.dart';

/// 账号本地加密缓存服务
///
/// 存储架构（多进程安全）：
/// - 加密密码 → [flutter_secure_storage]，每账号独立 key `acct_pwd_$username`
/// - 元数据   → Hive Box `accounts`，username 为 Hive key
///
/// 加密算法：AES-256-GCM（IV 12B, tag 16B），密钥存 secure_storage `account_cache_aes_key`。
class AccountCacheService {
  static const _pwdPrefix = 'acct_pwd_';
  static const _aesKeyName = 'account_cache_aes_key';
  static const _ivLength = 12;
  static const _tagLength = 16;

  final FlutterSecureStorage _secure = const FlutterSecureStorage();

  // ===========================================================================
  // AES 密钥管理
  // ===========================================================================

  Future<Uint8List> _getOrCreateAesKey() async {
    String? stored = await _secure.read(key: _aesKeyName);
    if (stored != null && stored.length >= 44) {
      // Base64 of 32 bytes
      return base64Decode(stored);
    }

    final key = Uint8List(32);
    final rng = Random.secure();
    for (int i = 0; i < 32; i++) {
      key[i] = rng.nextInt(256);
    }
    await _secure.write(key: _aesKeyName, value: base64Encode(key));
    debugPrint('[AccountCache] 已生成设备级 AES-256 密钥');
    return key;
  }

  // ===========================================================================
  // AES-256-GCM 加密/解密
  // ===========================================================================

  Future<String> _encrypt(String plaintext) async {
    final key = await _getOrCreateAesKey();
    final plainBytes = Uint8List.fromList(utf8.encode(plaintext));
    final iv = _randomBytes(_ivLength);

    final cipher = GCMBlockCipher(AESEngine())
      ..init(
          true,
          AEADParameters(
              KeyParameter(key), _tagLength * 8, iv, Uint8List(0)));

    final encrypted = cipher.process(plainBytes);
    final result = Uint8List(iv.length + encrypted.length);
    result.setAll(0, iv);
    result.setAll(iv.length, encrypted);
    return base64Encode(result);
  }

  Future<String> _decrypt(String ciphertextBase64) async {
    final key = await _getOrCreateAesKey();
    final combined = base64Decode(ciphertextBase64);

    if (combined.length < _ivLength + _tagLength) {
      throw ArgumentError('密文数据不完整');
    }

    final iv = combined.sublist(0, _ivLength);
    final ct = combined.sublist(_ivLength);

    final cipher = GCMBlockCipher(AESEngine())
      ..init(
          false,
          AEADParameters(
              KeyParameter(key), _tagLength * 8, iv, Uint8List(0)));

    final decrypted = cipher.process(ct);
    return utf8.decode(decrypted);
  }

  static Uint8List _randomBytes(int length) {
    final rng = Random.secure();
    final bytes = Uint8List(length);
    for (int i = 0; i < length; i++) {
      bytes[i] = rng.nextInt(256);
    }
    return bytes;
  }

  // ===========================================================================
  // 密码存储 (secure_storage)
  // ===========================================================================

  Future<void> savePassword(String username, String password) async {
    final encrypted = await _encrypt(password);
    await _secure.write(key: '$_pwdPrefix$username', value: encrypted);
  }

  Future<String?> getPassword(String username) async {
    final encrypted = await _secure.read(key: '$_pwdPrefix$username');
    if (encrypted == null) return null;
    try {
      return await _decrypt(encrypted);
    } catch (_) {
      return null;
    }
  }

  Future<void> deletePassword(String username) async {
    await _secure.delete(key: '$_pwdPrefix$username');
  }

  // ===========================================================================
  // 元数据存储 (Hive)
  // ===========================================================================

  Future<List<CachedAccount>> loadAccounts() async {
    final box = await Hive.openBox<Map>('accounts');
    final accounts = box.values
        .map((m) => CachedAccount.fromJson(Map<String, dynamic>.from(m)))
        .toList()
      ..sort((a, b) => b.lastUsedAt.compareTo(a.lastUsedAt));
    return accounts;
  }

  Future<void> saveAccount(CachedAccount account) async {
    final box = await Hive.openBox<Map>('accounts');
    await box.put(account.username, account.toJson());
  }

  Future<void> deleteAccount(String username) async {
    final box = await Hive.openBox<Map>('accounts');
    await box.delete(username);
    await deletePassword(username);
  }

  Future<void> deleteAll() async {
    final box = await Hive.openBox<Map>('accounts');
    final keys = box.keys.toList();
    for (final k in keys) {
      await deletePassword(k as String);
    }
    await box.clear();
  }
}
