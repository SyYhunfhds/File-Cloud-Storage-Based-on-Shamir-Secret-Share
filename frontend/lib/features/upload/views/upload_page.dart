import 'dart:io' as io;
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:desktop_drop/desktop_drop.dart';
import 'package:go_router/go_router.dart';
import '../providers/upload_provider.dart';
import '../models/upload_models.dart';
import '../../../core/constants.dart';
import '../../../core/formatters.dart';

/// 上传工作台页面
///
/// 三阶段：文件选择 → 上传中 → 成功展示 Recovery Code
/// 支持桌面端文件拖拽、快捷键 (ESC/Ctrl+Z/Backspace) 返回主页。
class UploadPage extends ConsumerWidget {
  const UploadPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(uploadProvider);
    final colorScheme = Theme.of(context).colorScheme;

    // 快捷键绑定 + Focus 确保焦点不丢失到系统
    return CallbackShortcuts(
      bindings: {
        const SingleActivator(LogicalKeyboardKey.escape): () => context.go('/'),
        const SingleActivator(LogicalKeyboardKey.keyZ, control: true): () => context.go('/'),
        const SingleActivator(LogicalKeyboardKey.backspace): () => context.go('/'),
      },
      child: Focus(
        autofocus: true,
        child: Center(
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: AppConstants.contentMaxWidth),
              child: Padding(
                padding: const EdgeInsets.all(32),
                child: switch (state.phase) {
                  UploadPhase.idle || UploadPhase.fileSelected => _FileSelectView(colorScheme: colorScheme),
                  UploadPhase.uploading => _UploadingView(colorScheme: colorScheme),
                  UploadPhase.success => _SuccessView(colorScheme: colorScheme),
                  UploadPhase.error => _ErrorView(colorScheme: colorScheme),
                },
              ),
            ),
          ),
        ),
    );
  }
}

// =============================================================================
// 阶段1：文件选择 + 可选参数
// =============================================================================

class _FileSelectView extends ConsumerWidget {
  final ColorScheme colorScheme;
  const _FileSelectView({required this.colorScheme});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(uploadProvider);
    final hasFile = state.filePath != null && state.fileName != null;

    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        // 文件选择区域（支持拖拽）
        _buildFilePicker(context, ref, hasFile),
        const SizedBox(height: 32),

        // 可选参数
        _buildOptions(ref),
        const SizedBox(height: 24),

        // 上传按钮
        SizedBox(
          width: 200,
          height: 40,
          child: FilledButton.icon(
            onPressed: hasFile
                ? () => ref.read(uploadProvider.notifier).startUpload()
                : null,
            icon: const Icon(Icons.cloud_upload_outlined, size: 20),
            label: const Text('上传', style: TextStyle(fontSize: 15)),
          ),
        ),
        if (!hasFile)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
              '请先选择一个文件',
              style: TextStyle(fontSize: 12, color: colorScheme.onSurfaceVariant),
            ),
          ),
      ],
    );
  }

  Widget _buildFilePicker(BuildContext context, WidgetRef ref, bool hasFile) {
    final state = ref.watch(uploadProvider);
    return SizedBox(
      width: 480,
      height: 180,
      child: DropTarget(
        onDragEntered: (_) {
          debugPrint('[INFO] 拖拽进入上传区域');
        },
        onDragExited: (_) {
          debugPrint('[INFO] 拖拽离开上传区域');
        },
        onDragDone: (details) {
          if (details.files.isEmpty) return;
          _handleDroppedFile(details, ref);
        },
        child: MouseRegion(
          cursor: SystemMouseCursors.click,
          child: InkWell(
            onTap: () => ref.read(uploadProvider.notifier).selectFile(),
            borderRadius: BorderRadius.circular(12),
            splashColor: colorScheme.primary.withValues(alpha: 0.12),
            highlightColor: colorScheme.primary.withValues(alpha: 0.06),
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 200),
              width: 480,
              height: 180,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: hasFile
                      ? colorScheme.primary
                      : colorScheme.outlineVariant,
                  width: hasFile ? 2 : 1,
                  strokeAlign: BorderSide.strokeAlignInside,
                ),
                color: hasFile
                    ? colorScheme.primaryContainer.withValues(alpha: 0.15)
                    : colorScheme.surfaceContainerLow,
              ),
              child: Padding(
                padding: const EdgeInsets.all(20),
                child: SingleChildScrollView(
                  child: hasFile ? _buildFileInfo(ref, state) : _buildFilePrompt(),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  void _handleDroppedFile(DropDoneDetails details, WidgetRef ref) {
    final file = details.files.first;
    final filePath = file.path;
    final ioFile = io.File(filePath);
    final bytes = ioFile.lengthSync();
    final sizeMB = bytes / (1024 * 1024);
    final fileName = file.name;

    ref.read(uploadProvider.notifier).setDroppedFile(
      filePath: filePath,
      fileName: fileName,
      fileSizeMB: sizeMB,
      fileSizeBytes: bytes,
    );

    debugPrint('[INFO] 拖拽文件已选择: $fileName ($bytes B)');
  }

  Widget _buildFilePrompt() {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Icon(Icons.cloud_upload_outlined, size: 40, color: colorScheme.onSurfaceVariant),
        const SizedBox(height: 12),
        Text(
          '点击选择文件 或 拖拽文件到此处',
          style: TextStyle(fontSize: 15, color: colorScheme.onSurfaceVariant),
        ),
        const SizedBox(height: 6),
        Text(
          '支持任意格式，最大 8MB（超过 5MB 将提示）',
          style: TextStyle(fontSize: 12, color: colorScheme.outline),
        ),
      ],
    );
  }

  Widget _buildFileInfo(WidgetRef ref, UploadState state) {
    final bytes = state.fileSizeBytes ?? 0;
    final isOver5MB = bytes > 5 * 1024 * 1024;

    return Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.insert_drive_file, size: 36, color: colorScheme.primary),
          const SizedBox(height: 10),
          Text(
            state.fileName!,
            style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: colorScheme.primary),
            overflow: TextOverflow.ellipsis,
          ),
          const SizedBox(height: 4),
          Text(
            formatFileSize(bytes),
            style: TextStyle(
              fontSize: 13,
              color: isOver5MB ? colorScheme.error : colorScheme.onSurfaceVariant,
            ),
          ),
          if (isOver5MB) ...[
            const SizedBox(height: 4),
            Text(
              '文件较大，建议压缩后上传（后端限制 8MB）',
              style: TextStyle(fontSize: 11, color: colorScheme.error),
            ),
          ],
          const SizedBox(height: 8),
          TextButton.icon(
            onPressed: () => ref.read(uploadProvider.notifier).selectFile(),
            icon: const Icon(Icons.refresh, size: 14),
            label: const Text('重新选择', style: TextStyle(fontSize: 12)),
          ),
        ],
    );
  }

  Widget _buildOptions(WidgetRef ref) {
    final state = ref.watch(uploadProvider);

    return SizedBox(
      width: 400,
      child: Column(
        children: [
          // Recovery Code 长度（UI 预留，后端暂不支持）
          Row(
            children: [
              SizedBox(
                width: 160,
                child: Text(
                  'Recovery Code 长度',
                  style: TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
                ),
              ),
              Expanded(
                child: DropdownButtonFormField<int>(
                  initialValue: state.recoveryCodeLength,
                  isDense: true,
                  decoration: InputDecoration(
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  ),
                  style: const TextStyle(fontSize: 13),
                  items: const [
                    DropdownMenuItem(value: 16, child: Text('16')),
                    DropdownMenuItem(value: 32, child: Text('32')),
                    DropdownMenuItem(value: 48, child: Text('48')),
                    DropdownMenuItem(value: 64, child: Text('64')),
                  ],
                  onChanged: (v) {
                    if (v != null) ref.read(uploadProvider.notifier).setRecoveryCodeLength(v);
                  },
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),

          // 初始公开可见（UI 预留，后端暂不支持）
          Row(
            children: [
              SizedBox(
                width: 160,
                child: Text(
                  '初始公开可见',
                  style: TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
                ),
              ),
              Switch(
                value: state.isPublic,
                onChanged: (v) => ref.read(uploadProvider.notifier).setPublic(v),
              ),
              Text(
                state.isPublic ? '是' : '否',
                style: TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

// =============================================================================
// 阶段2：上传中
// =============================================================================

class _UploadingView extends ConsumerWidget {
  final ColorScheme colorScheme;
  const _UploadingView({required this.colorScheme});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(uploadProvider);
    // 确保 progress 在 [0.0, 1.0] 范围内，防止溢出
    final progress = state.uploadProgress.clamp(0.0, 1.0);

    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Icon(Icons.cloud_upload_outlined, size: 48, color: colorScheme.primary),
        const SizedBox(height: 20),
        Text(
          '正在上传: ${state.fileName ?? ''}',
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w500),
        ),
        const SizedBox(height: 20),
        // 使用 Flexible + ConstrainedBox 防止溢出
        SizedBox(
          width: 360,
          child: ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: progress,
              minHeight: 6,
            ),
          ),
        ),
        const SizedBox(height: 8),
        Text(
          '${(progress * 100).toInt()}%',
          style: TextStyle(fontSize: 14, color: colorScheme.primary),
        ),
        const SizedBox(height: 24),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.info_outline, size: 14, color: colorScheme.onSurfaceVariant),
            const SizedBox(width: 6),
            Text(
              '离开此页面不会中断上传',
              style: TextStyle(fontSize: 12, color: colorScheme.onSurfaceVariant),
            ),
          ],
        ),
      ],
    );
  }
}

// =============================================================================
// 阶段3：上传成功 — Recovery Code 展示
// =============================================================================

class _SuccessView extends ConsumerStatefulWidget {
  final ColorScheme colorScheme;
  const _SuccessView({required this.colorScheme});

  @override
  ConsumerState<_SuccessView> createState() => _SuccessViewState();
}

class _SuccessViewState extends ConsumerState<_SuccessView> {
  bool _showRecoveryCode = false;

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(uploadProvider);
    final result = state.result;
    final code = result?.recoveryCode ?? '';
    final maskedCode = _showRecoveryCode ? code : '●' * code.length;

    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Icon(Icons.check_circle_outline, size: 48, color: Colors.green),
        const SizedBox(height: 16),
        Text(
          '上传成功',
          style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: Colors.green),
        ),
        const SizedBox(height: 8),
        if (result?.name != null)
          Text(
            '文件名: ${result!.name}',
            style: TextStyle(fontSize: 14, color: widget.colorScheme.onSurfaceVariant),
          ),
        const SizedBox(height: 28),

        // Recovery Code 展示区
        Text(
          'Recovery Code（请妥善保存）:',
          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500, color: widget.colorScheme.onSurface),
        ),
        const SizedBox(height: 8),
        Container(
          width: 420,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(8),
            color: widget.colorScheme.surfaceContainerHighest,
            border: Border.all(color: widget.colorScheme.outlineVariant),
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
        const SizedBox(height: 12),

        // 操作按钮
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            TextButton.icon(
              onPressed: () => setState(() => _showRecoveryCode = !_showRecoveryCode),
              icon: Icon(_showRecoveryCode ? Icons.visibility_off : Icons.visibility, size: 16),
              label: Text(_showRecoveryCode ? '隐藏' : '显示', style: const TextStyle(fontSize: 13)),
            ),
            const SizedBox(width: 16),
            FilledButton.tonalIcon(
              onPressed: () {
                Clipboard.setData(ClipboardData(text: code));
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Recovery Code 已复制到剪贴板'),
                    duration: Duration(seconds: 2),
                  ),
                );
              },
              icon: const Icon(Icons.copy, size: 16),
              label: const Text('复制到剪贴板', style: TextStyle(fontSize: 13)),
              style: FilledButton.styleFrom(
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),

        // 警告
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.warning_amber_rounded, size: 16, color: widget.colorScheme.error),
            const SizedBox(width: 6),
            Text(
              '此恢复码仅显示一次，丢失后无法找回',
              style: TextStyle(fontSize: 12, color: widget.colorScheme.error),
            ),
          ],
        ),
        const SizedBox(height: 24),

        // 确定按钮
        SizedBox(
          width: 120,
          child: FilledButton(
            onPressed: () => ref.read(uploadProvider.notifier).reset(),
            child: const Text('确定', style: TextStyle(fontSize: 14)),
          ),
        ),
      ],
    );
  }
}

// =============================================================================
// 错误状态
// =============================================================================

class _ErrorView extends ConsumerWidget {
  final ColorScheme colorScheme;
  const _ErrorView({required this.colorScheme});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(uploadProvider);

    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        Icon(Icons.error_outline, size: 48, color: colorScheme.error),
        const SizedBox(height: 16),
        Text(
          '上传失败',
          style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: colorScheme.error),
        ),
        const SizedBox(height: 8),
        if (state.errorMessage != null)
          Text(
            state.errorMessage!,
            style: TextStyle(fontSize: 14, color: colorScheme.onSurfaceVariant),
          ),
        const SizedBox(height: 24),
        FilledButton.icon(
          onPressed: () => ref.read(uploadProvider.notifier).reset(),
          icon: const Icon(Icons.refresh, size: 18),
          label: const Text('重试', style: TextStyle(fontSize: 14)),
        ),
      ],
    );
  }
}
