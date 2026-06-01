import 'dart:convert';
import 'dart:typed_data';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../../auth/models/auth_models.dart';
import '../models/item_models.dart';
import '../../upload/models/upload_models.dart';

/// 条目份额链路 API 调用封装
///
/// 所有接口路径前缀: v1/protected/item/
class ItemApiService {
  final String baseUrl;

  ItemApiService(this.baseUrl);

  // ===================== 条目上传 =====================

  /// 条目上传 — `POST v1/protected/item/submit` (multipart/form-data)
  ///
  /// 文件表单字段名必须为 `item`，后端仅接受该键名。
  Future<ApiResponse<ItemSubmitRes>> submitItem({
    required String filePath,
    required String token,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/item/submit');

    final request = http.MultipartRequest('POST', url);
    request.headers['Authorization'] = 'Bearer $token';
    request.files.add(await http.MultipartFile.fromPath('item', filePath));

    final streamedResp = await request.send();
    final response = await http.Response.fromStream(streamedResp);

    final apiResp = parseApiResponse<ItemSubmitRes>(
      response.body,
      (data) => ItemSubmitRes.fromJson(data as Map<String, dynamic>),
    );

    return apiResp;
  }

  // ===================== 条目列表查询 =====================

  /// 条目分页列表 — `POST v1/protected/item/list`
  ///
  /// [scope] 1=我的条目, 2=公开条目, 3=所有可见
  Future<ApiResponse<ItemListRes>> listItems({
    required String token,
    int page = 1,
    int size = 10,
    int scope = 1,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/item/list');

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

    final apiResp = parseApiResponse<ItemListRes>(
      response.body,
      (data) => ItemListRes.fromJson(data as Map<String, dynamic>),
    );

    return apiResp;
  }

  // ===================== 条目举报/申请下载权限 =====================

  /// 申请下载权限 — `POST v1/protected/report`
  Future<ApiResponse<void>> reportItem({
    required int itemId,
    required String token,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/report');

    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({'item_id': itemId}),
    );

    return parseApiResponse<void>(response.body, null);
  }

  // ===================== 条目下载 =====================

  /// 条目下载 — `POST v1/protected/item/download`
  ///
  /// 成功时返回文件字节流，失败时返回 JSON 错误响应。
  /// 调用方通过检查返回值区分:
  /// - `data != null` → 文件数据
  /// - `errorMessage != null` → 业务错误
  Future<({Uint8List? data, String? errorMessage})> downloadItem({
    required int itemId,
    required String share,
    required String token,
  }) async {
    final url = Uri.parse('$baseUrl/v1/protected/item/download');

    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      body: jsonEncode({
        'item_id': itemId,
        'share': share,
      }),
    );

    // 检查 Content-Type 区分 JSON 错误和文件流
    final contentType = response.headers['content-type'] ?? '';
    if (contentType.contains('application/json')) {
      // JSON 错误响应
      final apiResp = parseApiResponse<void>(response.body, null);
      return (
        data: null,
        errorMessage: apiResp.isSuccess ? null : (apiResp.message.isNotEmpty ? apiResp.message : '下载失败'),
      );
    }

    // 非 JSON → 文件字节流
    return (data: response.bodyBytes, errorMessage: null);
  }
}
