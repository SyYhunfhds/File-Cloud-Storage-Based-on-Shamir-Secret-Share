import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../shares/providers/share_refresh_provider.dart';

/// 对话框阶段
enum _Phase { input, progress, result }

/// 强制刷新份额对话框
///
/// 三阶段：输入 Recovery Code → SSE 进度 → 结果展示
/// 由条目列表的操作按钮触发。
class ForceRefreshShareDialog extends ConsumerStatefulWidget {
  final int itemId;
  final String itemName;
  final String deviceShare;

  const ForceRefreshShareDialog({
    super.key,
    required this.itemId,
    required this.itemName,
    required this.deviceShare,
  });

  /// 打开强制刷新份额对话框
  static Future<void> show(
    BuildContext context, {
    required int itemId,
    required String itemName,
    required String deviceShare,
  }) {
    return showDialog(
      context: context,
      barrierDismissible: false,
      builder: (_) => ForceRefreshShareDialog(
        itemId: itemId,
        itemName: itemName,
        deviceShare: deviceShare,
      ),
    );
  }

  @override
  ConsumerState<ForceRefreshShareDialog> createState() =>
      _ForceRefreshShareDialogState();
}

class _ForceRefreshShareDialogState
    extends ConsumerState<ForceRefreshShareDialog> {
  _Phase _phase = _Phase.input;
  final _recoveryController = TextEditingController();
  final _scrollController = ScrollController();
  bool _showRecoveryCode = false;

  @override
  void dispose() {
    _recoveryController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _startRefresh() {
    setState(() => _phase = _Phase.progress);

    final recoveryCode = _recoveryController.text.trim();
    ref.read(shareRefreshProvider.notifier).startRefresh(
          itemId: widget.itemId,
          deviceShare: widget.deviceShare,
          recoveryCode: recoveryCode.isEmpty ? null : recoveryCode,
        );
  }

  void _cancelRefresh() {
    ref.read(shareRefreshProvider.notifier).cancel();
    Navigator.of(context).pop();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(shareRefreshProvider);
    final colorScheme = Theme.of(context).colorScheme;

    // SSE 完成后切换到结果阶段
    if (_phase == _Phase.progress && state.result != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _phase == _Phase.progress) {
          setState(() => _phase = _Phase.result);
        }
      });
    }

    return AlertDialog(
      title: Text(_titleText(state)),
      content: SizedBox(
        width: 420,
        child: switch (_phase) {
          _Phase.input => _buildInputPhase(colorScheme),
          _Phase.progress => _buildProgressPhase(colorScheme, state),
          _Phase.result => _buildResultPhase(colorScheme, state),
        },
      ),
      actions: _buildActions(colorScheme, state),
    );
  }

  String _titleText(ShareRefreshState state) {
    return switch (_phase) {
      _Phase.input => '强制刷新份额',
      _Phase.progress =>
        state.errorMessage != null ? '刷新失败' : '正在刷新份额',
      _Phase.result => '份额刷新完成',
    };
  }

  // ===========================================================================
  // 阶段一：输入 Recovery Code
  // ===========================================================================

  Widget _buildInputPhase(ColorScheme colorScheme) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 条目信息
        Text(
          '条目 #${widget.itemId}: ${widget.itemName}',
          style: Theme.of(context)
              .textTheme
              .bodyMedium
              ?.copyWith(fontWeight: FontWeight.w500),
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
        ),
        const SizedBox(height: 8),
        Row(
          children: [
            Icon(Icons.check_circle, size: 14, color: Colors.green),
            const SizedBox(width: 6),
            Text(
              'Device Share 已就绪',
              style: TextStyle(
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
        const SizedBox(height: 20),

        // Recovery Code 输入
        TextField(
          controller: _recoveryController,
          decoration: InputDecoration(
            labelText: 'Recovery Code（可选）',
            hintText: '如更换设备时输入，否则留空',
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
            ),
            isDense: true,
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
          ),
          style: const TextStyle(fontSize: 14, fontFamily: 'monospace'),
        ),
      ],
    );
  }

  // ===========================================================================
  // 阶段二：SSE 进度
  // ===========================================================================

  Widget _buildProgressPhase(ColorScheme colorScheme, ShareRefreshState state) {
    return SizedBox(
      height: 300,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 条目信息
          Text(
            '条目 #${widget.itemId}: ${widget.itemName}',
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(fontWeight: FontWeight.w500),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 16),

          // 进度条
          if (state.isRefreshing || state.result != null) ...[
            LinearProgressIndicator(
              value: state.progress / 100.0,
              minHeight: 6,
              borderRadius: BorderRadius.circular(3),
            ),
            const SizedBox(height: 4),
            Text(
              '${state.progress}% — ${state.currentMessage}',
              style: TextStyle(
                fontSize: 12,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 8),
          ],

          // 错误信息
          if (state.errorMessage != null)
            Container(
              padding: const EdgeInsets.all(8),
              margin: const EdgeInsets.only(bottom: 8),
              decoration: BoxDecoration(
                color: colorScheme.errorContainer.withValues(alpha: 0.3),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                state.errorMessage!,
                style: TextStyle(
                  fontSize: 12,
                  color: colorScheme.error,
                ),
              ),
            ),

          // 日志区域
          const Text('处理日志',
              style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600)),
          const SizedBox(height: 4),
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: colorScheme.surfaceContainerHighest
                    .withValues(alpha: 0.4),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(
                  color: colorScheme.outlineVariant.withValues(alpha: 0.4),
                ),
              ),
              padding: const EdgeInsets.all(8),
              child: ListView.builder(
                controller: _scrollController,
                itemCount: state.logLines.length,
                itemBuilder: (context, index) {
                  final isLast = index == state.logLines.length - 1;
                  return Text(
                    state.logLines[index],
                    style: TextStyle(
                      fontSize: 11,
                      fontFamily: 'monospace',
                      color: isLast
                          ? colorScheme.onSurface
                          : colorScheme.onSurfaceVariant
                              .withValues(alpha: 0.7),
                    ),
                  );
                },
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ===========================================================================
  // 阶段三：结果展示
  // ===========================================================================

  Widget _buildResultPhase(ColorScheme colorScheme, ShareRefreshState state) {
    final result = state.result!;
    final code = result.recoveryCode;
    final maskedCode = _showRecoveryCode ? code : '●' * code.length;

    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Icon(Icons.check_circle_outline, size: 40, color: Colors.green),
        const SizedBox(height: 12),
        Text(
          '份额刷新成功',
          style: TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            color: Colors.green,
          ),
        ),
        const SizedBox(height: 20),

        // Recovery Code 是否重新生成
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
          decoration: BoxDecoration(
            color: colorScheme.surfaceContainerHighest,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'Recovery Code 状态: ',
                style: TextStyle(
                  fontSize: 13,
                  color: colorScheme.onSurfaceVariant,
                ),
              ),
              Icon(
                result.isRecoveryCodeReGenerated
                    ? Icons.refresh
                    : Icons.check,
                size: 16,
                color: result.isRecoveryCodeReGenerated
                    ? colorScheme.primary
                    : Colors.green,
              ),
              const SizedBox(width: 4),
              Text(
                result.isRecoveryCodeReGenerated ? '已重新生成' : '未变更',
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: result.isRecoveryCodeReGenerated
                      ? colorScheme.primary
                      : Colors.green,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 20),

        // Recovery Code 掩码展示
        Text(
          'Recovery Code:',
          style: TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w500,
            color: colorScheme.onSurface,
          ),
        ),
        const SizedBox(height: 8),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            color: colorScheme.surfaceContainerHighest,
            border: Border.all(color: colorScheme.outlineVariant),
          ),
          child: SelectableText(
            maskedCode,
            style: const TextStyle(
              fontSize: 16,
              fontFamily: 'monospace',
              letterSpacing: 2,
            ),
            textAlign: TextAlign.center,
          ),
        ),
        const SizedBox(height: 8),

        // 显示/隐藏切换
        TextButton.icon(
          onPressed: () =>
              setState(() => _showRecoveryCode = !_showRecoveryCode),
          icon: Icon(
            _showRecoveryCode ? Icons.visibility_off : Icons.visibility,
            size: 16,
          ),
          label: Text(
            _showRecoveryCode ? '隐藏' : '显示',
            style: const TextStyle(fontSize: 13),
          ),
        ),
      ],
    );
  }

  // ===========================================================================
  // 操作按钮
  // ===========================================================================

  List<Widget> _buildActions(ColorScheme colorScheme, ShareRefreshState state) {
    return switch (_phase) {
      _Phase.input => [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('取消'),
          ),
          FilledButton(
            onPressed: _startRefresh,
            child: const Text('开始刷新'),
          ),
        ],
      _Phase.progress => [
          if (state.errorMessage != null) ...[
            FilledButton(
              onPressed: _startRefresh,
              child: const Text('重试'),
            ),
            const SizedBox(width: 8),
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('关闭'),
            ),
          ] else ...[
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('后台运行'),
            ),
            if (state.isRefreshing)
              FilledButton(
                onPressed: _cancelRefresh,
                child: const Text('取消'),
              ),
          ],
        ],
      _Phase.result => [
          FilledButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('关闭'),
          ),
        ],
    };
  }
}
