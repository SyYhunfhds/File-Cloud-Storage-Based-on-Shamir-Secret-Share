import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/auth_models.dart';
import '../models/user_model.dart';

/// 认证相关 API 调用封装
///
/// baseUrl 从 [ApiConfig] 获取，支持运行时动态修改。
class AuthApiService {
  final String baseUrl;

  AuthApiService(this.baseUrl);

  // ===================== 业务方法 =====================

  /// 注册 — `POST /v1/user/`
  ///
  /// 注意：注册接口使用 RESTful 风格，路径为 `v1/user/`。
  Future<ApiResponse<void>> register({
    required String username,
    required String password,
    required String email,
  }) async {
    final url = Uri.parse('$baseUrl/v1/user/');
    final response = await http.post(
      url,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'username': username,
        'password': password,
        'email': email,
      }),
    );

    final apiResp = parseApiResponse<void>(response.body, null);
    return apiResp;
  }

  /// 登录 — `POST /v1/auth/login`
  ///
  /// 成功时 `data` 字段包含 JWT。
  Future<ApiResponse<String>> login({
    required String password,
    String? username,
    String? email,
  }) async {
    final url = Uri.parse('$baseUrl/v1/auth/login');
    final body = <String, dynamic>{
      'password': password,
    };
    if (username != null && username.isNotEmpty) {
      body['username'] = username;
    }
    if (email != null && email.isNotEmpty) {
      body['email'] = email;
    }

    final response = await http.post(
      url,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(body),
    );

    final apiResp = parseApiResponse<String>(
      response.body,
      (data) => data as String,
    );

    return apiResp;
  }

  /// 获取当前用户信息 — `GET /v1/protected/user/me`
  Future<ApiResponse<UserInfo>> getUserInfo(String token) async {
    final url = Uri.parse('$baseUrl/v1/protected/user/me');
    final response = await http.get(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
    );

    final apiResp = parseApiResponse<UserInfo>(
      response.body,
      (data) => UserInfo.fromJson(data as Map<String, dynamic>),
    );

    return apiResp;
  }

  /// 登出 — `GET /v1/protected/auth/logout`
  Future<ApiResponse<void>> logout(String token) async {
    final url = Uri.parse('$baseUrl/v1/protected/auth/logout');
    final response = await http.get(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
    );

    final apiResp = parseApiResponse<void>(response.body, null);
    return apiResp;
  }
}
