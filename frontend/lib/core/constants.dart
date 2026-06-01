/// 应用常量定义
class AppConstants {
  /// 窗口最小宽度
  static const double windowMinWidth = 1024;

  /// 窗口最小高度
  static const double windowMinHeight = 700;

  /// 内容区最大宽度（大屏居中约束）
  static const double contentMaxWidth = 1400;

  /// Compact 断点（< 640px 仅图标侧边栏，通常移动端 / 小屏）
  static const double compactBreakpoint = 640;

  /// Medium 断点（>= 1000px 桌面端默认展开侧边栏，图标 + 文本标签）
  static const double mediumBreakpoint = 1000;
}
