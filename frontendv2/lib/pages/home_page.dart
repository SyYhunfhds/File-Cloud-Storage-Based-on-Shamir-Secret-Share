import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../providers/auth_provider.dart';
import '../providers/service_providers.dart';
import '../utils/constants.dart';

class HomePage extends ConsumerStatefulWidget {
  const HomePage({super.key});

  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> {
  @override
  void initState() {
    super.initState();
    _loadUserInfo();
  }

  Future<void> _loadUserInfo() async {
    final apiService = ref.read(apiServiceProvider);
    final response = await apiService.getUserInfo();
    if (response.isSuccess && response.data != null) {
      await ref.read(authProvider.notifier).loginSuccess(response.data!);
    }
  }

  Future<void> _handleLogout() async {
    final apiService = ref.read(apiServiceProvider);
    final storageService = ref.read(storageServiceProvider);
    await apiService.logout();
    await storageService.clearAll();
    await ref.read(authProvider.notifier).logout();
    if (mounted) {
      context.go(Constants.routeLogin);
    }
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authProvider);
    final user = authState.user;

    return Scaffold(
      appBar: AppBar(
        title: const Text('企业财务审计系统'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: _handleLogout,
            tooltip: '退出登录',
          ),
        ],
      ),
      drawer: Drawer(
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            UserAccountsDrawerHeader(
              accountName: Text(user?.username ?? '用户'),
              accountEmail: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(user?.email ?? ''),
                  if (user?.job != null)
                    Text(
                      '${user!.job} | ${user.privilegeText}',
                      style: const TextStyle(fontSize: 12),
                    ),
                ],
              ),
              currentAccountPicture: CircleAvatar(
                child: Text(
                  user?.username.isNotEmpty == true
                      ? user!.username[0].toUpperCase()
                      : 'U',
                ),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.home),
              title: const Text('首页'),
              selected: true,
              onTap: () => Navigator.pop(context),
            ),
            ListTile(
              leading: const Icon(Icons.key),
              title: const Text('份额管理'),
              onTap: () {
                Navigator.pop(context);
                context.go(Constants.routeShares);
              },
            ),
            ListTile(
              leading: const Icon(Icons.description),
              title: const Text('审计条目'),
              onTap: () {
                Navigator.pop(context);
                context.go(Constants.routeAudit);
              },
            ),
            ListTile(
              leading: const Icon(Icons.info_outline),
              title: const Text('帮助信息'),
              onTap: () {
                Navigator.pop(context);
                context.go(Constants.routeAbout);
              },
            ),
            const Divider(),
            ListTile(
              leading: const Icon(Icons.logout),
              title: const Text('退出登录'),
              onTap: _handleLogout,
            ),
          ],
        ),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        CircleAvatar(
                          radius: 30,
                          child: Text(
                            user?.username.isNotEmpty == true
                                ? user!.username[0].toUpperCase()
                                : 'U',
                            style: const TextStyle(fontSize: 24),
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                user?.username ?? '未知用户',
                                style: const TextStyle(
                                  fontSize: 20,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              if (user?.job != null)
                                Text(
                                  user!.job!,
                                  style: const TextStyle(color: Colors.grey),
                                ),
                            ],
                          ),
                        ),
                        if (user?.privilege != null)
                          Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 6,
                            ),
                            decoration: BoxDecoration(
                              color: _getPrivilegeColor(user!.privilege!),
                              borderRadius: BorderRadius.circular(20),
                            ),
                            child: Text(
                              user.privilegeText,
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 12,
                              ),
                            ),
                          ),
                      ],
                    ),
                    const Divider(height: 32),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceAround,
                      children: [
                        _buildInfoItem(
                          '邮箱',
                          user?.email ?? '未设置',
                          Icons.email,
                        ),
                        if (user?.registeredAt != null)
                          _buildInfoItem(
                            '注册时间',
                            _formatDate(user!.registeredAt!),
                            Icons.calendar_today,
                          ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              '功能菜单',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              child: GridView.count(
                crossAxisCount: 2,
                crossAxisSpacing: 16,
                mainAxisSpacing: 16,
                children: [
                  _buildMenuCard(
                    icon: Icons.key,
                    title: '份额管理',
                    subtitle: '管理和更新密钥份额',
                    color: Colors.blue,
                    onTap: () => context.go(Constants.routeShares),
                  ),
                  _buildMenuCard(
                    icon: Icons.description,
                    title: '审计条目',
                    subtitle: '查看和管理审计条目',
                    color: Colors.green,
                    onTap: () => context.go(Constants.routeAudit),
                  ),
                  _buildMenuCard(
                    icon: Icons.shield,
                    title: '系统状态',
                    subtitle: '运行正常',
                    color: Colors.orange,
                    onTap: () {},
                  ),
                  _buildMenuCard(
                    icon: Icons.info,
                    title: '帮助信息',
                    subtitle: '查看系统信息',
                    color: Colors.purple,
                    onTap: () => context.go(Constants.routeAbout),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoItem(String label, String value, IconData icon) {
    return Column(
      children: [
        Icon(icon, color: Colors.grey),
        const SizedBox(height: 4),
        Text(
          label,
          style: const TextStyle(color: Colors.grey, fontSize: 12),
        ),
        const SizedBox(height: 2),
        Text(
          value,
          style: const TextStyle(fontWeight: FontWeight.w500),
        ),
      ],
    );
  }

  Widget _buildMenuCard({
    required IconData icon,
    required String title,
    required String subtitle,
    required Color color,
    required VoidCallback onTap,
  }) {
    return Card(
      elevation: 4,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon, size: 48, color: color),
              const SizedBox(height: 16),
              Text(
                title,
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                subtitle,
                style: const TextStyle(color: Colors.grey),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }

  Color _getPrivilegeColor(int privilege) {
    switch (privilege) {
      case 1:
        return Colors.green;
      case 2:
        return Colors.blue;
      case 3:
        return Colors.orange;
      case 4:
        return Colors.purple;
      case 5:
        return Colors.red;
      default:
        return Colors.grey;
    }
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }
}