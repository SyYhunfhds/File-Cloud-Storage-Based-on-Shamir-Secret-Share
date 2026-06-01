/// 常量定义
class Constants {
  // API配置
  static const String baseUrl = 'http://localhost:8000/';
  static const int timeoutSeconds = 30;

  // 存储键名
  static const String storageTokenKey = 'auth_token';
  static const String storageUserIdKey = 'user_id';
  static const String storageUsernameKey = 'username';

  // 路由路径
  static const String routeLogin = '/login';
  static const String routeRegister = '/register';
  static const String routeHome = '/';
  static const String routeShares = '/shares';
  static const String routeAudit = '/audit';
  static const String routeAbout = '/about';

  // API路径 (公开) - 注意：不能以/开头，否则Dio会跳过baseUrl
  static const String apiLogin = 'v1/auth/login';
  static const String apiRegister = 'v1/user'; // POST
  static const String apiAbout = 'v1/about';

  // API路径 (受保护 - 需要登录) - 注意：不能以/开头
  static const String apiLogout = 'v1/protected/auth/logout';
  static const String apiUserMe = 'v1/protected/user/me';
  static const String apiItemSubmit = 'v1/protected/item/submit';
  static const String apiItemList = 'v1/protected/item/list';
  static const String apiItemUpdate = 'v1/protected/item/update';
}

/// 错误码定义
class ErrorCode {
  static const int success = 0;
  static const int internalError = 61;
  static const int invalidRequest = 53;
  static const int invalidParameter = 51;
  static const int notAuthorized = 401;
  static const int tokenExpired = 66;
  static const int tokenInvalid = 61;
  static const int tokenFormatError = 55;
}
