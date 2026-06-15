import 'dart:io' as io;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;
import 'package:sanitize_filename/sanitize_filename.dart';
import '../../items/models/item_models.dart';
import '../../../core/constants.dart';
import '../../../core/api_config_provider.dart';
import '../../items/services/item_api_service.dart';
import '../../auth/providers/auth_provider.dart';
import '../../download/providers/download_provider.dart';
import '../../shares/providers/share_providers.dart';
import '../../shares/services/share_storage_service.dart';
import '../../shares/services/crypto_service.dart';
import '../providers/entry_provider.dart';
import 'force_refresh_share_dialog.dart';

/// 财务条目数据表格
///
/// 表头与数据行分离，数据切换使用 AnimatedSwitcher 防闪烁 + 动效。
class EntryTable extends ConsumerStatefulWidget {
  final List<ItemInfo> entries;
  final bool isLoading;

  /// 当前页第一条的全局序号偏移，用于跨页连续序号。
  /// 例如第 1 页 pageIndex=0，第 2 页 pageIndex=20。
  final int pageIndex;

  const EntryTable({
    super.key,
    required this.entries,
    this.isLoading = false,
    this.pageIndex = 0,
  });

  @override
  ConsumerState<EntryTable> createState() => _EntryTableState();
}

class _EntryTableState extends ConsumerState<EntryTable> {
  Future<void> _handleDownload(ItemInfo entry) async {
    if (!entry.canDownload) return;

    final shareService = ref.read(shareServiceProvider);

    // 先读取原始份额记录，获取原始文件名
    final record = await shareService.getShareRecord(entry.itemId);
    final share = record != null
        ? await shareService.getShareForDownload(entry.itemId)
        : null;
    final originalName =
        record?.originalFilename ?? 'item_${entry.itemId}';

    if (share == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('未找到该条目的份额信息，请重新上传')),
        );
      }
      return;
    }

    final success = await ref.read(downloadProvider.notifier).download(
      itemId: entry.itemId,
      share: share,
      defaultFileName: originalName,
      onFileReceived: _saveFile,
    );

    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('条目 #${entry.itemId} ($originalName) 下载成功')),
      );
    }
  }

  Future<void> _handleReport(ItemInfo entry) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('申请下载权限'),
        content: Text('确定要申请条目 #${entry.itemId} 的下载权限吗？'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('确定'),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    final authState = ref.read(authProvider);
    final apiConfig = ref.read(apiConfigProvider);
    final service = ItemApiService(apiConfig.baseUrl);

    final resp = await service.applyForViewing(
      itemIds: [entry.itemId],
      token: authState.token,
    );

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(resp.isSuccess
              ? '已提交 ${resp.data?.totalApplied ?? 1} 条申请'
              : '申请失败: ${resp.message}'),
        ),
      );
    }
  }

  Future<void> _handleDelete(ItemInfo entry) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('删除条目'),
        content: Text('确定要删除「${entry.filename}」(#${entry.itemId}) 吗？\n此操作不可撤销。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: Text('删除',
                style: TextStyle(color: Theme.of(context).colorScheme.error)),
          ),
        ],
      ),
    );

    if (confirmed != true || !mounted) return;

    final authState = ref.read(authProvider);
    final apiConfig = ref.read(apiConfigProvider);
    final service = ItemApiService(apiConfig.baseUrl);

    final resp = await service.deleteItems(
      itemIds: [entry.itemId],
      token: authState.token,
    );

    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(resp.isSuccess
            ? '已删除 ${resp.data?.totalDeleted ?? 1} 条'
            : '删除失败: ${resp.message}'),
      ),
    );

    if (resp.isSuccess) {
      _refreshList();
    }
  }

  Future<void> _handleEdit(ItemInfo entry) async {
    // 加载条目详情用于预填充
    final detail =
        await ref.read(entryListProvider.notifier).fetchItemDetail(entry.itemId);
    if (!mounted) return;

    final result = await showDialog<ItemUpdateReq>(
      context: context,
      builder: (ctx) => _ItemEditDialog(entry: entry, detail: detail),
    );
    if (result == null || !mounted) return;

    final notifier = ref.read(entryListProvider.notifier);
    final success = await notifier.updateItem(result);
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(success ? '更新成功' : '更新失败')),
    );
    if (success) _refreshList();
  }

  void _refreshList() {
    final mode = ref.read(entryListProvider).filterMode;
    switch (mode) {
      case EntryFilterMode.my:
        ref.read(entryListProvider.notifier).fetchMyEntries();
      case EntryFilterMode.public:
        ref.read(entryListProvider.notifier).fetchPublicEntries();
      case EntryFilterMode.all:
        ref.read(entryListProvider.notifier).fetchAllEntries();
    }
  }

  Future<void> _saveFile(List<int> data, String defaultFileName) async {
    try {
      // 防御层 1: p.basename() 剥离目录前缀，消除绝对路径和 .. 跳转
      // 防御层 2: sanitizeFilename() 移除所有系统非法字符
      final safeName = sanitizeFilename(p.basename(defaultFileName));
      final fileName = safeName.trim().isEmpty ? 'unnamed_file' : safeName;

      final downloadsPath =
          '${io.Platform.environment['USERPROFILE']}\\Downloads';
      var finalPath = '$downloadsPath\\$fileName';
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
    final colorScheme = Theme.of(context).colorScheme;
    final entries = widget.entries;
    final isLoading = widget.isLoading;

    if (entries.isEmpty && !isLoading) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.inbox_outlined, size: 64, color: colorScheme.onSurfaceVariant),
            const SizedBox(height: 12),
            Text(
              '暂无条目',
              style: TextStyle(fontSize: 15, color: colorScheme.onSurfaceVariant),
            ),
          ],
        ),
      );
    }

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: AppConstants.contentMaxWidth),
        child: Column(
          children: [
            // 表头（始终渲染，不随数据重建）
            _buildHeader(colorScheme),
            // 数据行 / 加载指示器（带切换动效）
            Expanded(
              child: AnimatedSwitcher(
                duration: const Duration(milliseconds: 200),
                switchInCurve: Curves.easeOut,
                switchOutCurve: Curves.easeIn,
                transitionBuilder: (child, animation) => FadeTransition(
                  opacity: animation,
                  child: child,
                ),
                child: isLoading
                    ? const Center(
                        key: ValueKey('loading'),
                        child: Padding(
                          padding: EdgeInsets.all(32),
                          child: CircularProgressIndicator(),
                        ),
                      )
                    : ListView(
                        key: ValueKey('rows_${entries.length}'),
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        children: entries.asMap().entries.map((entry) {
                          final index = entry.key;
                          final item = entry.value;
                          return _buildDataRow(index, item, colorScheme);
                        }).toList(),
                      ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// 表头行
  Widget _buildHeader(ColorScheme colorScheme) {
    return Container(
      key: const ValueKey('table_header'),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: colorScheme.surfaceContainerHighest,
        border: Border(
          bottom: BorderSide(color: colorScheme.outlineVariant),
        ),
      ),
      child: Row(
        children: [
          _headerCell('序号', 50),
          _headerCell('文件名', 150),
          _headerCell('所有者', 100),
          _headerCell('上传者', 100),
          _headerCell('上传时间', 110),
          _headerCell('修改时间', 110),
          _headerCell('可下载', 60),
          _headerCell('操作', 280),
        ],
      ),
    );
  }

  Widget _headerCell(String label, double width) {
    return SizedBox(
      width: width,
      child: Text(
        label,
        style: TextStyle(
          fontWeight: FontWeight.w600,
          fontSize: 13,
          color: Theme.of(context).colorScheme.onSurfaceVariant,
        ),
      ),
    );
  }

  /// 单行数据
  Widget _buildDataRow(int index, ItemInfo entry, ColorScheme colorScheme) {
    return Container(
      decoration: BoxDecoration(
        border: Border(
          bottom: BorderSide(color: colorScheme.outlineVariant.withValues(alpha: 0.3)),
        ),
      ),
      child: Row(
        children: [
          _dataCell('${widget.pageIndex + index + 1}', 50),
          _dataCell(entry.filename, 150, overflow: TextOverflow.ellipsis),
          _dataCell(entry.owner, 100),
          _dataCell(entry.uploader, 100),
          _dataCell(_formatDate(entry.uploadedAt), 110),
          _dataCell(_formatDate(entry.changedAt), 110),
          SizedBox(
            width: 60,
            child: Icon(
              entry.canDownload ? Icons.check_circle : Icons.cancel_outlined,
              size: 16,
              color: entry.canDownload ? Colors.green : colorScheme.onSurfaceVariant,
            ),
          ),
          SizedBox(
            width: 280,
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (entry.owner == ref.read(authProvider).userName)
                  _actionButton('强制刷新份额',
                      () => _handleForceRefresh(entry)),
                const SizedBox(width: 4),
                _actionButton('详情', () {
                  debugPrint('[详情] itemId=${entry.itemId}');
                }),
                const SizedBox(width: 4),
                _actionButton('修改', () => _handleEdit(entry)),
                const SizedBox(width: 4),
                _actionButton(
                  entry.canDownload ? '下载' : '申请下载权限',
                  () {
                    if (entry.canDownload) {
                      _handleDownload(entry);
                    } else {
                      _handleReport(entry);
                    }
                  },
                ),
                const SizedBox(width: 4),
                _actionButton('删除', () => _handleDelete(entry)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _dataCell(String text, double width, {TextOverflow overflow = TextOverflow.visible}) {
    return SizedBox(
      width: width,
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 10),
        child: Text(
          text,
          style: const TextStyle(fontSize: 13),
          overflow: overflow,
        ),
      ),
    );
  }

  Future<void> _handleForceRefresh(ItemInfo entry) async {
    final auth = ref.read(authProvider);
    final combinedId = '${auth.userId}_${auth.userName}';

    // 从 Hive 加载并解密 Device Share
    final storage = ShareStorageService();
    final record = await storage.get(combinedId, entry.itemId);
    if (record == null) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('未找到该条目的本地份额，无法刷新')),
        );
      }
      return;
    }
    final deviceShare = await ShareCryptoService.decrypt(
      record.encryptedShare,
      combinedId,
    );

    if (context.mounted) {
      await ForceRefreshShareDialog.show(
        context,
        itemId: entry.itemId,
        itemName: entry.filename,
        deviceShare: deviceShare,
      );
    }
  }

  Widget _actionButton(String label, VoidCallback? onPressed) {
    return TextButton(
      onPressed: onPressed,
      style: TextButton.styleFrom(
        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
        minimumSize: Size.zero,
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
        foregroundColor: onPressed == null
            ? Theme.of(context).colorScheme.onSurfaceVariant.withValues(alpha: 0.4)
            : null,
      ),
      child: Text(label, style: const TextStyle(fontSize: 12)),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }
}

// =============================================================================
// 条目修改对话框
// =============================================================================

class _ItemEditDialog extends StatefulWidget {
  final ItemInfo entry;
  final DetailItemInfo? detail;

  const _ItemEditDialog({required this.entry, this.detail});

  @override
  State<_ItemEditDialog> createState() => _ItemEditDialogState();
}

class _ItemEditDialogState extends State<_ItemEditDialog> {
  late final TextEditingController _filenameCtrl;
  late int _minimumPrivilege;
  bool _enablePublic = false;

  @override
  void initState() {
    super.initState();
    final d = widget.detail;
    _filenameCtrl = TextEditingController(
        text: d?.filename ?? widget.entry.filename);
    _minimumPrivilege = d?.minimumPrivilege ?? 1;
    _enablePublic = d?.isPublic ?? false;
  }

  @override
  void dispose() {
    _filenameCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return AlertDialog(
      title: const Text('修改条目'),
      content: SizedBox(
        width: 380,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 条目标识
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: colorScheme.surfaceContainerHighest.withValues(alpha: 0.5),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                '条目 #${widget.entry.itemId}: ${widget.entry.filename}',
                style: const TextStyle(fontSize: 13, fontFamily: 'monospace'),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(height: 20),

            // 文件名
            TextField(
              controller: _filenameCtrl,
              decoration: const InputDecoration(
                labelText: '文件名',
                helperText: '仅修改前端显示名称，不涉及物理文件',
                border: OutlineInputBorder(),
                isDense: true,
              ),
              style: const TextStyle(fontSize: 13),
            ),
            const SizedBox(height: 12),

            // 最低权限等级下拉
            DropdownButtonFormField<int>(
              value: _minimumPrivilege,
              decoration: const InputDecoration(
                labelText: '最低权限等级',
                helperText: '低于此权限的用户无法下载',
                border: OutlineInputBorder(),
                isDense: true,
              ),
              items: List.generate(5, (i) => i + 1)
                  .map((v) => DropdownMenuItem(
                        value: v,
                        child: Text('$v${v == 1 ? ' (默认)' : ''}',
                            style: const TextStyle(fontSize: 13)),
                      ))
                  .toList(),
              onChanged: (v) {
                if (v != null) setState(() => _minimumPrivilege = v);
              },
            ),
            const SizedBox(height: 16),

            // 公开搜索开关
            SwitchListTile(
              title: const Text('允许公开搜索', style: TextStyle(fontSize: 13)),
              subtitle: const Text('开启后所有用户可搜索到该条目',
                  style: TextStyle(fontSize: 11)),
              value: _enablePublic,
              onChanged: (v) => setState(() => _enablePublic = v),
              contentPadding: EdgeInsets.zero,
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('取消'),
        ),
        FilledButton(
          onPressed: () {
            final newName = _filenameCtrl.text.trim();
            Navigator.of(context).pop(ItemUpdateReq(
              itemId: widget.entry.itemId,
              newFilename: newName.isNotEmpty ? newName : null,
              minimumPrivilege: _minimumPrivilege,
              enablePublic: _enablePublic,
            ));
          },
          child: const Text('保存'),
        ),
      ],
    );
  }
}
