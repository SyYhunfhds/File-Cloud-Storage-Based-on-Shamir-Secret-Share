import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../models/share.dart';
import '../providers/auth_provider.dart';
import '../services/api_service.dart';
import '../services/storage_service.dart';
import '../utils/constants.dart';

class SharePage extends ConsumerStatefulWidget {
  const SharePage({super.key});

  @override
  ConsumerState<SharePage> createState() => _SharePageState();
}

class _SharePageState extends ConsumerState<SharePage> {
  final StorageService _storageService = StorageService();
  bool _isLoading = false;
  List<LocalShare> _localShares = [];

  @override
  void initState() {
    super.initState();
    _loadShares();
  }

  Future<void> _loadShares() async {
    setState(() => _isLoading = true);

    // 从本地存储读取，不再调用后端API
    await _storageService.init();
    final shares = await _storageService.getShares();
    setState(() {
      _localShares = shares;
    });

    setState(() => _isLoading = false);
  }

  void _showShareDetail(LocalShare share) {
    showDialog(
      context: context,
      builder: (context) => _ShareDetailDialog(share: share),
    );
  }

  Future<void> _copyShare(LocalShare share) async {
    await Clipboard.setData(ClipboardData(text: share.shareValue));
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('份额已复制到剪贴板')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('份额管理'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              final apiService = ApiService();
              await apiService.init();
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
                final apiService = ApiService();
                await apiService.init();
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
                    const Text('本地份额数量:'),
                    Text(
                      '${_localShares.length}',
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 18,
                      ),
                    ),
                    ElevatedButton(
                      onPressed: _isLoading
                          ? null
                          : () {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(content: Text('刷新份额列表...')),
                              );
                              _loadShares();
                            },
                      child: const Text('刷新列表'),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : _localShares.isEmpty
                      ? const Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.key, size: 64, color: Colors.grey),
                              SizedBox(height: 16),
                              Text(
                                '暂无本地份额',
                                style: TextStyle(color: Colors.grey),
                              ),
                              SizedBox(height: 8),
                              Text(
                                '上传文件后会自动保存份额',
                                style: TextStyle(color: Colors.grey),
                              ),
                            ],
                          ),
                        )
                      : ListView.builder(
                          itemCount: _localShares.length,
                          itemBuilder: (context, index) {
                            final share = _localShares[index];
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
                                          '序号: ${index + 1}',
                                          style: const TextStyle(
                                            fontWeight: FontWeight.bold,
                                            fontSize: 16,
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    Text('文件名: ${share.filename}'),
                                    const SizedBox(height: 8),
                                    Container(
                                      padding: const EdgeInsets.symmetric(
                                          horizontal: 12, vertical: 8),
                                      decoration: BoxDecoration(
                                        color: Colors.grey[100],
                                        borderRadius: BorderRadius.circular(4),
                                      ),
                                      child: Text(
                                        '份额值: ${'*' * 32}...',
                                        style: const TextStyle(fontFamily: 'monospace'),
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    Text(
                                      '保存时间: ${_formatDate(share.createdAt)}',
                                      style: const TextStyle(color: Colors.grey),
                                    ),
                                    const SizedBox(height: 12),
                                    Row(
                                      children: [
                                        ElevatedButton.icon(
                                          onPressed: () =>
                                              _showShareDetail(share),
                                          icon: const Icon(Icons.visibility),
                                          label: const Text('查看份额'),
                                        ),
                                        const SizedBox(width: 12),
                                        // TODO: [上线前移除] 调试用的复制按钮，请在正式上线前移除
                                        OutlinedButton.icon(
                                          onPressed: () => _copyShare(share),
                                          icon: const Icon(Icons.copy),
                                          label: const Text('复制份额'),
                                        ),
                                      ],
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

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')} ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
  }
}

class _ShareDetailDialog extends StatefulWidget {
  final LocalShare share;

  const _ShareDetailDialog({required this.share});

  @override
  State<_ShareDetailDialog> createState() => _ShareDetailDialogState();
}

class _ShareDetailDialogState extends State<_ShareDetailDialog> {
  int _countdown = 3;
  bool _canConfirm = false;

  @override
  void initState() {
    super.initState();
    _startCountdown();
  }

  void _startCountdown() {
    Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_countdown > 0) {
        setState(() => _countdown--);
      } else {
        setState(() => _canConfirm = true);
        timer.cancel();
      }
    });
  }

  Future<void> _copyShare() async {
    await Clipboard.setData(ClipboardData(text: widget.share.shareValue));
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('份额已复制到剪贴板')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('份额详情'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('文件名: ${widget.share.filename}'),
          const SizedBox(height: 8),
          Text('保存时间: ${widget.share.createdAt.toString()}'),
          const SizedBox(height: 16),
          const Text(
            '⚠️ 注意：份额是敏感的安全信息，请确保周围无人偷窥',
            style: TextStyle(color: Colors.orange),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.grey[800],
              borderRadius: BorderRadius.circular(8),
            ),
            child: SelectableText(
              widget.share.shareValue,
              style: const TextStyle(
                fontFamily: 'monospace',
                color: Colors.white,
                fontSize: 14,
                letterSpacing: 1.2,
              ),
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        Row(
          children: [
            if (!_canConfirm)
              Text(
                '请等待 ${_countdown} 秒',
                style: const TextStyle(color: Colors.grey),
              ),
            const SizedBox(width: 8),
            ElevatedButton(
              onPressed: _canConfirm
                  ? () {
                      Navigator.pop(context);
                    }
                  : null,
              child: const Text('确认'),
            ),
          ],
        ),
        // TODO: [上线前移除] 调试用的复制按钮，请在正式上线前移除
        OutlinedButton(
          onPressed: _canConfirm ? () => _copyShare() : null,
          child: const Text('复制'),
        ),
      ],
    );
  }
}