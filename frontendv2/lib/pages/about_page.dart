import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/about.dart';
import '../providers/auth_provider.dart';
import '../providers/service_providers.dart';
import '../utils/constants.dart';

class AboutPage extends ConsumerStatefulWidget {
  const AboutPage({super.key});

  @override
  ConsumerState<AboutPage> createState() => _AboutPageState();
}

class _AboutPageState extends ConsumerState<AboutPage> {
  About? _aboutData;
  bool _isLoading = false;
  bool _isRefreshing = false;
  String? _errorMessage;

  static const String _storageKey = 'about_cache';
  static const String _storageTimeKey = 'about_cache_time';

  @override
  void initState() {
    super.initState();
    _loadFromCache();
  }

  Future<void> _loadFromCache() async {
    setState(() => _isLoading = true);
    try {
      final prefs = await SharedPreferences.getInstance();
      final cachedJson = prefs.getString(_storageKey);
      if (cachedJson != null) {
        setState(() {
          _aboutData = About.fromJson(json.decode(cachedJson));
        });
      }
      await _fetchFromApi(useCache: cachedJson != null);
    } catch (e) {
      await _fetchFromApi(useCache: false);
    }
  }

  Future<void> _fetchFromApi({bool useCache = true}) async {
    if (!useCache) {
      setState(() => _isRefreshing = true);
    }
    final apiService = ref.read(apiServiceProvider);
    final response = await apiService.getAboutInfo();

    if (mounted) {
      setState(() {
        _isLoading = false;
        _isRefreshing = false;
      });

      if (response.isSuccess && response.data != null) {
        setState(() {
          _aboutData = response.data;
          _errorMessage = null;
        });
        await _saveToCache(response.data!);
      } else {
        if (!useCache) {
          setState(() {
            _errorMessage = response.message;
          });
        }
      }
    }
  }

  Future<void> _saveToCache(About data) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_storageKey, json.encode(data.toJson()));
    await prefs.setInt(_storageTimeKey, DateTime.now().millisecondsSinceEpoch);
  }

  Future<void> _clearCache() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_storageKey);
    await prefs.remove(_storageTimeKey);
  }

  Future<void> _refresh() async {
    await _clearCache();
    await _fetchFromApi(useCache: false);
  }

  Future<void> _handleLogout() async {
    final apiService = ref.read(apiServiceProvider);
    await apiService.logout();
    await ref.read(authProvider.notifier).logout();
    if (mounted) {
      context.go(Constants.routeLogin);
    }
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('帮助信息'),
        leading: Builder(
          builder: (context) => IconButton(
            icon: const Icon(Icons.menu),
            onPressed: () => Scaffold.of(context).openDrawer(),
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.home),
            onPressed: () => context.go(Constants.routeHome),
            tooltip: '返回首页',
          ),
          IconButton(
            icon: _isRefreshing
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.refresh),
            onPressed: _isRefreshing ? null : _refresh,
            tooltip: '刷新',
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
              onTap: () {
                Navigator.pop(context);
                context.go(Constants.routeHome);
              },
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
              selected: true,
              onTap: () => Navigator.pop(context),
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
      body: RefreshIndicator(
        onRefresh: _refresh,
        child: _buildBody(),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading && _aboutData == null) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_errorMessage != null && _aboutData == null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.error_outline, size: 64, color: Colors.grey[400]),
              const SizedBox(height: 16),
              Text(
                _errorMessage!,
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.grey[600]),
              ),
              const SizedBox(height: 24),
              ElevatedButton.icon(
                onPressed: _refresh,
                icon: const Icon(Icons.refresh),
                label: const Text('重试'),
              ),
            ],
          ),
        ),
      );
    }

    return SingleChildScrollView(
      padding: const EdgeInsets.all(16),
      physics: const AlwaysScrollableScrollPhysics(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _buildAppHeader(),
          const SizedBox(height: 24),
          _buildVersionCard(),
          const SizedBox(height: 16),
          _buildLeaderCard(),
          const SizedBox(height: 16),
          _buildDevelopersCard(),
        ],
      ),
    );
  }

  Widget _buildAppHeader() {
    return Card(
      elevation: 4,
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          children: [
            CircleAvatar(
              radius: 48,
              backgroundColor: Theme.of(context).colorScheme.primaryContainer,
              child: Icon(
                Icons.account_balance_outlined,
                size: 48,
                color: Theme.of(context).colorScheme.primary,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              '企业财务审计系统',
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Text(
              'Enterprise Financial Audit System',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVersionCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Icon(Icons.tag, color: Colors.blue[600], size: 28),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '版本号',
                    style: TextStyle(
                      color: Colors.grey[600],
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _aboutData?.version ?? '--',
                    style: const TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLeaderCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Icon(Icons.person_outline, color: Colors.purple[600], size: 28),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '项目领导人',
                    style: TextStyle(
                      color: Colors.grey[600],
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _aboutData?.leader ?? '--',
                    style: const TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDevelopersCard() {
    final developers = _aboutData?.developers ?? [];
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.people_outline, color: Colors.green[600], size: 28),
                const SizedBox(width: 16),
                Text(
                  '开发团队',
                  style: TextStyle(
                    color: Colors.grey[600],
                    fontSize: 14,
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.green[100],
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    '${developers.length}人',
                    style: TextStyle(
                      color: Colors.green[700],
                      fontWeight: FontWeight.w600,
                      fontSize: 12,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            if (developers.isEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 24),
                child: Center(
                  child: Text(
                    '暂无开发者信息',
                    style: TextStyle(color: Colors.grey[500]),
                  ),
                ),
              )
            else
              ...developers.asMap().entries.map((entry) {
                final index = entry.key;
                final developer = entry.value;
                return Padding(
                  padding: EdgeInsets.only(
                    bottom: index < developers.length - 1 ? 12 : 0,
                  ),
                  child: Container(
                    decoration: BoxDecoration(
                      color: Colors.grey[100],
                      borderRadius: BorderRadius.circular(8),
                    ),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 12,
                    ),
                    child: Row(
                      children: [
                        Container(
                          width: 32,
                          height: 32,
                          decoration: BoxDecoration(
                            color: Colors.blue[100],
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Center(
                            child: Text(
                              '${index + 1}',
                              style: TextStyle(
                                color: Colors.blue[800],
                                fontWeight: FontWeight.bold,
                                fontSize: 16,
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        CircleAvatar(
                          radius: 20,
                          backgroundColor: Colors.primaries[
                              index % Colors.primaries.length][200],
                          child: Text(
                            developer.isNotEmpty ? developer[0].toUpperCase() : '?',
                            style: TextStyle(
                              color: Colors.primaries[
                                  index % Colors.primaries.length][800],
                              fontWeight: FontWeight.bold,
                              fontSize: 18,
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Text(
                            developer,
                            style: const TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w500,
                              color: Colors.black87,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                );
              }),
          ],
        ),
      ),
    );
  }
}
