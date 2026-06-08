import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api_config_provider.dart';
import '../services/auth_api_service.dart';
import '../../shares/providers/share_providers.dart';

/// 认证状态
class AuthState {
  final bool isLoggedIn;
  final String token;
  final int userId;
  final String userName;
  final String email;
  final String userPosition;
  final DateTime? registeredAt;
  final int privilege;
  final bool isLoading;
  final String? errorMessage;

  const AuthState({
    this.isLoggedIn = false,
    this.token = '',
    this.userId = 0,
    this.userName = '',
    this.email = '',
    this.userPosition = '',
    this.registeredAt,
    this.privilege = 0,
    this.isLoading = false,
    this.errorMessage,
  });

  AuthState copyWith({
    bool? isLoggedIn,
    String? token,
    int? userId,
    String? userName,
    String? email,
    String? userPosition,
    DateTime? registeredAt,
    int? privilege,
    bool? isLoading,
    String? errorMessage,
    bool clearError = false,
  }) {
    return AuthState(
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
      token: token ?? this.token,
      userId: userId ?? this.userId,
      userName: userName ?? this.userName,
      email: email ?? this.email,
      userPosition: userPosition ?? this.userPosition,
      registeredAt: registeredAt ?? this.registeredAt,
      privilege: privilege ?? this.privilege,
      isLoading: isLoading ?? this.isLoading,
      errorMessage: clearError ? null : (errorMessage ?? this.errorMessage),
    );
  }
}

/// 认证 Notifier（Riverpod 3.x）
class AuthNotifier extends Notifier<AuthState> {
  @override
  AuthState build() => const AuthState();

  AuthApiService _createService() {
    final config = ref.read(apiConfigProvider);
    return AuthApiService(config.baseUrl);
  }

  /// 登录
  Future<bool> login(String username, String email, String password) async {
    debugPrint(
        '[DEBUG] AuthNotifier.login - 开始登录, username: $username, email: $email');
    state = state.copyWith(isLoading: true, clearError: true);

    try {
      final service = _createService();
      debugPrint('[DEBUG] AuthNotifier.login - 调用 service.login()');
      final resp = await service.login(
        password: password,
        username: username,
        email: email,
      );

      debugPrint(
          '[DEBUG] AuthNotifier.login - API响应: code=${resp.code}, message=${resp.message}, isSuccess=${resp.isSuccess}, data=${resp.data}');

      if (!resp.isSuccess) {
        debugPrint('[DEBUG] AuthNotifier.login - 登录失败: ${resp.message}');
        state = state.copyWith(
          isLoading: false,
          errorMessage: resp.message.isNotEmpty ? resp.message : '登录失败',
        );
        return false;
      }

      final token = resp.data ?? '';
      debugPrint(
          '[DEBUG] AuthNotifier.login - 获取到token: ${token.isNotEmpty ? '***' : '空'}');
      if (token.isEmpty) {
        debugPrint('[DEBUG] AuthNotifier.login - token为空');
        state = state.copyWith(
          isLoading: false,
          errorMessage: '未获取到认证凭据',
        );
        return false;
      }

      // 从 JWT payload 提取 userId（fallback，user/me 可能无 id 字段）
      int jwtUserId = 0;
      try {
        final parts = token.split('.');
        if (parts.length >= 2) {
          final payload = utf8.decode(
            base64Url.decode(base64Url.normalize(parts[1])),
          );
          final payloadJson = jsonDecode(payload) as Map<String, dynamic>;
          jwtUserId = (payloadJson['Id'] as num?)?.toInt() ?? 0;
        }
      } catch (_) {
        debugPrint('[DEBUG] AuthNotifier.login - JWT解析失败，userId使用user/me接口的值');
      }

      // 获取用户信息
      debugPrint('[DEBUG] AuthNotifier.login - 准备获取用户信息');
      final result = await _fetchUserInfo(token, jwtUserId: jwtUserId);
      debugPrint('[DEBUG] AuthNotifier.login - 获取用户信息结果: $result');
      return result;
    } catch (e, stackTrace) {
      debugPrint('[ERROR] AuthNotifier.login - 异常: $e');
      debugPrint('[ERROR] 堆栈: $stackTrace');
      state = state.copyWith(
        isLoading: false,
        errorMessage: '连接失败',
      );
      return false;
    }
  }

  /// 注册
  Future<bool> register(
      String username, String password, String email) async {
    state = state.copyWith(isLoading: true, clearError: true);

    try {
      final service = _createService();
      final resp = await service.register(
        username: username,
        password: password,
        email: email,
      );

      if (!resp.isSuccess) {
        state = state.copyWith(
          isLoading: false,
          errorMessage: resp.message.isNotEmpty ? resp.message : '注册失败',
        );
        return false;
      }

      return await login(username, email, password);
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: '连接失败',
      );
      return false;
    }
  }

  /// 登出 — 关闭 Hive Box + 清除状态
  Future<void> logout() async {
    final currentUserId = state.userId;
    final currentUserName = state.userName;
    final currentToken = state.token;

    if (currentToken.isNotEmpty) {
      try {
        final service = _createService();
        await service.logout(currentToken);
      } catch (_) {
        // 即使登出 API 失败，也要清除本地状态
      }
    }

    // 关闭当前用户的 Hive Box（文件保留磁盘）
    if (currentUserName.isNotEmpty) {
      try {
        final storageService = ref.read(shareStorageServiceProvider);
        await storageService.closeBox('${currentUserId}_$currentUserName');
      } catch (_) {
        // 忽略清理错误
      }
    }

    state = const AuthState();
  }

  /// 内部：获取用户信息
  Future<bool> _fetchUserInfo(String token, {int jwtUserId = 0}) async {
    debugPrint('[DEBUG] AuthNotifier._fetchUserInfo - 开始获取用户信息');
    try {
      final service = _createService();
      debugPrint(
          '[DEBUG] AuthNotifier._fetchUserInfo - 调用 service.getUserInfo()');
      final resp = await service.getUserInfo(token);

      debugPrint(
          '[DEBUG] AuthNotifier._fetchUserInfo - API响应: code=${resp.code}, message=${resp.message}, isSuccess=${resp.isSuccess}');

      if (!resp.isSuccess) {
        debugPrint(
            '[DEBUG] AuthNotifier._fetchUserInfo - 获取用户信息失败: ${resp.message}');
        state = state.copyWith(
          isLoading: false,
          errorMessage:
              resp.message.isNotEmpty ? resp.message : '获取用户信息失败',
        );
        return false;
      }

      final info = resp.data;
      debugPrint(
          '[DEBUG] AuthNotifier._fetchUserInfo - 用户信息: ${info?.toString()}');
      if (info == null) {
        debugPrint('[DEBUG] AuthNotifier._fetchUserInfo - 用户信息为空');
        state = state.copyWith(
          isLoading: false,
          errorMessage: '用户信息为空',
        );
        return false;
      }

      // 优先使用 user/me 返回的 userId，fallback 到 JWT 中的 Id
      final effectiveUserId = info.userId > 0 ? info.userId : jwtUserId;

      debugPrint(
          '[DEBUG] AuthNotifier._fetchUserInfo - 更新状态: isLoggedIn=true, username=${info.username}, email=${info.email}, userId=$effectiveUserId');
      state = AuthState(
        isLoggedIn: true,
        token: token,
        userId: effectiveUserId,
        userName: info.username,
        email: info.email,
        userPosition: info.job,
        registeredAt: info.registeredAt,
        privilege: info.privilege,
        isLoading: false,
      );
      return true;
    } catch (e, stackTrace) {
      debugPrint('[ERROR] AuthNotifier._fetchUserInfo - 异常: $e');
      debugPrint('[ERROR] 堆栈: $stackTrace');
      state = state.copyWith(
        isLoading: false,
        errorMessage: '连接失败',
      );
      return false;
    }
  }
}

/// 认证 Provider
final authProvider = NotifierProvider<AuthNotifier, AuthState>(
  AuthNotifier.new,
);
