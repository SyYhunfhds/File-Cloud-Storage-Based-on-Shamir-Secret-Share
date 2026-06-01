import 'dart:async';
import 'dart:math';

import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/about.dart';
import '../models/item.dart';
import '../models/response.dart';
import '../models/share.dart';
import '../models/user.dart';
import '../providers/auth_provider.dart';
import '../utils/constants.dart';

typedef OnTokenExpired = void Function();

/// API服务类
class ApiService {
  late Dio _dio;
  late SharedPreferences _prefs;
  String? _cachedToken;
  OnTokenExpired? _onTokenExpired;

  void setOnTokenExpired(OnTokenExpired callback) {
    _onTokenExpired = callback;
  }

  Future<void> init() async {
    _dio = Dio(BaseOptions(
      baseUrl: Constants.baseUrl,
      connectTimeout: Duration(seconds: Constants.timeoutSeconds),
      receiveTimeout: Duration(seconds: Constants.timeoutSeconds),
    ));

    _prefs = await SharedPreferences.getInstance();

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        _cachedToken ??= _prefs.getString(Constants.storageTokenKey);

        if (_cachedToken != null) {
          options.headers['Cookie'] = 'authorization=$_cachedToken';
        }

        print('Request: ${options.method} ${options.path}');
        print('Headers: ${options.headers}');
        if (options.data != null) {
          print('Data: ${options.data}');
        }
        handler.next(options);
      },
      onResponse: (response, handler) {
        print('Response: ${response.statusCode} ${response.statusMessage}');
        handler.next(response);
      },
      onError: (error, handler) {
        print('Error: ${error.response?.statusCode} ${error.message}');
        if (error.response?.statusCode == 401) {
          _handleTokenExpired();
        }
        handler.next(error);
      },
    ));
  }

  void _handleTokenExpired() async {
    await _prefs.remove(Constants.storageTokenKey);
    _cachedToken = null;
    _onTokenExpired?.call();
    AuthNotifier.onTokenExpired?.call();
  }

  void _clearCachedToken() {
    _cachedToken = null;
  }

  Future<ApiResponse<String>> login({
    required String username,
    required String password,
  }) async {
    try {
      final response = await _dio.post(
        Constants.apiLogin,
        data: {
          'username': username,
          'password': password,
        },
      );

      final code = response.data['code'] as int? ?? -1;
      final message = response.data['message'] as String? ?? '';

      if (code == ErrorCode.success) {
        final token = response.data['data'] as String?;
        if (token != null) {
          await _prefs.setString(Constants.storageTokenKey, token);
          _cachedToken = token;
        }
      }

      return ApiResponse<String>(
        code: code,
        message: message,
        data: code == ErrorCode.success ? response.data['data'] as String? : null,
      );
    } catch (e) {
      return ApiResponse<String>(
        code: ErrorCode.internalError,
        message: '登录失败: ${e.toString()}',
      );
    }
  }

  Future<ApiResponse<Map<String, dynamic>>> register({
    required String username,
    required String password,
    required String email,
  }) async {
    try {
      final response = await _dio.post(
        Constants.apiRegister,
        data: {
          'username': username,
          'password': password,
          'email': email,
        },
      );

      return ApiResponse<Map<String, dynamic>>(
        code: response.data['code'] ?? -1,
        message: response.data['message'] ?? '',
        data: response.data['data'] != null 
            ? Map<String, dynamic>.from(response.data['data']) 
            : null,
      );
    } catch (e) {
      return ApiResponse<Map<String, dynamic>>(
        code: ErrorCode.internalError,
        message: '注册失败: ${e.toString()}',
      );
    }
  }

  Future<ApiResponse<void>> logout() async {
    try {
      final response = await _dio.get(Constants.apiLogout);

      await _prefs.remove(Constants.storageTokenKey);
      _clearCachedToken();

      return ApiResponse<void>(
        code: response.data['code'] ?? 0,
        message: response.data['message'] ?? '',
      );
    } catch (e) {
      await _prefs.remove(Constants.storageTokenKey);
      _clearCachedToken();
      return ApiResponse<void>(
        code: ErrorCode.internalError,
        message: '登出失败: ${e.toString()}',
      );
    }
  }

  Future<ApiResponse<User>> getUserInfo() async {
    try {
      final response = await _dio.get(Constants.apiUserMe);

      if (response.data['code'] == ErrorCode.success) {
        return ApiResponse<User>(
          code: 0,
          message: 'OK',
          data: User.fromJson(response.data['data']),
        );
      }

      return ApiResponse<User>(
        code: response.data['code'] ?? -1,
        message: response.data['message'] ?? '',
      );
    } catch (e) {
      return ApiResponse<User>(
        code: ErrorCode.internalError,
        message: '获取用户信息失败: ${e.toString()}',
      );
    }
  }

  Future<ApiResponse<ItemSubmitResponse>> submitItem(
    List<int> fileBytes,
    String filename,
  ) async {
    try {
      final formData = FormData.fromMap({
        'item': MultipartFile.fromBytes(
          fileBytes,
          filename: filename,
        ),
      });

      final response = await _dio.post(
        Constants.apiItemSubmit,
        data: formData,
      );

      if (response.data['code'] == ErrorCode.success) {
        return ApiResponse<ItemSubmitResponse>(
          code: 0,
          message: response.data['message'] ?? '上传成功',
          data: ItemSubmitResponse.fromJson(response.data['data']),
        );
      }

      return ApiResponse<ItemSubmitResponse>(
        code: response.data['code'] ?? -1,
        message: response.data['cause'] ?? response.data['message'] ?? '',
      );
    } catch (e) {
      return ApiResponse<ItemSubmitResponse>(
        code: ErrorCode.internalError,
        message: '上传失败: ${e.toString()}',
      );
    }
  }

  Future<ApiResponse<ItemListResponse>> getItemList({
    int page = 1,
    int size = 10,
  }) async {
    try {
      final response = await _dio.get(
        Constants.apiItemList,
        queryParameters: {
          'page': page,
          'size': size,
        },
      );

      if (response.data['code'] == ErrorCode.success) {
        return ApiResponse<ItemListResponse>(
          code: 0,
          message: 'OK',
          data: ItemListResponse.fromJson(response.data['data']),
        );
      }

      return ApiResponse<ItemListResponse>(
        code: response.data['code'] ?? -1,
        message: response.data['message'] ?? '',
      );
    } catch (e) {
      return ApiResponse<ItemListResponse>(
        code: ErrorCode.internalError,
        message: '获取条目列表失败: ${e.toString()}',
      );
    }
  }

  Future<ApiResponse<void>> updateItem({
    required String filename,
    String? newFilename,
    int? minimumPrivilege,
  }) async {
    try {
      final data = <String, dynamic>{
        'filename': filename,
      };

      if (newFilename != null) {
        data['new_filename'] = newFilename;
      }
      if (minimumPrivilege != null) {
        data['minimum_privilege'] = minimumPrivilege;
      }

      final response = await _dio.post(
        Constants.apiItemUpdate,
        data: data,
      );

      return ApiResponse<void>(
        code: response.data['code'] ?? 0,
        message: response.data['message'] ?? '',
      );
    } catch (e) {
      return ApiResponse<void>(
        code: ErrorCode.internalError,
        message: '更新条目失败: ${e.toString()}',
      );
    }
  }

  String generateShareValue() {
    final random = Random.secure();
    final bytes = List<int>.generate(32, (i) => random.nextInt(256));
    return _base64Encode(bytes);
  }

  String _base64Encode(List<int> bytes) {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
    final buffer = StringBuffer();
    for (var i = 0; i < bytes.length; i += 3) {
      final b1 = bytes[i];
      final b2 = i + 1 < bytes.length ? bytes[i + 1] : 0;
      final b3 = i + 2 < bytes.length ? bytes[i + 2] : 0;
      buffer.write(chars[(b1 >> 2) & 0x3F]);
      buffer.write(chars[((b1 << 4) | (b2 >> 4)) & 0x3F]);
      buffer.write(i + 1 < bytes.length ? chars[((b2 << 2) | (b3 >> 6)) & 0x3F] : '=');
      buffer.write(i + 2 < bytes.length ? chars[b3 & 0x3F] : '=');
    }
    return buffer.toString();
  }

  Future<ApiResponse<List<Share>>> getShares() async {
    await Future.delayed(const Duration(milliseconds: 500));

    final shares = [
      Share(
        id: 1,
        userId: 1,
        value: generateShareValue(),
        version: 1,
        createdAt: DateTime.now().subtract(const Duration(days: 1)),
      ),
      Share(
        id: 2,
        userId: 2,
        value: generateShareValue(),
        version: 1,
        createdAt: DateTime.now().subtract(const Duration(days: 2)),
      ),
      Share(
        id: 3,
        userId: 3,
        value: generateShareValue(),
        version: 1,
        createdAt: DateTime.now().subtract(const Duration(days: 3)),
      ),
    ];

    return ApiResponse<List<Share>>(
      code: 0,
      message: 'OK',
      data: shares,
    );
  }

  Future<ApiResponse<About>> getAboutInfo() async {
    try {
      final response = await _dio.get(Constants.apiAbout);

      if (response.data['code'] == ErrorCode.success) {
        return ApiResponse<About>(
          code: 0,
          message: 'OK',
          data: About.fromJson(response.data['data']),
        );
      }

      return ApiResponse<About>(
        code: response.data['code'] ?? -1,
        message: response.data['message'] ?? '',
      );
    } catch (e) {
      return ApiResponse<About>(
        code: ErrorCode.internalError,
        message: '获取帮助信息失败: ${e.toString()}',
      );
    }
  }
}