import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/constants.dart';
import '../../../desktop/title_bar.dart';
import '../../shares/models/share_record_data.dart';
import '../../shares/services/share_service.dart';
import '../providers/share_list_provider.dart';
import '../widgets/share_list.dart';
import '../widgets/share_detail_panel.dart';
import '../../shares/providers/share_providers.dart';

/// 份额管理页面
///
/// 数据来源：[ShareStorageService]，纯本地 SharedPreferences 存储，无后端 API。
/// 布局：大屏主-从双栏，小屏单列表 + 点击弹出详情 Sheet。
class SharePage extends ConsumerStatefulWidget {
  const SharePage({super.key});

  @override
  ConsumerState<SharePage> createState() => _SharePageState();
}

class _SharePageState extends ConsumerState<SharePage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(shareListProvider.notifier).fetch();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(shareListProvider);
    final pageShares = state.currentPageShares;
    final pageIndex = (state.currentPage - 1) * sharePageSize;
    final totalPages = state.totalPages;

    final selectedRecord = state.selectedItemId != null
        ? state.shares.where((s) => s.itemId == state.selectedItemId).firstOrNull
        : null;

    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth >= AppConstants.mediumBreakpoint;

        return Column(
          children: [
            const AppTitleBar(),
            // 工具栏
            _buildToolbar(state.totalCount),
            const Divider(height: 1),
            // 主体
            Expanded(
              child: isWide && selectedRecord != null
                  ? _buildWideLayout(pageShares, pageIndex, state,
                      totalPages, selectedRecord)
                  : _buildSingleLayout(pageShares, pageIndex, state,
                      totalPages, selectedRecord),
            ),
            // 分页栏
            if (state.totalCount > 0)
              _buildPaginationBar(state, totalPages),
          ],
        );
      },
    );
  }

  // ===========================================================================
  // 工具栏
  // ===========================================================================

  Widget _buildToolbar(int totalCount) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
      child: Row(
        children: [
          Icon(Icons.vpn_key_outlined, size: 22, color: colorScheme.primary),
          const SizedBox(width: 10),
          Text('份额管理', style: textTheme.titleMedium),
          // 刷新按钮
          IconButton(
            icon: const Icon(Icons.refresh, size: 20),
            tooltip: '刷新份额列表',
            onPressed: () => ref.read(shareListProvider.notifier).fetch(),
          ),
          const SizedBox(width: 8),
          const Spacer(),
          // 统计卡片
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
            decoration: BoxDecoration(
              color: colorScheme.primaryContainer.withValues(alpha: 0.4),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Text(
              '共 $totalCount 条记录',
              style: textTheme.labelMedium?.copyWith(
                color: colorScheme.onPrimaryContainer,
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ===========================================================================
  // 大屏双栏布局
  // ===========================================================================

  Widget _buildWideLayout(
    List<dynamic> pageShares,
    int pageIndex,
    ShareListState state,
    int totalPages,
    ShareRecordData selectedRecord,
  ) {
    final colorScheme = Theme.of(context).colorScheme;
    final shareService = ref.read(shareServiceProvider);

    return Row(
      children: [
        Expanded(
          flex: 3,
          child: ShareList(
            shares: pageShares.cast(),
            pageIndex: pageIndex,
            selectedItemId: state.selectedItemId,
            isLoading: state.isLoading,
            onSelect: (id) => ref.read(shareListProvider.notifier).select(id),
            onDelete: (id) => _confirmDelete(id),
          ),
        ),
        VerticalDivider(
          width: 1,
          thickness: 1,
          color: colorScheme.outlineVariant.withValues(alpha: 0.3),
        ),
        SizedBox(
          width: 420,
          child: ShareDetailPanel(
            record: selectedRecord,
            shareService: shareService,
            onClose: () =>
                ref.read(shareListProvider.notifier).select(null),
          ),
        ),
      ],
    );
  }

  // ===========================================================================
  // 小屏单列表 + 底部 Sheet
  // ===========================================================================

  Widget _buildSingleLayout(
    List<dynamic> pageShares,
    int pageIndex,
    ShareListState state,
    int totalPages,
    ShareRecordData? selectedRecord,
  ) {
    final shareService = ref.read(shareServiceProvider);

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (selectedRecord != null && mounted) {
        _showDetailSheet(selectedRecord, shareService);
        ref.read(shareListProvider.notifier).select(null);
      }
    });

    if (state.shares.isEmpty && !state.isLoading) {
      return _buildEmptyState();
    }

    return ShareList(
      shares: pageShares.cast(),
      pageIndex: pageIndex,
      selectedItemId: state.selectedItemId,
      isLoading: state.isLoading,
      onSelect: (id) {
        final record =
            state.shares.where((s) => s.itemId == id).firstOrNull;
        if (record != null) {
          _showDetailSheet(record, shareService);
        }
      },
      onDelete: (id) => _confirmDelete(id),
    );
  }

  void _showDetailSheet(ShareRecordData record, ShareService shareService) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      constraints: BoxConstraints(
        maxHeight: MediaQuery.of(context).size.height * 0.65,
      ),
      builder: (_) => ShareDetailPanel(
        record: record,
        shareService: shareService,
        onClose: () => Navigator.of(context).pop(),
      ),
    );
  }

  // ===========================================================================
  // 空状态
  // ===========================================================================

  Widget _buildEmptyState() {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.vpn_key_off_outlined,
              size: 64, color: colorScheme.onSurfaceVariant.withValues(alpha: 0.4)),
          const SizedBox(height: 16),
          Text('暂无份额记录',
              style:
                  textTheme.titleMedium?.copyWith(color: colorScheme.onSurfaceVariant)),
          const SizedBox(height: 8),
          Text(
            '上传条目后份额将自动显示在此处',
            style: textTheme.bodyMedium
                ?.copyWith(color: colorScheme.onSurfaceVariant.withValues(alpha: 0.7)),
          ),
        ],
      ),
    );
  }

  // ===========================================================================
  // 删除确认
  // ===========================================================================

  Future<void> _confirmDelete(int itemId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('确认删除'),
        content: const Text(
            '删除后份额将无法恢复。如需下载对应条目，请确保已妥善保存 Recovery Code。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            style: TextButton.styleFrom(
              foregroundColor: Theme.of(context).colorScheme.error,
            ),
            child: const Text('删除'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await ref.read(shareListProvider.notifier).delete(itemId);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('已删除份额记录')),
        );
      }
    }
  }

  // ===========================================================================
  // 分页栏
  // ===========================================================================

  Widget _buildPaginationBar(ShareListState state, int totalPages) {
    final colorScheme = Theme.of(context).colorScheme;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: colorScheme.surface,
        border: Border(
          top: BorderSide(
              color: colorScheme.outlineVariant.withValues(alpha: 0.4)),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          TextButton.icon(
            onPressed: state.currentPage > 1
                ? () =>
                    ref.read(shareListProvider.notifier).goToPage(state.currentPage - 1)
                : null,
            icon: const Icon(Icons.chevron_left, size: 18),
            label: const Text('上一页', style: TextStyle(fontSize: 13)),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              '第 ${state.currentPage} / $totalPages 页（共 ${state.totalCount} 条）',
              style:
                  TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
            ),
          ),
          TextButton.icon(
            onPressed: state.currentPage < totalPages
                ? () =>
                    ref.read(shareListProvider.notifier).goToPage(state.currentPage + 1)
                : null,
            icon: const Icon(Icons.chevron_right, size: 18),
            label: const Text('下一页', style: TextStyle(fontSize: 13)),
          ),
        ],
      ),
    );
  }
}
