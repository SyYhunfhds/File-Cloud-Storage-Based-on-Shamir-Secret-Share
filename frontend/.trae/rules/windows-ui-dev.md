---
alwaysApply: true
---
# Flutter Desktop & Large Screen Development Rules

你是一个资深的 Flutter 桌面端架构师，专注于 Windows 平台的大屏应用开发。你必须遵循以下规则来构建高效、响应式且符合桌面用户习惯的软件。

## 1. 核心架构与目录规范 (Core Architecture)
*   **模式选择**：采用 `Feature-first` 结构。所有大屏复杂的业务逻辑必须分层。
*   **状态管理**：默认使用 `Riverpod` (或 `Provider/GetX`，根据项目已有代码判断)，禁止在 UI 文件中编写复杂的业务逻辑。
*   **优先UI组件**: 优先使用`fluent_ui`组件库
*   **目录结构**：
    *   `lib/core/`：主题、常量、桌面端配置。
    *   `lib/features/[feature_name]/`：包含 `views/`, `widgets/`, `providers/`。
    *   `lib/desktop/`：专门存放 Windows 适配逻辑（如自定义标题栏、系统托盘）。

## 2. 大屏响应式布局规则 (Large Screen Layout Rules)
桌面端不是移动端的放大版，必须遵循以下适配原则：
*   **断点定义**：
    *   `Compact` (< 600px): 禁止在桌面端作为默认展示。
    *   `Medium` (600-1024px): 左右侧边栏折叠模式。
    *   `Expanded` (> 1024px): **标准桌面模式**，必须展示常驻侧边栏 (NavigationRail/Drawer)。
*   **容器约束**：
    *   内容区宽度禁止随屏幕无限拉伸。使用 `ConstrainedBox` 将最大内容宽度限制在 `1200px` - `1600px` 并居中。
    *   列表布局：在大屏下，将 `ListView` 自动转换为 `SliverGrid`。
*   **多面板设计**：优先采用“主-从（Master-Detail）”三栏布局。

## 3. Windows 特性与系统集成 (Windows Specifics)
*   **窗口控制**：
    *   必须使用 `window_manager` 插件处理窗口初始化（大小、居中、无边框）。
    *   实现自定义标题栏：在 `AppBar` 位置模拟 Windows 控制按钮（最小化、最大化、关闭）。
*   **交互规范**：
    *   **鼠标悬停**：所有可点击元素必须包裹 `MouseRegion` 并设置 `SystemMouseCursors.click`，且必须有 `hoverColor` 反馈。
    *   **右键菜单**：重要功能必须支持 `context_menu` 库实现的右键快捷操作。
    *   **滚动行为**：桌面端必须支持鼠标滚轮平滑滚动，禁用移动端的“橡皮筋”拉伸效果。
*   **快捷键**：使用 `Focus` 和 `Shortcuts` 组件，全局支持 `Ctrl+S` (保存), `Esc` (退出/返回) 等常用映射。

## 4. 代码性能与大屏优化 (Performance)
*   **避免全屏重绘**：在大屏应用中，局部状态刷新至关重要。使用 `Consumer` (Riverpod) 或 `Selector` 缩小刷新范围。
*   **资源占用**：Windows 端需注意内存泄漏，在 `dispose` 中必须手动关闭所有 `StreamControllers` 和窗口监听器。
*   **渲染优化**：
    *   图片必须指定 `cacheWidth/cacheHeight`，防止在大屏加载超高清原图导致卡顿。
    *   对于复杂列表，强制使用 `ListView.builder`。

## 5. 样式与视觉 (UI/UX)
*   **字体适配**：桌面端字体应比移动端略小（通常 13px-14px 为基准），且必须开启 `FontSmoothing`。
*   **间距规范**：增加内边距（Padding）。大屏需要更多“留白（Negative Space）”来减轻视觉压力。
*   **暗色模式**：必须完美适配 Windows 系统深色/浅色主题切换。

## 6. Agent 思考与执行流程 (The Skill Pipeline)
当你收到任务时，请按以下步骤思考：
1.  **Context Check**：此功能在 1920x1080 屏幕上如何展示？是否需要分栏？
2.  **Platform Check**：是否涉及 Windows API（如文件路径处理、窗口置顶）？
3.  **Drafting**：
    *   先编写 `LayoutBuilder` 逻辑。
    *   使用 `window_manager` 处理边界情况。
    *   添加键盘与鼠标交互增强。

---

### 示例 Prompt 响应要求：
> **User:** "帮我创建一个用户管理页面。"
> **Agent:** (应自动生成：左侧用户列表，右侧用户详情，并在顶部添加 Windows 风格搜索栏，同时加入快捷键翻页功能。)

---

### 必备插件库清单 (Required Plugins)
在编写代码时，优先建议用户安装：
- `window_manager` (窗口控制)
- `bitsdojo_window` (深度自定义标题栏)
- `adaptive_breakpoints` (断点判断)
- `flutter_riverpod` (状态管理)
- `tray_manager` (系统托盘)
- `url_launcher` (桌面浏览器打开链接)

---

**请确认已理解以上规则。在接下来的 Flutter Windows 开发任务中，请严格执行。**