import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'pages/about_page.dart';
import 'pages/audit_page.dart';
import 'pages/home_page.dart';
import 'pages/login_page.dart';
import 'pages/register_page.dart';
import 'pages/share_page.dart';
import 'providers/auth_provider.dart';
import 'utils/constants.dart';

/// 受保护的路由列表
final _protectedRoutes = [
  Constants.routeHome,
  Constants.routeShares,
  Constants.routeAudit,
];

/// 判断路由是否需要登录
bool _requiresAuth(String path) {
  return _protectedRoutes.contains(path);
}

/// 路由配置
final router = Provider((ref) {
  return GoRouter(
    initialLocation: Constants.routeLogin,
    redirect: (context, state) {
      final authState = ref.read(authProvider);
      final isAuthenticated = authState.isAuthenticated;
      final currentPath = state.uri.path;

      if (_requiresAuth(currentPath) && !isAuthenticated) {
        return Constants.routeLogin;
      }

      return null;
    },
    routes: [
      GoRoute(
        path: Constants.routeLogin,
        builder: (context, state) => const LoginPage(),
      ),
      GoRoute(
        path: Constants.routeRegister,
        builder: (context, state) => const RegisterPage(),
      ),
      GoRoute(
        path: Constants.routeHome,
        builder: (context, state) => const HomePage(),
      ),
      GoRoute(
        path: Constants.routeShares,
        builder: (context, state) => const SharePage(),
      ),
      GoRoute(
        path: Constants.routeAudit,
        builder: (context, state) => const AuditPage(),
      ),
      GoRoute(
        path: Constants.routeAbout,
        builder: (context, state) => const AboutPage(),
      ),
    ],
  );
});