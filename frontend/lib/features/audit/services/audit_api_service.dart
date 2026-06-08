import 'dart:convert';

import 'package:http/http.dart' as http;

import '../../auth/models/auth_models.dart';
import '../models/audit_models.dart';

/// 审计链路 API 服务
///
/// 封装 `POST v1/protected/audit/list/` 和 `POST v1/protected/audit/operate/`。
class AuditApiService {
  final String baseUrl;
  const AuditApiService(this.baseUrl);

  /// 分页获取审计列表
  ///
  /// [token] JWT 令牌，[page] 页码(从1开始)，[size] 每页数量，[scope] 筛选范围。
  Future<ApiResponse<AuditListRes>> listAudits({
    required String token,
    int page = 1,
    int size = 20,
    required List<int> scope,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/audit/list/');
    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({
        'page': page,
        'size': size,
        'scope': scope,
      }),
    );
    return parseApiResponse<AuditListRes>(
      response.body,
      (data) => AuditListRes.fromJson(data as Map<String, dynamic>),
    );
  }

  /// 审计操作（通过/拒绝）
  ///
  /// 新 API 仅接收 operations map。
  /// key 为 audit_id（整数），value 为 "pass" 或 "reject"。
  Future<ApiResponse<AuditOperationRes>> operateAudit({
    required String token,
    required Map<int, String> operations,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/audit/operate/');
    // 手动构造 JSON body — Dart jsonEncode 不支持 Map<int,String> 的整数 key
    final opsEntries = operations.entries
        .map((e) => '"${e.key}":"${e.value}"')
        .join(',');
    final body = '{"operations":{$opsEntries}}';
    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: body,
    );
    return parseApiResponse<AuditOperationRes>(
      response.body,
      (data) => AuditOperationRes.fromJson(data as Map<String, dynamic>),
    );
  }
}