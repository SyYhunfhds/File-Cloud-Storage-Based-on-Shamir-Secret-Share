import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../utils/constants.dart';

/// 存储服务类
class StorageService {
  late FlutterSecureStorage _secureStorage;
  late SharedPreferences _prefs;

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

  /// 清除所有存储数据
  Future<void> clearAll() async {
    await clearToken();
    await clearUserId();
    await clearUsername();
  }

  /// 检查是否已登录（通过检查是否有Token）
  bool isLoggedIn() {
    return getToken() != null;
  }
}
