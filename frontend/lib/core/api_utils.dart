import 'dart:async';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';

import '../features/auth/models/auth_models.dart';

/// API 异常，不泄漏路径
class ApiException implements Exception {
  final String userMessage;
  final String? debugInfo;

  const ApiException({required this.userMessage, this.debugInfo});

  @override
  String toString() => userMessage;
}

/// 安全执行 API 调用，统一处理连接异常
///
/// 连接失败 → 弹窗"无法连接到服务器"（不显示具体路径）
/// 业务失败 → 调用方自行处理 [ApiResponse.isSuccess]
Future<T?> safeApiCall<T>(
  BuildContext? context,
  Future<T> Function() apiCall, {
  bool showErrorDialog = true,
}) async {
  try {
    return await apiCall();
  } on SocketException catch (e) {
    debugPrint('[API] 连接失败: $e');
    if (showErrorDialog && context != null && context.mounted) {
      _showConnectionFailedDialog(context);
    }
  } on TimeoutException catch (e) {
    debugPrint('[API] 请求超时: $e');
    if (showErrorDialog && context != null && context.mounted) {
      _showConnectionFailedDialog(context, message: '请求超时，请检查网络或后端服务。');
    }
  } catch (e) {
    debugPrint('[API] 未知异常: $e');
    if (showErrorDialog && context != null && context.mounted) {
      _showConnectionFailedDialog(context, message: '发生未知错误，请稍后重试。');
    }
  }
  return null;
}

/// 检查业务是否成功，失败时弹窗显示 [message] 字段内容
bool checkApiSuccess<T>(
  BuildContext context,
  ApiResponse<T> resp, {
  bool showErrorDialog = true,
}) {
  if (resp.isSuccess) return true;
  if (showErrorDialog && context.mounted) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('请求失败'),
        content: Text(resp.message.isNotEmpty ? resp.message : '操作失败'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('确定'),
          ),
        ],
      ),
    );
  }
  return false;
}

void _showConnectionFailedDialog(BuildContext context,
    {String message = '无法连接到服务器，请检查网络连接或后端服务是否正常。'}) {
  showDialog(
    context: context,
    builder: (ctx) => AlertDialog(
      title: const Text('连接失败'),
      content: Text(message),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(ctx).pop(),
          child: const Text('确定'),
        ),
      ],
    ),
  );
}
