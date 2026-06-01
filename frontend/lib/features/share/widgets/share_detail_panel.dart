import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../shares/models/share_record_data.dart';
import '../../shares/services/share_service.dart';

/// 份额详情面板（大屏右侧 / 小屏底部 Sheet）
///
/// 显示非敏感字段 + Debug 模式下提供"复制解密份额"按钮。
class ShareDetailPanel extends StatefulWidget {
  final ShareRecordData record;
  final ShareService shareService;
  final VoidCallback onClose;

  const ShareDetailPanel({
    super.key,
    required this.record,
    required this.shareService,
    required this.onClose,
  });

  @override
  State<ShareDetailPanel> createState() => _ShareDetailPanelState();
}

class _ShareDetailPanelState extends State<ShareDetailPanel> {
  bool _isDecrypting = false;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final textTheme = Theme.of(context).textTheme;
    final r = widget.record;

    return Container(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 标题栏
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('份额详情', style: textTheme.titleMedium),
              IconButton(
                onPressed: widget.onClose,
                icon: const Icon(Icons.close, size: 20),
                tooltip: '关闭',
                visualDensity: VisualDensity.compact,
              ),
            ],
          ),
          const SizedBox(height: 16),
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _infoRow('条目 ID', '${r.itemId}', textTheme, colorScheme),
                  _infoRow(
                      '原始文件名', r.originalFilename, textTheme, colorScheme),
                  _infoRow(
                      '服务端文件名', r.serverFilename, textTheme, colorScheme),
                  _infoRow('保存时间',
                      _formatDateTime(r.createdAt), textTheme, colorScheme),
                  const SizedBox(height: 16),
                  // 掩码份额（不可关闭遮蔽）
                  Text(
                    '掩码份额 (Base64)',
                    style: textTheme.labelMedium?.copyWith(
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color:
                          colorScheme.surfaceContainerHighest.withValues(alpha: 0.4),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: SelectableText(
                      _maskShare(r.encryptedShare),
                      style: textTheme.bodySmall?.copyWith(
                        fontFamily: 'monospace',
                        fontSize: 12,
                      ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  // Debug 模式：复制解密份额
                  if (kDebugMode)
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        onPressed: _isDecrypting ? null : _copyDecryptedShare,
                        icon: _isDecrypting
                            ? const SizedBox(
                                width: 16,
                                height: 16,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Icon(Icons.copy, size: 18),
                        label: Text(
                            _isDecrypting ? '解密中…' : '复制解密份额 (Debug)'),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: colorScheme.tertiaryContainer,
                          foregroundColor: colorScheme.onTertiaryContainer,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ===========================================================================
  // 复制解密份额
  // ===========================================================================

  Future<void> _copyDecryptedShare() async {
    setState(() => _isDecrypting = true);
    try {
      final plaintext =
          await widget.shareService.getDecryptedShare(widget.record.itemId);
      if (plaintext != null && mounted) {
        await Clipboard.setData(ClipboardData(text: plaintext));
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('已复制解密份额到剪贴板')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('解密失败: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isDecrypting = false);
      }
    }
  }

  // ===========================================================================
  // 工具方法
  // ===========================================================================

  Widget _infoRow(
      String label, String value, TextTheme textTheme, ColorScheme colorScheme) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: textTheme.labelMedium?.copyWith(
              color: colorScheme.onSurfaceVariant,
              fontSize: 12,
            ),
          ),
          const SizedBox(height: 2),
          SelectableText(
            value,
            style: textTheme.bodyMedium?.copyWith(fontSize: 13),
          ),
        ],
      ),
    );
  }

  String _maskShare(String encrypted) {
    if (encrypted.length <= 16) return '****';
    return '${encrypted.substring(0, 8)}****${encrypted.substring(encrypted.length - 8)}';
  }

  String _formatDateTime(DateTime dt) {
    return '${dt.year}-${_pad(dt.month)}-${_pad(dt.day)} '
        '${_pad(dt.hour)}:${_pad(dt.minute)}:${_pad(dt.second)}';
  }

  String _pad(int n) => n.toString().padLeft(2, '0');
}
