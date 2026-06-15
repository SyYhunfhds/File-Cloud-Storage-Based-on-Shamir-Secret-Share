import 'dart:convert';
import 'package:http/http.dart' as http;
import '../../auth/models/auth_models.dart';
import '../models/server_share_models.dart';

/// 服务器份额 API 服务
///
/// 调用后端 share/list 和 share/pull 接口。
/// 复用 [parseApiResponse] 模式与现有 API 服务保持一致。
class ShareApiService {
  final String baseUrl;

  ShareApiService(this.baseUrl);

  /// 获取远端可拉取份额列表 — `POST v1/protected/share/list`
  Future<ApiResponse<ShareListRes>> list({
    required String token,
    int page = 1,
    int size = 20,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/share/list');
    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'page': page, 'size': size}),
    );
    return parseApiResponse<ShareListRes>(
      response.body,
      (json) => ShareListRes.fromJson(json),
    );
  }

  /// 拉取份额（取后即焚） — `POST v1/protected/share/pull`
  Future<ApiResponse<SharePullRes>> pull({
    required String token,
    required List<int> shareIds,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/share/pull');
    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'share_ids': shareIds}),
    );
    return parseApiResponse<SharePullRes>(
      response.body,
      (json) => SharePullRes.fromJson(json),
    );
  }
}
