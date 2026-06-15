import 'dart:io' as io;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path/path.dart' as p;
import 'package:sanitize_filename/sanitize_filename.dart';
import '../../../core/constants.dart';
import '../../download/providers/download_provider.dart';
import '../providers/server_share_provider.dart';
import '../models/server_share_models.dart';

/// 服务器份额页面
///
/// 大屏 (≥1000px): 左侧份额列表 + 右侧拉取结果
/// 窄屏 (<1000px): 单列堆叠
class ServerSharePage extends ConsumerStatefulWidget {
  const ServerSharePage({super.key});

  @override
  ConsumerState<ServerSharePage> createState() => _ServerSharePageState();
}

class _ServerSharePageState extends ConsumerState<ServerSharePage> {
  bool _hasInit = false;

  @override
  void initState() {
    super.initState();
  }

  // ===========================================================================
  // 文件保存（复制自 entry_table.dart _saveFile）
  // ===========================================================================

  Future<void> _saveFile(List<int> data, String defaultFileName) async {
    try {
      final safeName = sanitizeFilename(p.basename(defaultFileName));
      final fileName = safeName.trim().isEmpty ? 'unnamed_file' : safeName;

      final downloadsPath =
          '${io.Platform.environment['USERPROFILE']}\\Downloads';
      var finalPath = '$downloadsPath\\$fileName';
      var counter = 1;
      while (io.File(finalPath).existsSync()) {
        final extIndex = fileName.lastIndexOf('.');
        if (extIndex > 0) {
          finalPath =
              '$downloadsPath\\${fileName.substring(0, extIndex)}_$counter${fileName.substring(extIndex)}';
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

  // ===========================================================================
  // 下载
  // ===========================================================================

  Future<void> _handleDownload(PulledShare pulled) async {
    final success = await ref.read(downloadProvider.notifier).download(
      itemId: pulled.itemId,
      share: pulled.deviceShare,
      defaultFileName: pulled.filename,
      onFileReceived: _saveFile,
    );

    if (success && mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${pulled.filename} 下载成功'),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final screenWidth = MediaQuery.sizeOf(context).width;
    final isWide = screenWidth >= AppConstants.mediumBreakpoint;

    if (!_hasInit) {
      _hasInit = true;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        ref.read(serverShareProvider.notifier).fetch();
      });
    }

    return isWide
        ? _buildWideLayout(colorScheme)
        : _buildNarrowLayout(colorScheme);
  }

  // ===========================================================================
  // 大屏布局：左侧列表 + 右侧拉取结果
  // ===========================================================================

  Widget _buildWideLayout(ColorScheme colorScheme) {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildHeader(),
          const SizedBox(height: 16),
          Expanded(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(flex: 2, child: _buildShareListCard(colorScheme)),
                const SizedBox(width: 16),
                Expanded(flex: 1, child: _buildPulledResultCard(colorScheme)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ===========================================================================
  // 窄屏布局
  // ===========================================================================

  Widget _buildNarrowLayout(ColorScheme colorScheme) {
    return Padding(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildHeader(),
          const SizedBox(height: 12),
          Expanded(child: _buildShareListCard(colorScheme)),
          const SizedBox(height: 12),
          Expanded(child: _buildPulledResultCard(colorScheme)),
        ],
      ),
    );
  }

  // ===========================================================================
  // 页面标题
  // ===========================================================================

  Widget _buildHeader() {
    return Row(
      children: [
        Icon(Icons.cloud_download, size: 24, color: Theme.of(context).colorScheme.primary),
        const SizedBox(width: 8),
        Text(
          '服务器份额',
          style: Theme.of(context).textTheme.titleLarge,
        ),
      ],
    );
  }

  // ===========================================================================
  // 左侧：远端份额列表卡片
  // ===========================================================================

  Widget _buildShareListCard(ColorScheme colorScheme) {
    final state = ref.watch(serverShareProvider);
    final notifier = ref.read(serverShareProvider.notifier);
    final totalPages = state.total == 0 ? 0 : ((state.total - 1) ~/ state.pageSize + 1);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 工具栏
            _buildListToolbar(state, notifier, totalPages),
            const Divider(),
            // 列表
            Expanded(child: _buildShareTable(state, notifier)),
          ],
        ),
      ),
    );
  }

  Widget _buildListToolbar(
    ServerShareState state,
    ServerShareNotifier notifier,
    int totalPages,
  ) {
    final isAllSelected = state.shares.isNotEmpty &&
        state.selectedShareIds.length == state.shares.length;

    return Row(
      children: [
        Text('共 ${state.total} 项', style: const TextStyle(fontSize: 13)),
        const SizedBox(width: 12),
        TextButton.icon(
          onPressed: isAllSelected ? notifier.clearSelection : notifier.selectAll,
          icon: Icon(isAllSelected ? Icons.deselect : Icons.select_all, size: 18),
          label: Text(isAllSelected ? '取消全选' : '全选'),
        ),
        const SizedBox(width: 8),
        FilledButton.icon(
          onPressed: state.isPulling || state.selectedShareIds.isEmpty
              ? null
              : () => notifier.pullSelected(),
          icon: const Icon(Icons.download, size: 18),
          label: Text(state.isPulling
              ? '拉取中…'
              : '拉取选中 (${state.selectedShareIds.length})'),
        ),
        const Spacer(),
        // 页码
        IconButton(
          onPressed: state.currentPage > 1 && !state.isLoading
              ? () => notifier.loadPage(state.currentPage - 1)
              : null,
          icon: const Icon(Icons.chevron_left),
          tooltip: '上一页',
        ),
        Text('${state.currentPage}/$totalPages', style: const TextStyle(fontSize: 13)),
        IconButton(
          onPressed: state.currentPage < totalPages && !state.isLoading
              ? () => notifier.loadPage(state.currentPage + 1)
              : null,
          icon: const Icon(Icons.chevron_right),
          tooltip: '下一页',
        ),
      ],
    );
  }

  Widget _buildShareTable(ServerShareState state, ServerShareNotifier notifier) {
    if (state.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.errorMessage != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, size: 48, color: Colors.red.shade300),
            const SizedBox(height: 12),
            Text(state.errorMessage!, style: const TextStyle(color: Colors.red)),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: () => notifier.fetch(),
              child: const Text('重试'),
            ),
          ],
        ),
      );
    }

    if (state.shares.isEmpty) {
      return const Center(
        child: Text('暂无待拉取的份额', style: TextStyle(color: Colors.grey)),
      );
    }

    return ListView(
      children: [
        // 表头
        _buildTableHeader(),
        ...state.shares.map((share) => _buildTableRow(share, state, notifier)),
      ],
    );
  }

  Widget _buildTableHeader() {
    final textStyle = TextStyle(
      fontSize: 12,
      fontWeight: FontWeight.bold,
      color: Colors.grey.shade600,
    );
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          const SizedBox(width: 48),
          Expanded(flex: 1, child: Text('#', style: textStyle)),
          Expanded(flex: 1, child: Text('条目ID', style: textStyle)),
          Expanded(flex: 3, child: Text('文件名', style: textStyle)),
          Expanded(flex: 2, child: Text('所有者', style: textStyle)),
          Expanded(flex: 1, child: Text('状态', style: textStyle)),
          Expanded(flex: 2, child: Text('过期时间', style: textStyle)),
        ],
      ),
    );
  }

  Widget _buildTableRow(
    ServerShareInfo share,
    ServerShareState state,
    ServerShareNotifier notifier,
  ) {
    final isSelected = state.selectedShareIds.contains(share.shareId);
    final isExpired = share.isExpired;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: InkWell(
        onTap: () => notifier.toggleSelection(share.shareId),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8),
          child: Row(
            children: [
              SizedBox(
                width: 48,
                child: Checkbox(
                  value: isSelected,
                  onChanged: (_) => notifier.toggleSelection(share.shareId),
                ),
              ),
              Expanded(
                flex: 1,
                child: Text('${share.shareId}',
                    style: const TextStyle(fontSize: 13)),
              ),
              Expanded(
                flex: 1,
                child: Text('${share.itemId}',
                    style: const TextStyle(fontSize: 13)),
              ),
              Expanded(
                flex: 3,
                child: Text(share.filename,
                    style: const TextStyle(fontSize: 13),
                    overflow: TextOverflow.ellipsis),
              ),
              Expanded(
                flex: 2,
                child: Text(share.owner,
                    style: const TextStyle(fontSize: 13),
                    overflow: TextOverflow.ellipsis),
              ),
              Expanded(
                flex: 1,
                child: isExpired
                    ? Text('已过期',
                        style: TextStyle(fontSize: 13, color: Colors.red.shade400))
                    : Text('有效',
                        style: TextStyle(fontSize: 13, color: Colors.green.shade600)),
              ),
              Expanded(
                flex: 2,
                child: Text(
                  share.expireAt.length >= 19
                      ? share.expireAt.substring(0, 19).replaceFirst('T', ' ')
                      : share.expireAt,
                  style: const TextStyle(fontSize: 12, color: Colors.grey),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // ===========================================================================
  // 右侧：拉取结果卡片
  // ===========================================================================

  Widget _buildPulledResultCard(ColorScheme colorScheme) {
    final state = ref.watch(serverShareProvider);
    final notifier = ref.read(serverShareProvider.notifier);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 标题栏
            Row(
              children: [
                const Icon(Icons.check_circle_outline, size: 20),
                const SizedBox(width: 8),
                const Text('拉取结果', style: TextStyle(fontWeight: FontWeight.w600)),
                const Spacer(),
                if (state.pulledShares.isNotEmpty)
                  TextButton(
                    onPressed: notifier.resetPulled,
                    child: const Text('清除'),
                  ),
              ],
            ),
            const Divider(),
            // 拉取状态
            if (state.isPulling)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Center(
                    child: Column(children: [
                  CircularProgressIndicator(),
                  SizedBox(height: 12),
                  Text('正在拉取…'),
                ])),
              ),
            // 拉取错误
            if (state.pullError != null && !state.isPulling)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: Center(
                  child: Column(children: [
                    Icon(Icons.error_outline, size: 32,
                        color: Colors.red.shade300),
                    const SizedBox(height: 8),
                    Text(state.pullError!,
                        style: const TextStyle(color: Colors.red)),
                  ]),
                ),
              ),
            // 拉取结果为空
            if (state.pulledShares.isEmpty &&
                !state.isPulling &&
                state.pullError == null)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 24),
                child: Center(
                  child: Text('暂无拉取结果',
                      style: TextStyle(color: Colors.grey)),
                ),
              ),
            // 拉取结果列表
            if (state.pulledShares.isNotEmpty)
              Expanded(
                child: ListView.builder(
                  itemCount: state.pulledShares.length,
                  itemBuilder: (context, index) {
                    final pulled = state.pulledShares[index];
                    return _buildPulledRow(pulled);
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildPulledRow(PulledShare pulled) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Icon(Icons.insert_drive_file_outlined,
              size: 18, color: Colors.grey.shade600),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(pulled.filename,
                    style: const TextStyle(fontSize: 13),
                    overflow: TextOverflow.ellipsis),
                Text('#${pulled.itemId}',
                    style: TextStyle(fontSize: 11, color: Colors.grey.shade600)),
              ],
            ),
          ),
          FilledButton.tonal(
            onPressed: () => _handleDownload(pulled),
            child: const Text('下载'),
          ),
        ],
      ),
    );
  }
}
