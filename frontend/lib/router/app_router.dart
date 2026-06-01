import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../features/home/views/home_page.dart';
import '../features/upload/views/upload_page.dart';
import '../features/share/views/share_page.dart';
import '../features/settings/views/settings_page.dart';
import '../features/home/widgets/sidebar_nav.dart';

/// GoRouter 路由配置 — StatefulShellRoute
///
/// 侧边栏位于全局 AppShell 层，在所有页面间保持持久可见。
final appRouter = GoRouter(
  initialLocation: '/',
  routes: [
    StatefulShellRoute.indexedStack(
      builder: (context, state, navigationShell) => AppShell(
        navigationShell: navigationShell,
      ),
      branches: [
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: HomePage(),
              ),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/upload',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: UploadPage(),
              ),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/share',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: SharePage(),
              ),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/settings',
              pageBuilder: (context, state) => const NoTransitionPage(
                child: SettingsPage(),
              ),
            ),
          ],
        ),
      ],
    ),
  ],
);

/// 应用外壳：全局 NavigationRail 侧边栏 + 内容区
///
/// 这是应用内唯一的 Scaffold。子页面不再包裹独立的 Scaffold，
/// 以避免嵌套脚手架造成的边距与行为差异。
class AppShell extends StatelessWidget {
  final StatefulNavigationShell navigationShell;

  const AppShell({
    super.key,
    required this.navigationShell,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Row(
        children: [
          AppSidebarNav(
            selectedIndex: navigationShell.currentIndex,
            onDestinationSelected: (index) {
              navigationShell.goBranch(index);
            },
          ),
          const VerticalDivider(width: 1, thickness: 1),
          Expanded(child: navigationShell),
        ],
      ),
    );
  }
}
