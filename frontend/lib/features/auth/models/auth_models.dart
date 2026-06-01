import 'dart:convert';

/// 后端统一响应结构
///
/// 所有 API 返回的 JSON 格式均为：
/// ```json
/// { "code": 0, "message": "OK", "data": { ... } }
/// ```
class ApiResponse<T> {
  final int code;
  final String message;
  final T? data;

  const ApiResponse({required this.code, required this.message, this.data});

  bool get isSuccess => code == 0;

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(dynamic)? fromDataJson,
  ) {
    return ApiResponse(
      code: json['code'] as int? ?? 0,
      message: json['message'] as String? ?? '',
      data: json['data'] != null && fromDataJson != null
          ? fromDataJson(json['data'])
          : null,
    );
  }
}

/// 为简单类型（如 String）提供转换辅助

/// 将 JSON 字符串解析为 [ApiResponse]
ApiResponse<T> parseApiResponse<T>(
  String body,
  T Function(dynamic)? fromDataJson,
) {
  try {
    print('[DEBUG] parseApiResponse - 原始响应体: $body');
    final Map<String, dynamic> json = jsonDecode(body) as Map<String, dynamic>;
    print('[DEBUG] parseApiResponse - 解析后的JSON: $json');
    final resp = ApiResponse.fromJson(json, fromDataJson);
    print('[DEBUG] parseApiResponse - code: ${resp.code}, message: ${resp.message}, isSuccess: ${resp.isSuccess}');
    return resp;
  } catch (e, stackTrace) {
    print('[ERROR] parseApiResponse 解析失败: $e');
    print('[ERROR] 堆栈: $stackTrace');
    rethrow;
  }
}
