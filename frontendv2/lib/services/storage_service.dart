import 'dart:convert';

import 'package:encrypt/encrypt.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/share.dart';
import '../utils/constants.dart';

/// 存储服务类
class StorageService {
  late FlutterSecureStorage _secureStorage;
  late SharedPreferences _prefs;

  // TODO: [安全改进] 请使用更安全的密钥存储方式（如设备密钥链或用户密码派生密钥）
  // 当前使用硬编码密钥仅用于开发测试
  static const String _encryptionKeyBase64 = 'temporary-aes-key-32-bytes-for-dev-only!';

  /// 初始化存储服务
  Future<void> init() async {
    _secureStorage = const FlutterSecureStorage();
    _prefs = await SharedPreferences.getInstance();
  }

  /// 保存Token
  Future<void> saveToken(String token) async {
    await _secureStorage.write(
      key: Constants.storageTokenKey,
      value: token,
    );
  }

  /// 获取Token
  Future<String?> getToken() async {
    return await _secureStorage.read(key: Constants.storageTokenKey);
  }

  /// 清除Token
  Future<void> clearToken() async {
    await _secureStorage.delete(key: Constants.storageTokenKey);
  }

  /// 保存用户ID
  Future<void> saveUserId(int userId) async {
    await _prefs.setInt(Constants.storageUserIdKey, userId);
  }

  /// 获取用户ID
  int? getUserId() {
    return _prefs.getInt(Constants.storageUserIdKey);
  }

  /// 清除用户ID
  Future<void> clearUserId() async {
    await _prefs.remove(Constants.storageUserIdKey);
  }

  /// 保存用户名
  Future<void> saveUsername(String username) async {
    await _prefs.setString(Constants.storageUsernameKey, username);
  }

  /// 获取用户名
  String? getUsername() {
    return _prefs.getString(Constants.storageUsernameKey);
  }

  /// 清除用户名
  Future<void> clearUsername() async {
    await _prefs.remove(Constants.storageUsernameKey);
  }

  /// 保存份额到本地（使用AES-GCM加密）
  Future<void> saveShare(LocalShare share) async {
    final shares = await getShares();
    
    final existingIndex = shares.indexWhere((s) => s.filename == share.filename);
    if (existingIndex >= 0) {
      shares[existingIndex] = share;
    } else {
      shares.add(share);
    }
    
    final jsonString = jsonEncode(shares.map((s) => s.toJson()).toList());
    final encrypted = _aesEncrypt(jsonString);
    await _prefs.setString(Constants.storageSharesKey, encrypted);
  }

  /// 获取份额列表（AES-GCM解密）
  Future<List<LocalShare>> getShares() async {
    final encrypted = _prefs.getString(Constants.storageSharesKey);
    if (encrypted == null) return [];
    
    try {
      final jsonString = _aesDecrypt(encrypted);
      final List<dynamic> list = jsonDecode(jsonString) as List<dynamic>;
      return list.map((item) => LocalShare.fromJson(item as Map<String, dynamic>)).toList();
    } catch (e, stackTrace) {
      print('[StorageService] Failed to decrypt shares: $e');
      print('[StorageService] Stack trace: $stackTrace');
      print('[StorageService] Encrypted data (first 100 chars): ${encrypted.substring(0, encrypted.length > 100 ? 100 : encrypted.length)}');
      return [];
    }
  }

  /// 清除份额列表
  Future<void> clearShares() async {
    await _prefs.remove(Constants.storageSharesKey);
  }

  /// 清除所有存储数据
  Future<void> clearAll() async {
    await clearToken();
    await clearUserId();
    await clearUsername();
    await clearShares();
  }

  /// 检查是否已登录（通过检查是否有Token）
  bool isLoggedIn() {
    return getToken() != null;
  }

  /// AES-GCM加密
  String _aesEncrypt(String plaintext) {
    final key = Key.fromUtf8(_encryptionKeyBase64.padRight(32).substring(0, 32));
    final iv = IV.fromSecureRandom(12);
    final encrypter = Encrypter(AES(key, mode: AESMode.gcm));
    
    final encrypted = encrypter.encrypt(plaintext, iv: iv);
    
    return '${iv.base64}:${encrypted.base64}';
  }

  /// AES-GCM解密
  String _aesDecrypt(String ciphertext) {
    final parts = ciphertext.split(':');
    if (parts.length != 2) {
      throw const FormatException('Invalid encrypted format');
    }
    
    final iv = IV.fromBase64(parts[0]);
    final encryptedData = Encrypted.fromBase64(parts[1]);
    
    final key = Key.fromUtf8(_encryptionKeyBase64.padRight(32).substring(0, 32));
    final encrypter = Encrypter(AES(key, mode: AESMode.gcm));
    
    return encrypter.decrypt(encryptedData, iv: iv);
  }
}