import 'package:flutter_riverpod/flutter_riverpod.dart';

/// API 连接配置
class ApiConfig {
  final String protocol;
  final String host;
  final int port;

  const ApiConfig({
    this.protocol = 'http',
    this.host = 'localhost',
    this.port = 8000,
  });

  String get baseUrl => '$protocol://$host:$port';
}

/// API 配置 Provider
///
/// 开发者可以在"开发者模式"设置页面中动态修改此值。
final apiConfigProvider = NotifierProvider<ApiConfigNotifier, ApiConfig>(
  ApiConfigNotifier.new,
);

class ApiConfigNotifier extends Notifier<ApiConfig> {
  @override
  ApiConfig build() => const ApiConfig();

  void update({
    required String protocol,
    required String host,
    required int port,
  }) {
    state = ApiConfig(protocol: protocol, host: host, port: port);
  }
}
