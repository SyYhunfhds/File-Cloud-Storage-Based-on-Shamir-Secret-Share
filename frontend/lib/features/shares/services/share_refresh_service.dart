import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/foundation.dart';

import '../models/share_refresh_models.dart';

/// 份额刷新 SSE 服务
///
/// 使用 `dart:io` HttpClient 发起 POST 请求并逐行读取 SSE 流。
/// 通过 [onProgress] 回调推送每条进度消息。
/// 使用 [Completer] 确保 [refresh] 方法在 SSE 流完全结束前不返回。
class ShareRefreshService {
  final String baseUrl;
  HttpClient? _client;
  StreamSubscription<String>? _subscription;

  ShareRefreshService(this.baseUrl);

  /// 发起 SSE 份额刷新请求
  ///
  /// [token] JWT，[itemId] 条目ID，[recoveryCode]/[deviceShare] 至少提供一个。
  /// [onProgress] 每收到一条 SSE 消息时回调。
  Future<void> refresh({
    required String token,
    required int itemId,
    String? recoveryCode,
    String? deviceShare,
    required Future<void> Function(ShareRefreshProgressMessage msg) onProgress,
    required void Function(String error) onError,
  }) async {
    _client?.close();
    _client = HttpClient();
    _client!.connectionTimeout = const Duration(seconds: 10);
    final completer = Completer<void>();
    var lastProgress = 0;

    try {
      final uri = Uri.parse('$baseUrl/v1/protected/share/refresh/');
      final request = await _client!.postUrl(uri);
      request.headers.set('Authorization', 'Bearer $token');
      request.headers.set('Content-Type', 'application/json');
      request.headers.set('Accept', 'text/event-stream');

      final req = ShareRefreshReq(
        itemId: itemId,
        recoveryCode: recoveryCode,
        deviceShare: deviceShare,
      );
      request.write(jsonEncode(req.toJson()));

      final response = await request.close();

      if (response.statusCode != 200) {
        final body = await response.transform(utf8.decoder).join();
        onError('服务器返回 ${response.statusCode}: $body');
        return;
      }

      final lineStream = response
          .transform(utf8.decoder)
          .transform(const LineSplitter());

      _subscription = lineStream.listen(
        (line) async {
          final trimmed = line.trim();
          if (trimmed.isEmpty) return;
          if (trimmed.startsWith(':')) return; // SSE 心跳/注释行

          try {
            final json = jsonDecode(trimmed) as Map<String, dynamic>;
            final msg = ShareRefreshProgressMessage.fromJson(json);
            lastProgress = msg.progress;
            debugPrint('[SSE] 收到消息: progress=${msg.progress}, message=${msg.message}');
            await onProgress(msg);

            // progress=100 或有 data → 完成，关闭流
            if (msg.progress >= 100 || msg.data != null) {
              debugPrint('[SSE] 刷新完成: progress=${msg.progress}, hasData=${msg.data != null}');
              if (!completer.isCompleted) completer.complete();
              _subscription?.cancel();
              // 不再调用 _client?.close()，避免强制关闭 TCP 连接触发 onError
            }
          } catch (e) {
            debugPrint('[SSE] JSON 解析失败: $e, 原始行: ${trimmed.length > 100 ? '${trimmed.substring(0, 100)}...' : trimmed}');
          }
        },
        onError: (e) {
          if (!completer.isCompleted) {
            debugPrint('[SSE] 流错误（未完成状态）: $e');
            onError('SSE 流异常: $e');
            completer.complete();
          } else {
            debugPrint('[SSE] 流错误（已完成状态，忽略）: $e');
          }
        },
        onDone: () {
          if (!completer.isCompleted && lastProgress < 100) {
            debugPrint('[SSE] 流意外终止（最后进度=$lastProgress）');
            onError('份额刷新异常中断，请联系管理员');
            completer.complete();
          } else {
            debugPrint('[SSE] 流正常关闭（lastProgress=$lastProgress）');
          }
        },
        cancelOnError: false,
      );

      return completer.future;
    } catch (e) {
      if (e is SocketException && e.osError?.errorCode == 995) {
        // 客户端主动取消（HttpClient.close()）→ 静默
        return;
      }
      onError('连接失败: $e');
    }
  }

  /// 取消正在进行的 SSE 流
  void cancel() {
    _subscription?.cancel();
    _subscription = null;
    _client?.close();
    _client = null;
  }
}
