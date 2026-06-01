import 'dart:io' as io;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;
import 'package:sanitize_filename/sanitize_filename.dart';
import '../../../core/constants.dart';
import '../../../core/api_config_provider.dart';
import '../providers/download_provider.dart';

/// 条目下载页面
class DownloadPage extends ConsumerStatefulWidget {
  const DownloadPage({super.key});

  @override
  ConsumerState<DownloadPage> createState() => _DownloadPageState();
}

class _DownloadPageState extends ConsumerState<DownloadPage> {
  final _itemIdController = TextEditingController();
  final _shareController = TextEditingController();
  bool _shareVisible = false;

  @override
  void dispose() {
    _itemIdController.dispose();
    _shareController.dispose();
    super.dispose();
  }

  Future<void> _onDownload() async {
    final itemIdText = _itemIdController.text.trim();
    final share = _shareController.text.trim();

    if (itemIdText.isEmpty || share.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请输入条目ID和Share')),
      );
      return;
    }

    final itemId = int.tryParse(itemIdText);
    if (itemId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('条目ID必须为数字')),
      );
      return;
    }

    // 协议检测：HTTP 时警告
    final apiConfig = ref.read(apiConfigProvider);
    if (apiConfig.protocol == 'http') {
      final confirmed = await showDialog<bool>(
        context: context,
        builder: (ctx) => AlertDialog(
          title: const Text('安全警告'),
          content: const Text('当前网络环境不安全（HTTP），Share 将以明文传输。\n确定继续吗？'),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(ctx).pop(false),
              child: const Text('取消'),
            ),
            FilledButton(
              onPressed: () => Navigator.of(ctx).pop(true),
              child: const Text('继续'),
            ),
          ],
        ),
      );
      if (confirmed != true) return;
    }

    final success = await ref.read(downloadProvider.notifier).download(
      itemId: itemId,
      share: share,
      onFileReceived: _saveFile,
    );

    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('文件下载成功，已保存到下载文件夹')),
      );
    }
  }

  Future<void> _saveFile(List<int> data, String defaultFileName) async {
    try {
      // 安全防御：sanitizeFilename + p.basename 防路径穿越
      final safeName = sanitizeFilename(p.basename(defaultFileName));
      final fileName = safeName.trim().isEmpty ? 'unnamed_file' : safeName;

      // Windows: 保存到用户 Downloads 目录
      final downloadsPath = '${io.Platform.environment['USERPROFILE']}\\Downloads';
      var finalPath = '$downloadsPath\\$fileName';

      // 防止覆盖：重名时追加序号
      var counter = 1;
      while (io.File(finalPath).existsSync()) {
        final extIndex = fileName.lastIndexOf('.');
        if (extIndex > 0) {
          finalPath = '$downloadsPath\\${fileName.substring(0, extIndex)}_$counter${fileName.substring(extIndex)}';
        } else {
          finalPath = '$downloadsPath\\${fileName}_$counter';
        }
        counter++;
      }

      await io.File(finalPath).writeAsBytes(data);
      debugPrint('[INFO] 文件已保存: $finalPath');
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('文件保存失败: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final downloadState = ref.watch(downloadProvider);
    final colorScheme = Theme.of(context).colorScheme;

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: AppConstants.contentMaxWidth),
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: [
              Icon(Icons.download, size: 48, color: colorScheme.primary),
              const SizedBox(height: 16),
              Text(
                '条目下载',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w600,
                  color: colorScheme.onSurface,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                '输入条目ID和Share份额来下载文件',
                style: TextStyle(
                  fontSize: 13,
                  color: colorScheme.onSurfaceVariant,
                ),
              ),
              const SizedBox(height: 32),

              // 输入区
              SizedBox(
                width: 400,
                child: TextField(
                  controller: _itemIdController,
                  keyboardType: TextInputType.number,
                  decoration: InputDecoration(
                    labelText: '条目 ID',
                    border: const OutlineInputBorder(),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    isDense: true,
                    labelStyle: const TextStyle(fontSize: 14),
                  ),
                  style: const TextStyle(fontSize: 14),
                  enabled: !downloadState.isDownloading,
                ),
              ),
              const SizedBox(height: 12),
              SizedBox(
                width: 400,
                child: TextField(
                  controller: _shareController,
                  obscureText: !_shareVisible,
                  decoration: InputDecoration(
                    labelText: 'Device Share',
                    border: const OutlineInputBorder(),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    isDense: true,
                    labelStyle: const TextStyle(fontSize: 14),
                    suffixIcon: IconButton(
                      icon: Icon(
                        _shareVisible ? Icons.visibility_off : Icons.visibility,
                        size: 18,
                      ),
                      onPressed: () => setState(() => _shareVisible = !_shareVisible),
                    ),
                  ),
                  style: const TextStyle(fontSize: 14),
                  enabled: !downloadState.isDownloading,
                ),
              ),
              const SizedBox(height: 24),

              // 操作按钮
              if (downloadState.isDownloading)
                const CircularProgressIndicator()
              else
                SizedBox(
                  width: 200,
                  child: FilledButton.icon(
                    onPressed: _onDownload,
                    icon: const Icon(Icons.download, size: 18),
                    label: const Text('下载', style: TextStyle(fontSize: 14)),
                  ),
                ),

              // 错误信息
              if (downloadState.errorMessage != null) ...[
                const SizedBox(height: 12),
                Text(
                  downloadState.errorMessage!,
                  style: TextStyle(fontSize: 13, color: colorScheme.error),
                  textAlign: TextAlign.center,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
