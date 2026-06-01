import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../auth/providers/auth_provider.dart';
import '../../auth/widgets/login_dialog.dart';
import '../providers/entry_provider.dart';

/// 主页顶部工具栏
///
/// 固定高度 56（上下各 8 padding，内部内容高度 40）。包含搜索框、用户信息区。
class HomeToolbar extends ConsumerWidget {
  static const double _toolbarHeight = 56;
  static const double _childHeight = 40;

  final TextEditingController searchController;
  final ValueChanged<String> onSearchChanged;

  const HomeToolbar({
    super.key,
    required this.searchController,
    required this.onSearchChanged,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final colorScheme = Theme.of(context).colorScheme;

    return Container(
      height: _toolbarHeight,
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: colorScheme.surface,
        border: Border(
          bottom: BorderSide(color: colorScheme.outlineVariant.withValues(alpha: 0.5)),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          // 搜索框
          SizedBox(
            width: 260,
            height: _childHeight,
            child: TextField(
              controller: searchController,
              onChanged: onSearchChanged,
              style: const TextStyle(fontSize: 13),
              decoration: InputDecoration(
                hintText: '搜索文件名…',
                hintStyle: TextStyle(fontSize: 13, color: colorScheme.onSurfaceVariant),
                prefixIcon: Icon(Icons.search, size: 18, color: colorScheme.onSurfaceVariant),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: colorScheme.outlineVariant),
                ),
                contentPadding: const EdgeInsets.symmetric(vertical: 0, horizontal: 12),
                isDense: true,
              ),
            ),
          ),
          // 刷新按钮
          IconButton(
            icon: const Icon(Icons.refresh, size: 20),
            tooltip: '刷新条目列表',
            onPressed: () {
              final mode = ref.read(entryListProvider).filterMode;
              switch (mode) {
                case EntryFilterMode.my:
                  ref.read(entryListProvider.notifier).fetchMyEntries();
                case EntryFilterMode.public:
                  ref.read(entryListProvider.notifier).fetchPublicEntries();
                case EntryFilterMode.all:
                  ref.read(entryListProvider.notifier).fetchAllEntries();
              }
            },
          ),
          const Spacer(),

          // 用户信息区
          _buildUserArea(context, ref, authState),
        ],
      ),
    );
  }

  Widget _buildUserArea(BuildContext context, WidgetRef ref, AuthState authState) {
    final colorScheme = Theme.of(context).colorScheme;

    if (authState.isLoggedIn) {
      final registeredDate = authState.registeredAt != null
          ? '${authState.registeredAt!.year}-'
            '${authState.registeredAt!.month.toString().padLeft(2, '0')}-'
            '${authState.registeredAt!.day.toString().padLeft(2, '0')}'
          : '';
      final tooltipMessage = authState.email.isNotEmpty
          ? '${authState.userName} (${authState.email})\n'
            '${authState.userPosition} · 权限等级 ${authState.privilege}\n'
            '注册于: $registeredDate'
          : '${authState.userName} — ${authState.userPosition}';

      return MouseRegion(
        cursor: SystemMouseCursors.click,
        child: Tooltip(
          message: tooltipMessage,
          child: InkWell(
            onTap: () {},
            borderRadius: BorderRadius.circular(8),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  CircleAvatar(
                    radius: 14,
                    backgroundColor: colorScheme.primary,
                    child: Text(
                      authState.userName.isNotEmpty
                          ? authState.userName.characters.first
                          : '',
                      style: TextStyle(
                        color: colorScheme.onPrimary,
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        authState.userName,
                        style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
                      ),
                      Text(
                        authState.userPosition,
                        style: TextStyle(
                          fontSize: 11,
                          color: colorScheme.onSurfaceVariant,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(width: 8),
                  MouseRegion(
                    cursor: SystemMouseCursors.click,
                    child: Tooltip(
                      message: '登出',
                      child: InkWell(
                        onTap: () async {
                          await ref.read(authProvider.notifier).logout();
                        },
                        borderRadius: BorderRadius.circular(8),
                        child: Padding(
                          padding: const EdgeInsets.all(4),
                          child: Icon(
                            Icons.logout,
                            size: 18,
                            color: colorScheme.onSurfaceVariant,
                          ),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
    } else {
      return Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          MouseRegion(
            cursor: SystemMouseCursors.click,
            child: Tooltip(
              message: '点击登录',
              child: InkWell(
                onTap: () => _showLoginDialog(context),
                borderRadius: BorderRadius.circular(8),
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  child: CircleAvatar(
                    radius: 14,
                    backgroundColor: Colors.grey.shade400,
                    child: Icon(
                      Icons.person,
                      size: 16,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            height: _childHeight,
            child: FilledButton.tonal(
              onPressed: () => _showLoginDialog(context),
              style: FilledButton.styleFrom(
                minimumSize: const Size(0, _childHeight),
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                padding: const EdgeInsets.symmetric(horizontal: 14),
              ),
              child: const Text('登录', style: TextStyle(fontSize: 13)),
            ),
          ),
        ],
      );
    }
  }

  void _showLoginDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (_) => const LoginDialog(),
    );
  }
}
