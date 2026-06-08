import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../shares/providers/share_refresh_provider.dart';

/// 份额刷新进度对话框
///
/// 显示 SSE 进度条、阶段日志和取消按钮。
/// progress=100 时自动关闭。
class ShareRefreshDialog extends ConsumerStatefulWidget {
  final int itemId;
  final String itemName;
  final String? deviceShare;
  final String? recoveryCode;

  const ShareRefreshDialog({
    super.key,
    required this.itemId,
    required this.itemName,
    this.deviceShare,
    this.recoveryCode,
  });

  /// 打开份额刷新对话框
  static Future<void> show(
    BuildContext context, {
    required int itemId,
    required String itemName,
    String? deviceShare,
    String? recoveryCode,
  }) {
    return showDialog(
      context: context,
      barrierDismissible: false,
      builder: (_) => ShareRefreshDialog(
        itemId: itemId,
        itemName: itemName,
        deviceShare: deviceShare,
        recoveryCode: recoveryCode,
      ),
    );
  }

  @override
  ConsumerState<ShareRefreshDialog> createState() =>
      _ShareRefreshDialogState();
}

class _ShareRefreshDialogState extends ConsumerState<ShareRefreshDialog> {
  final _scrollController = ScrollController();
  bool _started = false;

  @override
  void initState() {
    super.initState();
    // 延迟一帧启动，确保 Provider 已挂载
    Future.microtask(() {
      if (!_started) {
        _started = true;
        ref.read(shareRefreshProvider.notifier).startRefresh(
              itemId: widget.itemId,
              deviceShare: widget.deviceShare,
              recoveryCode: widget.recoveryCode,
            );
      }
    });
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(shareRefreshProvider);

    // 完成 → 关闭对话框
    if (state.result != null) {
      _scheduleClose(context);
    }

    return AlertDialog(
      title: Text(state.result != null ? '份额刷新完成' : '正在刷新份额'),
      content: SizedBox(
        width: 420,
        height: 320,
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
                value: (state.result != null ? 100 : state.progress) / 100.0,
                minHeight: 6,
                borderRadius: BorderRadius.circular(3),
              ),
              const SizedBox(height: 4),
              Text(
                state.result != null
                    ? '100% — 完成'
                    : '${state.progress}% — ${state.currentMessage}',
                style: TextStyle(
                  fontSize: 12,
                  color: Theme.of(context).colorScheme.onSurfaceVariant,
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
                  color: Theme.of(context)
                      .colorScheme
                      .errorContainer
                      .withValues(alpha: 0.3),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  state.errorMessage!,
                  style: TextStyle(
                    fontSize: 12,
                    color: Theme.of(context).colorScheme.error,
                  ),
                ),
              ),

            // 日志区域
            const Text('处理日志', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600)),
            const SizedBox(height: 4),
            Expanded(
              child: Container(
                decoration: BoxDecoration(
                  color: Theme.of(context)
                      .colorScheme
                      .surfaceContainerHighest
                      .withValues(alpha: 0.4),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: Theme.of(context)
                        .colorScheme
                        .outlineVariant
                        .withValues(alpha: 0.4),
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
                            ? Theme.of(context).colorScheme.onSurface
                            : Theme.of(context)
                                .colorScheme
                                .onSurfaceVariant
                                .withValues(alpha: 0.7),
                      ),
                    );
                  },
                ),
              ),
            ),
          ],
        ),
      ),
      actions: [
        if (state.result != null)
          FilledButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('关闭'),
          )
        else ...[
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('后台运行'),
          ),
          if (state.isRefreshing)
            FilledButton(
              onPressed: () {
                ref.read(shareRefreshProvider.notifier).cancel();
                Navigator.of(context).pop();
              },
              child: const Text('取消'),
            ),
          if (state.errorMessage != null)
            FilledButton(
              onPressed: () {
                ref.read(shareRefreshProvider.notifier).startRefresh(
                      itemId: widget.itemId,
                      deviceShare: widget.deviceShare,
                      recoveryCode: widget.recoveryCode,
                    );
              },
              child: const Text('重试'),
            ),
        ],
      ],
    );
  }

  bool _closing = false;

  void _scheduleClose(BuildContext context) {
    if (_closing) return;
    _closing = true;
    Future.delayed(const Duration(milliseconds: 500), () {
      if (context.mounted) {
        Navigator.of(context).pop();
      }
    });
  }
}