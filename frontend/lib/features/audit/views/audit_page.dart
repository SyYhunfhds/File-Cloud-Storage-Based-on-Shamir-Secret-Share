import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../desktop/title_bar.dart';
import '../models/audit_models.dart';
import '../providers/audit_list_provider.dart';
import '../widgets/audit_filter_bar.dart';
import '../widgets/audit_table.dart';
import '../widgets/share_refresh_dialog.dart';

/// 审计管理页面
///
/// 分页查看审计条目列表，支持按状态筛选，通过/拒绝操作按钮。
class AuditPage extends ConsumerStatefulWidget {
  const AuditPage({super.key});

  @override
  ConsumerState<AuditPage> createState() => _AuditPageState();
}

class _AuditPageState extends ConsumerState<AuditPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(auditListProvider.notifier).fetch();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(auditListProvider);
    final colorScheme = Theme.of(context).colorScheme;

    return Column(
      children: [
        const AppTitleBar(),
        _buildToolbar(state),
        const AuditFilterBar(),
        const Divider(height: 1),
        Expanded(
          child: state.isLoading
              ? const Center(child: CircularProgressIndicator())
              : state.errorMessage != null
                  ? _buildError(context, state.errorMessage!)
                  : AuditTable(
                      audits: state.audits,
                      onAction: (audit, action) =>
                          _handleAction(context, audit, action),
                    ),
        ),
        if (state.totalCount > 0)
          _buildPaginationBar(state, colorScheme),
      ],
    );
  }

  Widget _buildToolbar(AuditListState state) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          Text('审计管理',
              style:
                  Theme.of(context).textTheme.titleLarge?.copyWith(fontWeight: FontWeight.w600)),
          const SizedBox(width: 12),
          IconButton(
            icon: const Icon(Icons.refresh, size: 20),
            tooltip: '刷新审计列表',
            onPressed: () => ref.read(auditListProvider.notifier).fetch(),
          ),
          const Spacer(),
          if (state.totalCount > 0)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              decoration: BoxDecoration(
                color: Theme.of(context)
                    .colorScheme
                    .primaryContainer
                    .withValues(alpha: 0.5),
                borderRadius: BorderRadius.circular(16),
              ),
              child: Text('共 ${state.totalCount} 条',
                  style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                      color: Theme.of(context).colorScheme.primary)),
            ),
        ],
      ),
    );
  }

  Widget _buildError(BuildContext context, String message) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.error_outline,
              size: 40, color: Theme.of(context).colorScheme.error),
          const SizedBox(height: 12),
          Text(message,
              style: TextStyle(
                  color: Theme.of(context).colorScheme.error)),
          const SizedBox(height: 16),
          FilledButton.tonal(
            onPressed: () => ref.read(auditListProvider.notifier).fetch(),
            child: const Text('重试'),
          ),
        ],
      ),
    );
  }

  Widget _buildPaginationBar(AuditListState state, ColorScheme colorScheme) {
    final totalPages = state.totalPages;

    return Container(
      padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
      decoration: BoxDecoration(
        border: Border(
            top: BorderSide(color: colorScheme.outlineVariant.withValues(alpha: 0.4))),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          IconButton(
            icon: const Icon(Icons.chevron_left, size: 20),
            tooltip: '上一页',
            onPressed: state.currentPage > 1 && !state.isLoading
                ? () => ref
                    .read(auditListProvider.notifier)
                    .goToPage(state.currentPage - 1)
                : null,
          ),
          Text(
            '${state.currentPage} / $totalPages 页',
            style: TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
          ),
          IconButton(
            icon: const Icon(Icons.chevron_right, size: 20),
            tooltip: '下一页',
            onPressed: state.currentPage < totalPages && !state.isLoading
                ? () => ref
                    .read(auditListProvider.notifier)
                    .goToPage(state.currentPage + 1)
                : null,
          ),
        ],
      ),
    );
  }

  void _handleAction(
      BuildContext context, LessDetailedAudit audit, String action) {
    final allIds = audit.auditIds;
    if (allIds.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('没有审计记录')),
      );
      return;
    }

    final label = action == 'pass' ? '通过' : '拒绝';
    final labelColor = action == 'pass'
        ? Colors.green.shade700
        : Theme.of(context).colorScheme.error;

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('确认$label审核'),
        content: Text(
          '条目「${audit.itemName}」(#${audit.itemId}) 的 ${allIds.length} 条审计记录。\n确定$label这些审核申请吗？',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('取消'),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: labelColor),
            onPressed: () {
              Navigator.of(ctx).pop();
              _doOperate(audit.itemId, audit.itemName, allIds, action);
            },
            child: Text(label),
          ),
        ],
      ),
    );
  }

  Future<void> _doOperate(
      int itemId, String itemName, List<int> auditIds, String action) async {
    final notifier = ref.read(auditListProvider.notifier);
    final label = action == 'pass' ? '通过' : '拒绝';
    await notifier.performOperation(
      itemId: itemId,
      auditIds: auditIds,
      action: action,
    );

    if (!mounted) return;
    final state = ref.read(auditListProvider);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(state.errorMessage ?? '审核操作已$label'),
      ),
    );

    // 通过操作成功后，询问是否刷新份额
    if (action == 'pass' && state.errorMessage == null) {
      _showRefreshPrompt(itemId, itemName);
    }
  }

  void _showRefreshPrompt(int itemId, String itemName) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('刷新份额'),
        content: Text(
          '审计操作已通过，是否立即刷新条目「$itemName」(#$itemId) 的份额？',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('稍后'),
          ),
          FilledButton(
            onPressed: () {
              Navigator.of(ctx).pop();
              ShareRefreshDialog.show(
                context,
                itemId: itemId,
                itemName: itemName,
              );
            },
            child: const Text('刷新'),
          ),
        ],
      ),
    );
  }
}