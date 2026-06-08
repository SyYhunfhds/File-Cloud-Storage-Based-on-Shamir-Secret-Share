import 'package:flutter/material.dart';

import '../models/audit_models.dart';

/// 审计数据表格
///
/// 参照 `share_list.dart` 的列表模式，使用 `ListView.builder` 渲染。
class AuditTable extends StatelessWidget {
  final List<LessDetailedAudit> audits;
  final void Function(LessDetailedAudit audit, String action)? onAction;

  const AuditTable({
    super.key,
    required this.audits,
    this.onAction,
  });

  @override
  Widget build(BuildContext context) {
    if (audits.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.inbox_outlined,
                size: 48,
                color: Theme.of(context)
                    .colorScheme
                    .onSurfaceVariant
                    .withValues(alpha: 0.5)),
            const SizedBox(height: 8),
            Text('暂无审计数据',
                style: TextStyle(
                    color: Theme.of(context)
                        .colorScheme
                        .onSurfaceVariant
                        .withValues(alpha: 0.7))),
          ],
        ),
      );
    }

    return Column(
      children: [
        _buildHeader(context),
        const Divider(height: 1),
        Expanded(
          child: ListView.builder(
            itemCount: audits.length,
            itemBuilder: (context, index) =>
                _buildRow(context, audits[index]),
          ),
        ),
      ],
    );
  }

  Widget _buildHeader(BuildContext context) {
    final textStyle = Theme.of(context)
        .textTheme
        .labelSmall
        ?.copyWith(color: Theme.of(context).colorScheme.onSurfaceVariant);
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: Row(
        children: [
          SizedBox(width: 60, child: Text('条目ID', style: textStyle)),
          Expanded(child: Text('条目名称', style: textStyle)),
          SizedBox(width: 80, child: Text('状态', style: textStyle)),
          SizedBox(width: 72, child: Text('数量', style: textStyle)),
          SizedBox(width: 80, child: Text('申请人', style: textStyle)),
          SizedBox(width: 140, child: Text('创建时间', style: textStyle)),
          SizedBox(width: 144, child: Text('操作', style: textStyle)),
        ],
      ),
    );
  }

  Widget _buildRow(BuildContext context, LessDetailedAudit audit) {
    final colorScheme = Theme.of(context).colorScheme;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: Container(
        padding:
            const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          border: Border(
            bottom: BorderSide(
                color: colorScheme.outlineVariant.withValues(alpha: 0.4)),
          ),
        ),
        child: Row(
          children: [
            // 条目ID
            SizedBox(
              width: 60,
              child: Text(
                '#${audit.itemId}',
                style: TextStyle(
                  color: colorScheme.primary,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
            // 条目名称
            Expanded(
              child: Tooltip(
                message: audit.itemName,
                child: Text(
                  audit.itemName,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ),
            // 状态
            SizedBox(width: 80, child: _buildStatusChip(context, audit)),
            // 审计数量
            SizedBox(
              width: 72,
              child: Text(
                '${audit.auditCount} 条',
                style: TextStyle(color: colorScheme.onSurfaceVariant),
              ),
            ),
            // 申请人
            SizedBox(
              width: 80,
              child: Text(
                audit.latestApplicant,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            // 创建时间
            SizedBox(
              width: 140,
              child: Text(
                audit.latestCreatedAt,
                style:
                    TextStyle(fontSize: 12, color: colorScheme.onSurfaceVariant),
              ),
            ),
            // 操作按钮
            SizedBox(
              width: 144,
              child: Row(
                children: [
                  SizedBox(
                    height: 28,
                    child: FilledButton.tonal(
                      style: FilledButton.styleFrom(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 10, vertical: 0),
                        minimumSize: Size.zero,
                        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                      ),
                      onPressed: () => onAction?.call(audit, 'pass'),
                      child: const Text('通过', style: TextStyle(fontSize: 12)),
                    ),
                  ),
                  const SizedBox(width: 6),
                  SizedBox(
                    height: 28,
                    child: OutlinedButton(
                      style: OutlinedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 10, vertical: 0),
                        minimumSize: Size.zero,
                        tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        foregroundColor: colorScheme.error,
                        side: BorderSide(color: colorScheme.error),
                      ),
                      onPressed: () => onAction?.call(audit, 'reject'),
                      child: const Text('拒绝', style: TextStyle(fontSize: 12)),
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

  Widget _buildStatusChip(BuildContext context, LessDetailedAudit audit) {
    final (label, bgColor, fgColor) = switch (audit.status) {
      AggregateStatus.pending => ('待审查', Colors.orange.shade100, Colors.orange.shade800),
      AggregateStatus.approved => ('已通过', Colors.green.shade100, Colors.green.shade800),
      AggregateStatus.rejected => ('已拒绝', Colors.red.shade100, Colors.red.shade800),
      AggregateStatus.mixed => ('混合', Colors.grey.shade100, Colors.grey.shade600),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: fgColor,
        ),
      ),
    );
  }
}
