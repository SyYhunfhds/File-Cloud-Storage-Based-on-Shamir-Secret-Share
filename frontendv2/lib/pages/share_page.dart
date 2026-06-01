import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../models/share.dart';
import '../providers/auth_provider.dart';
import '../providers/share_provider.dart';
import '../providers/service_providers.dart';
import '../utils/constants.dart';

class SharePage extends ConsumerStatefulWidget {
  const SharePage({super.key});

  @override
  ConsumerState<SharePage> createState() => _SharePageState();
}

class _SharePageState extends ConsumerState<SharePage> {
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadShares();
  }

  Future<void> _loadShares() async {
    setState(() => _isLoading = true);

    final apiService = ref.read(apiServiceProvider);
    final response = await apiService.getShares();

    if (response.isSuccess && response.data != null) {
      final shares = (response.data as List)
          .map((item) => Share.fromJson(item))
          .toList();
      ref.read(shareProvider.notifier).setShares(shares);
    }

    setState(() => _isLoading = false);
  }

  @override
  Widget build(BuildContext context) {
    final shareState = ref.watch(shareProvider);
    final authState = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('份额管理'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              final apiService = ref.read(apiServiceProvider);
              await apiService.logout();
              await ref.read(authProvider.notifier).logout();
              context.go(Constants.routeLogin);
            },
          ),
        ],
      ),
      drawer: Drawer(
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            UserAccountsDrawerHeader(
              accountName: Text(authState.user?.username ?? '用户'),
              accountEmail: Text(authState.user?.email ?? ''),
              currentAccountPicture: const CircleAvatar(
                child: Icon(Icons.person),
              ),
            ),
            ListTile(
              leading: const Icon(Icons.home),
              title: const Text('首页'),
              onTap: () => context.go(Constants.routeHome),
            ),
            ListTile(
              leading: const Icon(Icons.key),
              title: const Text('份额管理'),
              selected: true,
              onTap: () => Navigator.pop(context),
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
              onTap: () async {
                final apiService = ref.read(apiServiceProvider);
                await apiService.logout();
                await ref.read(authProvider.notifier).logout();
                context.go(Constants.routeLogin);
              },
            ),
          ],
        ),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('当前份额版本:'),
                    Text(
                      'v${shareState.currentVersion}',
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 18,
                      ),
                    ),
                    ElevatedButton(
                      onPressed: _isLoading ? null : () {
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text('份额更新中...')),
                        );
                        Future.delayed(const Duration(seconds: 2), () {
                          ref.read(shareProvider.notifier).updateVersion(
                                shareState.currentVersion + 1,
                              );
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(content: Text('份额更新成功')),
                          );
                        });
                      },
                      child: const Text('更新份额'),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : ListView.builder(
                      itemCount: shareState.shares.length,
                      itemBuilder: (context, index) {
                        final share = shareState.shares[index];
                        return Card(
                          elevation: 2,
                          margin: const EdgeInsets.symmetric(vertical: 8),
                          child: Padding(
                            padding: const EdgeInsets.all(16),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  mainAxisAlignment:
                                      MainAxisAlignment.spaceBetween,
                                  children: [
                                    Text(
                                      '份额 #${share.id}',
                                      style: const TextStyle(
                                        fontWeight: FontWeight.bold,
                                        fontSize: 16,
                                      ),
                                    ),
                                    Text(
                                      '版本 ${share.version}',
                                      style: const TextStyle(color: Colors.grey),
                                    ),
                                  ],
                                ),
                                const SizedBox(height: 8),
                                Text('用户ID: ${share.userId}'),
                                const SizedBox(height: 8),
                                Text(
                                  '份额值: ${share.value.substring(0, 32)}...',
                                  style: const TextStyle(fontFamily: 'monospace'),
                                ),
                                const SizedBox(height: 8),
                                Text(
                                  '创建时间: ${share.createdAt.toString()}',
                                  style: const TextStyle(color: Colors.grey),
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
            ),
          ],
        ),
      ),
    );
  }
}