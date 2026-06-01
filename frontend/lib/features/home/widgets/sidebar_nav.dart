import 'package:flutter/material.dart';
import '../../../core/constants.dart';

/// NavigationRail 侧边导航栏
///
/// 内置宽度与展开态判断（依据屏幕整体宽度），便于在 AppShell 的 Row 中直接使用。
class AppSidebarNav extends StatelessWidget {
  final int selectedIndex;
  final ValueChanged<int> onDestinationSelected;

  const AppSidebarNav({
    super.key,
    required this.selectedIndex,
    required this.onDestinationSelected,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final screenWidth = MediaQuery.sizeOf(context).width;
    final isExtended = screenWidth >= AppConstants.mediumBreakpoint;
    final width = isExtended ? 220.0 : 80.0;

    return SizedBox(
      width: width,
      child: NavigationRail(
        selectedIndex: selectedIndex,
        onDestinationSelected: onDestinationSelected,
        extended: isExtended,
        backgroundColor: colorScheme.surfaceContainerHighest,
        labelType: isExtended
            ? NavigationRailLabelType.none
            : NavigationRailLabelType.selected,
        destinations: const [
          NavigationRailDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home),
            label: Text('主页'),
          ),
          NavigationRailDestination(
            icon: Icon(Icons.upload_file_outlined),
            selectedIcon: Icon(Icons.upload_file),
            label: Text('上传条目'),
          ),
          NavigationRailDestination(
              icon: Icon(Icons.vpn_key_outlined),
              selectedIcon: Icon(Icons.vpn_key),
              label: Text('份额管理'),
            ),
          NavigationRailDestination(
            icon: Icon(Icons.settings_outlined),
            selectedIcon: Icon(Icons.settings),
            label: Text('设置'),
          ),
        ],
      ),
    );
  }
}
