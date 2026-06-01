import 'package:flutter/material.dart';
import '../../shares/models/share_record_data.dart';

/// 份额列表表格组件
///
/// 表头与数据行分离，切换页面时仅数据行区域通过 AnimatedSwitcher 重建。
class ShareList extends StatelessWidget {
  final List<ShareRecordData> shares;
  final int pageIndex;
  final int? selectedItemId;
  final bool isLoading;
  final ValueChanged<int> onSelect;
  final ValueChanged<int> onDelete;

  const ShareList({
    super.key,
    required this.shares,
    required this.pageIndex,
    required this.selectedItemId,
    required this.isLoading,
    required this.onSelect,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;

    return Column(
      children: [
        // === 表头 ===
        _buildHeader(colorScheme, textTheme),
        const Divider(height: 1),
        // === 数据行 ===
        Expanded(
          child: isLoading
              ? const Center(child: CircularProgressIndicator())
              : AnimatedSwitcher(
                  duration: const Duration(milliseconds: 200),
                  child: ListView.builder(
                    key: ValueKey('share_page_$pageIndex'),
                    itemCount: shares.length,
                    itemBuilder: (context, index) {
                      final record = shares[index];
                      final globalIndex = pageIndex + index + 1;
                      final isSelected = selectedItemId == record.itemId;
                      return _buildRow(
                        context, record, globalIndex, isSelected, colorScheme);
                    },
                  ),
                ),
        ),
      ],
    );
  }

  // ===========================================================================
  // 表头
  // ===========================================================================

  Widget _buildHeader(ColorScheme colorScheme, TextTheme textTheme) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      color: colorScheme.surfaceContainerHighest.withValues(alpha: 0.5),
      child: Row(
        children: [
          SizedBox(
            width: 60,
            child: Text('#', style: _headerStyle(textTheme, colorScheme)),
          ),
          Expanded(
            flex: 2,
            child:
                Text('原始文件名', style: _headerStyle(textTheme, colorScheme)),
          ),
          Expanded(
            flex: 3,
            child: Text('掩码份额 (Base64)', style: _headerStyle(textTheme, colorScheme)),
          ),
          SizedBox(
            width: 170,
            child:
                Text('保存时间', style: _headerStyle(textTheme, colorScheme)),
          ),
          SizedBox(
            width: 130,
            child: Text('操作', style: _headerStyle(textTheme, colorScheme)),
          ),
        ],
      ),
    );
  }

  TextStyle _headerStyle(TextTheme textTheme, ColorScheme colorScheme) {
    return (textTheme.labelMedium ?? const TextStyle()).copyWith(
      color: colorScheme.onSurfaceVariant,
      fontWeight: FontWeight.w600,
      fontSize: 12,
    );
  }

  // ===========================================================================
  // 数据行
  // ===========================================================================

  Widget _buildRow(BuildContext context, ShareRecordData record, int index,
      bool isSelected, ColorScheme colorScheme) {
    final textTheme = Theme.of(context).textTheme;
    final bodyStyle = textTheme.bodyMedium?.copyWith(fontSize: 13);

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: InkWell(
        onTap: () => onSelect(record.itemId),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          decoration: BoxDecoration(
            color: isSelected
                ? colorScheme.primaryContainer.withValues(alpha: 0.3)
                : null,
            border: Border(
              bottom: BorderSide(
                color: colorScheme.outlineVariant.withValues(alpha: 0.3),
              ),
            ),
          ),
          child: Row(
            children: [
              SizedBox(
                width: 60,
                child: Text('$index', style: bodyStyle),
              ),
              Expanded(
                flex: 2,
                child: Text(
                  record.originalFilename,
                  style: bodyStyle,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Expanded(
                flex: 3,
                child: Text(
                  _maskShare(record.encryptedShare),
                  style: bodyStyle?.copyWith(
                    fontFamily: 'monospace',
                    fontSize: 12,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
              SizedBox(
                width: 170,
                child: Text(
                  _formatDateTime(record.createdAt),
                  style: bodyStyle?.copyWith(
                    fontSize: 12,
                    color: colorScheme.onSurfaceVariant,
                  ),
                ),
              ),
              SizedBox(
                width: 130,
                child: Row(
                  children: [
                    _smallButton(context, '详情', () => onSelect(record.itemId)),
                    const SizedBox(width: 4),
                    _smallButton(
                      context,
                      '删除',
                      () => onDelete(record.itemId),
                      isDanger: true,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _smallButton(
    BuildContext context,
    String label,
    VoidCallback onPressed, {
    bool isDanger = false,
  }) {
    final colorScheme = Theme.of(context).colorScheme;
    return SizedBox(
      height: 28,
      child: TextButton(
        onPressed: onPressed,
        style: TextButton.styleFrom(
          padding: const EdgeInsets.symmetric(horizontal: 8),
          tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          minimumSize: Size.zero,
          foregroundColor:
              isDanger ? colorScheme.error : colorScheme.primary,
        ),
        child: Text(label, style: const TextStyle(fontSize: 12)),
      ),
    );
  }

  // ===========================================================================
  // 工具方法
  // ===========================================================================

  /// 掩码份额：前8字符 + **** + 后8字符
  String _maskShare(String encrypted) {
    if (encrypted.length <= 16) return '****';
    return '${encrypted.substring(0, 8)}****${encrypted.substring(encrypted.length - 8)}';
  }

  String _formatDateTime(DateTime dt) {
    return '${dt.year}-${_pad(dt.month)}-${_pad(dt.day)} '
        '${_pad(dt.hour)}:${_pad(dt.minute)}';
  }

  String _pad(int n) => n.toString().padLeft(2, '0');
}
