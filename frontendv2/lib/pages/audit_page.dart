import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:file_picker/file_picker.dart';

import '../models/item.dart';
import '../models/share.dart';
import '../providers/auth_provider.dart';
import '../providers/audit_provider.dart';
import '../providers/service_providers.dart';
import '../services/storage_service.dart';
import '../utils/constants.dart';

class AuditPage extends ConsumerStatefulWidget {
  const AuditPage({super.key});

  @override
  ConsumerState<AuditPage> createState() => _AuditPageState();
}

class _AuditPageState extends ConsumerState<AuditPage> {
  final StorageService _storageService = StorageService();
  bool _isLoading = false;
  bool _isUploading = false;
  int _currentPage = 1;
  final int _pageSize = 10;
  int _totalCount = 0;

  @override
  void initState() {
    super.initState();
    _initAndLoad();
  }

  Future<void> _initAndLoad() async {
    await _storageService.init();
    await _loadItems();
  }

  Future<void> _loadItems() async {
    setState(() => _isLoading = true);

    final apiService = ref.read(apiServiceProvider);
    final response = await apiService.getItemList(
      page: _currentPage,
      size: _pageSize,
    );

    if (response.isSuccess && response.data != null) {
      setState(() {
        _totalCount = response.data!.count;
      });
      ref.read(auditProvider.notifier).setItems(response.data!.items);
    }

    setState(() => _isLoading = false);
  }

  Future<void> _handleUpload() async {
    try {
      final result = await FilePicker.platform.pickFiles();

      if (result != null && result.files.isNotEmpty) {
        final file = result.files.single;
        List<int> bytes;

        if (file.bytes != null) {
          bytes = file.bytes!;
        } else if (file.path != null) {
          bytes = await File(file.path!).readAsBytes();
        } else {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(content: Text('无法获取文件')),
            );
          }
          return;
        }

        setState(() => _isUploading = true);

        final apiService = ref.read(apiServiceProvider);
        final response = await apiService.submitItem(
          bytes,
          file.name,
        );

        setState(() => _isUploading = false);

        if (response.isSuccess && response.data != null) {
          // 保存份额到本地（直接保存Base64字符串，不做任何解析）
          final localShare = LocalShare(
            filename: file.name,
            shareValue: response.data!.authShare,
            createdAt: DateTime.now(),
          );
          await _storageService.saveShare(localShare);

          _showRecoveryCodeDialog(response.data!);
          _loadItems();
        } else {
          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(response.message)),
            );
          }
        }
      }
    } catch (e) {
      setState(() => _isUploading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('选择文件失败: $e')),
        );
      }
    }
  }

  void _showRecoveryCodeDialog(ItemSubmitResponse data) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => _RecoveryCodeDialog(data: data),
    );
  }

  @override
  Widget build(BuildContext context) {
    final auditState = ref.watch(auditProvider);
    final authState = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('审计条目'),
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
              selected: true,
              onTap: () => Navigator.pop(context),
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
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          '条目总数',
                          style: TextStyle(color: Colors.grey),
                        ),
                        Text(
                          '$_totalCount',
                          style: const TextStyle(
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                    TextButton.icon(
                      onPressed: _isUploading ? null : _handleUpload,
                      icon: _isUploading
                          ? const SizedBox(
                              width: 16,
                              height: 16,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.upload_file),
                      label: Text(_isUploading ? '上传中...' : '上传文件'),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : auditState.items.isEmpty
                      ? const Center(
                          child: Column(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(Icons.folder_open,
                                  size: 64, color: Colors.grey),
                              SizedBox(height: 16),
                              Text(
                                '暂无审计条目',
                                style: TextStyle(color: Colors.grey),
                              ),
                              SizedBox(height: 8),
                              Text(
                                '点击上方"上传文件"按钮添加',
                                style: TextStyle(color: Colors.grey),
                              ),
                            ],
                          ),
                        )
                      : ListView.builder(
                          itemCount: auditState.items.length,
                          itemBuilder: (context, index) {
                            final item = auditState.items[index];
                            return Card(
                              elevation: 2,
                              margin: const EdgeInsets.symmetric(vertical: 8),
                              child: ListTile(
                                leading: const CircleAvatar(
                                  child: Icon(Icons.description),
                                ),
                                title: Text(item.filename),
                                subtitle: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '所有者: ${item.owner} | 上传者: ${item.uploader}',
                                      style: const TextStyle(fontSize: 12),
                                    ),
                                    Text(
                                      '上传时间: ${_formatDate(item.uploadedAt)}',
                                      style: const TextStyle(
                                        fontSize: 12,
                                        color: Colors.grey,
                                      ),
                                    ),
                                  ],
                                ),
                                isThreeLine: true,
                                onTap: () => _showItemDetail(item),
                              ),
                            );
                          },
                        ),
            ),
            if (_totalCount > _pageSize)
              Padding(
                padding: const EdgeInsets.only(top: 16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    IconButton(
                      icon: const Icon(Icons.chevron_left),
                      onPressed: _currentPage > 1
                          ? () {
                              setState(() => _currentPage--);
                              _loadItems();
                            }
                          : null,
                    ),
                    Text('第 $_currentPage 页'),
                    IconButton(
                      icon: const Icon(Icons.chevron_right),
                      onPressed: _currentPage * _pageSize < _totalCount
                          ? () {
                              setState(() => _currentPage++);
                              _loadItems();
                            }
                          : null,
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: _isUploading ? null : _handleUpload,
        child: _isUploading
            ? const SizedBox(
                width: 24,
                height: 24,
                child: CircularProgressIndicator(
                  color: Colors.white,
                  strokeWidth: 2,
                ),
              )
            : const Icon(Icons.add),
      ),
    );
  }

  void _showItemDetail(ItemInfo item) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(item.filename),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('所有者: ${item.owner}'),
            const SizedBox(height: 8),
            Text('上传者: ${item.uploader}'),
            const SizedBox(height: 8),
            Text('上传时间: ${_formatDate(item.uploadedAt)}'),
            const SizedBox(height: 8),
            Text('修改时间: ${_formatDate(item.changedAt)}'),
            const SizedBox(height: 16),
            const Text(
              '条目内容已使用Shamir秘密共享加密保护，需凑够足够份额才能解密查看。',
              style: TextStyle(color: Colors.grey),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('关闭'),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')} ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
  }
}

class _RecoveryCodeDialog extends StatefulWidget {
  final ItemSubmitResponse data;

  const _RecoveryCodeDialog({required this.data});

  @override
  State<_RecoveryCodeDialog> createState() => _RecoveryCodeDialogState();
}

class _RecoveryCodeDialogState extends State<_RecoveryCodeDialog> {
  bool _isObscured = true;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('上传成功'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('文件名: ${widget.data.name}'),
          const SizedBox(height: 16),
          const Text(
            '请务必保存以下Recovery Code，这是恢复份额的唯一方式：',
            style: TextStyle(color: Colors.red, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 12),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 16),
            decoration: BoxDecoration(
              color: Colors.grey[800],
              borderRadius: BorderRadius.circular(8),
            ),
            child: Row(
              children: [
                Expanded(
                  child: SelectableText(
                    _isObscured
                        ? widget.data.recoveryCode.replaceAll(
                            RegExp(r'.'), '*')
                        : widget.data.recoveryCode,
                    style: const TextStyle(
                      fontFamily: 'monospace',
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: 1.2,
                      fontSize: 14,
                    ),
                  ),
                ),
                IconButton(
                  icon: Icon(
                    _isObscured ? Icons.visibility : Icons.visibility_off,
                    color: Colors.white70,
                  ),
                  onPressed: () {
                    setState(() => _isObscured = !_isObscured);
                  },
                ),
              ],
            ),
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              TextButton.icon(
                onPressed: () async {
                  await Clipboard.setData(
                    ClipboardData(text: widget.data.recoveryCode),
                  );
                  if (mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('已复制到剪贴板')),
                    );
                  }
                },
                icon: const Icon(Icons.copy),
                label: const Text('复制'),
              ),
            ],
          ),
        ],
      ),
      actions: [
        ElevatedButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('我已保存'),
        ),
      ],
    );
  }
}