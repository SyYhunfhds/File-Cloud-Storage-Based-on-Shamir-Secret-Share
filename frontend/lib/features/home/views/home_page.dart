import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../items/models/item_models.dart';
import '../providers/entry_provider.dart';
import '../widgets/toolbar.dart';
import '../widgets/entry_table.dart';
import '../../upload/widgets/upload_fab.dart';
import '../../../desktop/title_bar.dart';

/// 主页 — 条目审计视图
///
/// 侧边栏由全局层 AppShell 提供，本页仅包含内容区。
/// 布局：标题栏 + 工具栏 + 筛选按钮组 + 数据表格 + 分页栏 + FAB 悬浮按钮
class HomePage extends ConsumerStatefulWidget {
  const HomePage({super.key});

  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> {
  final _searchController = TextEditingController();
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(entryListProvider.notifier).fetchAllEntries();
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<ItemInfo> _getFilteredEntries(EntryListState state) {
    if (_searchQuery.isEmpty) return state.entries;
    return ref.read(entryListProvider.notifier).searchByFilename(_searchQuery);
  }

  void _onFilterModeChanged(EntryFilterMode mode) {
    switch (mode) {
      case EntryFilterMode.my:
        ref.read(entryListProvider.notifier).fetchMyEntries();
      case EntryFilterMode.public:
        ref.read(entryListProvider.notifier).fetchPublicEntries();
      case EntryFilterMode.all:
        ref.read(entryListProvider.notifier).fetchAllEntries();
    }
  }

  @override
  Widget build(BuildContext context) {
    final entryState = ref.watch(entryListProvider);
    final notifier = ref.read(entryListProvider.notifier);
    final filteredEntries = _getFilteredEntries(entryState);
    final colorScheme = Theme.of(context).colorScheme;

    final pageIndex = (entryState.currentPage - 1) * pageSize;
    final totalPages = notifier.totalPages;

    return Stack(
      children: [
        Column(
          children: [
            const AppTitleBar(),
            HomeToolbar(
              searchController: _searchController,
              onSearchChanged: (value) {
                setState(() => _searchQuery = value);
              },
            ),
            // 条目来源筛选按钮组
            _buildFilterModeBar(entryState.filterMode, colorScheme),
            // 数据表格
            Expanded(
              child: EntryTable(
                entries: filteredEntries,
                isLoading: entryState.isLoading,
                pageIndex: pageIndex,
              ),
            ),
            // 分页栏
            if (entryState.totalCount > 0)
              _buildPaginationBar(entryState, totalPages, colorScheme),
          ],
        ),
        // 上传条目 FAB
        const Positioned(
          right: 24,
          bottom: 24,
          child: UploadFab(),
        ),
      ],
    );
  }

  /// 底部翻页栏
  Widget _buildPaginationBar(
    EntryListState state,
    int totalPages,
    ColorScheme colorScheme,
  ) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: colorScheme.surface,
        border: Border(
          top: BorderSide(color: colorScheme.outlineVariant.withValues(alpha: 0.4)),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          TextButton.icon(
            onPressed: state.currentPage > 1
                ? () => ref.read(entryListProvider.notifier).goToPage(state.currentPage - 1)
                : null,
            icon: const Icon(Icons.chevron_left, size: 18),
            label: const Text('上一页', style: TextStyle(fontSize: 13)),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Text(
              '第 ${state.currentPage} / $totalPages 页（共 ${state.totalCount} 条）',
              style: TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
            ),
          ),
          TextButton.icon(
            onPressed: state.currentPage < totalPages
                ? () => ref.read(entryListProvider.notifier).goToPage(state.currentPage + 1)
                : null,
            icon: const Icon(Icons.chevron_right, size: 18),
            label: const Text('下一页', style: TextStyle(fontSize: 13)),
          ),
        ],
      ),
    );
  }

  /// 条目来源筛选按钮组 — "我的条目" / "公开条目" / "所有可见条目"
  Widget _buildFilterModeBar(EntryFilterMode currentMode, ColorScheme colorScheme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: colorScheme.surface,
        border: Border(
          bottom: BorderSide(color: colorScheme.outlineVariant.withValues(alpha: 0.4)),
        ),
      ),
      child: Row(
        children: [
          Text(
            '条目来源：',
            style: TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
          ),
          const SizedBox(width: 12),
          SegmentedButton<EntryFilterMode>(
            segments: const [
              ButtonSegment(
                value: EntryFilterMode.my,
                label: Text('我的条目', style: TextStyle(fontSize: 13)),
                icon: Icon(Icons.person, size: 16),
              ),
              ButtonSegment(
                value: EntryFilterMode.public,
                label: Text('公开条目', style: TextStyle(fontSize: 13)),
                icon: Icon(Icons.public, size: 16),
              ),
              ButtonSegment(
                value: EntryFilterMode.all,
                label: Text('所有可见条目', style: TextStyle(fontSize: 13)),
                icon: Icon(Icons.list, size: 16),
              ),
            ],
            selected: {currentMode},
            onSelectionChanged: (selection) {
              _onFilterModeChanged(selection.first);
            },
            style: ButtonStyle(
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              visualDensity: VisualDensity.compact,
            ),
          ),
        ],
      ),
    );
  }
}
