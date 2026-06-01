import 'package:flutter/material.dart';
import 'package:window_manager/window_manager.dart';

/// 自定义 Windows 标题栏（Material 风格）
///
/// 仅显示"应用图标 + 名称"，窗口控制完全交给 Windows 原生标题栏，
/// 因此这里不再渲染最小化 / 最大化 / 关闭按钮。整行区域仍然可拖拽。
class AppTitleBar extends StatelessWidget implements PreferredSizeWidget {
  const AppTitleBar({super.key});

  @override
  Size get preferredSize => const Size.fromHeight(32);

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return GestureDetector(
      onPanStart: (_) => windowManager.startDragging(),
      child: Container(
        height: 32,
        color: colorScheme.surfaceContainerHighest,
        padding: const EdgeInsets.symmetric(horizontal: 8),
        child: Row(
          children: [
            Icon(Icons.account_balance, size: 18, color: colorScheme.primary),
            const SizedBox(width: 8),
            Text(
              '财务条目托管终端',
              style: TextStyle(
                fontSize: 13,
                color: colorScheme.onSurface,
              ),
            ),
            const Spacer(),
          ],
        ),
      ),
    );
  }
}
