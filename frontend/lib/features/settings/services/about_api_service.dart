import 'package:http/http.dart' as http;
import '../../auth/models/auth_models.dart';
import '../models/about_model.dart';

/// `/v1/about` API 服务
///
/// 无需认证（不在 protected 路由组下）。
class AboutApiService {
  final String baseUrl;
  const AboutApiService(this.baseUrl);

  /// 获取项目版本和开发团队信息
  Future<ApiResponse<AboutInfo>> getAbout() async {
    final url = Uri.parse('$baseUrl/v1/about');
    final response = await http.get(url);
    return parseApiResponse<AboutInfo>(
      response.body,
      (data) => AboutInfo.fromJson(data as Map<String, dynamic>),
    );
  }
}
