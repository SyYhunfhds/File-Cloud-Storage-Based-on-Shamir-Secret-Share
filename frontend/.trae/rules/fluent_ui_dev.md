---
alwaysApply: false
---
这是一份专门为 **Flutter for Windows (Fluent UI 风格)** 深度定制的 Agent 规则文档。你可以将其保存为项目根目录下的 `.cursorrules`、`CLAUDE.md` 或作为 Agent 的系统提示词（System Prompt）。

---

# Flutter Windows (Fluent UI) 开发规范与 Agent 指令集

## 1. 项目核心定位
*   **目标平台**：仅限 Windows 桌面端（大屏优化）。
*   **视觉语言**：遵循微软 **Fluent Design System** (Windows 11 审美)。
*   **核心库**：使用 `fluent_ui` 替代 `material`。

## 2. 核心技术栈规范
*   **UI 框架**：`fluent_ui: ^latest`
*   **窗口管理**：`window_manager`（处理窗口置顶、无边框、自定义标题栏）。
*   **状态管理**：`Riverpod`，禁止在 UI 文件中编写复杂的业务逻辑。
*   **图标库**：优先使用 `fluentui_system_icons`。

---

## 3. UI 布局规则 (Layout Rules)

### 3.1 基础架构 (Scaffold Replacement)
*   **禁止使用** `Scaffold` 和 `AppBar`。
*   **必须使用** `NavigationView` 作为主外框。
*   **标题栏**：必须实现自定义标题栏（支持拖拽区域），使用 `DragToMoveArea` 配合 `window_manager`。

### 3.2 响应式断点 (Desktop-First)
针对桌面端大屏调整断点策略：
*   **Compact (< 640px)**: `NavigationPaneDisplayMode.compact` (仅显示图标)。
*   **Medium (640px - 1000px)**: `NavigationPaneDisplayMode.compact` 或 `top`。
*   **Expanded (> 1000px)**: `NavigationPaneDisplayMode.open` (固定侧边栏)。
*   **内容最大化控制**：对于设置页面或详情页，使用 `ConstrainedBox(constraints: BoxConstraints(maxWidth: 1200))` 防止在大屏上拉伸过度。

### 3.3 组件映射表 (Component Mapping)
Agent 必须严格执行以下替换：
| 禁用 (Material) | 必须使用 (Fluent UI) |
| :--- | :--- |
| `ElevatedButton` / `TextButton` | `FilledButton` / `Button` |
| `Switch` | `ToggleSwitch` |
| `Checkbox` | `Checkbox` (来自 fluent_ui) |
| `TextField` | `TextBox` |
| `AlertDialog` | `ContentDialog` |
| `ListView` (简单列表) | `ListView` 或 `TreeLayout` |
| `CircularProgressIndicator` | `ProgressRing` |
| `SnackBar` | `InfoBar` |

---

## 4. Windows 专属技能 (Desktop Skills)

### 4.1 窗口行为 (WindowManager Skill)
*   **启动设置**：在 `main.dart` 中初始化 `window_manager`，设置最小尺寸（如 800x600）。
*   **交互逻辑**：
    *   实现“点击关闭按钮时最小化到系统托盘”逻辑。
    *   窗口失去焦点时，UI 自动调整（如标题栏颜色变浅）。

### 4.2 交互增强
*   **鼠标悬停 (Hover)**：所有可点击区域必须有 `MouseCursor.clickable` 和背景色高亮反馈。
*   **右键菜单**：使用 `context_menu_bin` 或 `fluent_ui` 内置菜单实现桌面端特有的右键操作。
*   **键盘快捷键**：使用 `FocusableActionDetector` 绑定常用快捷键（如 Ctrl+S, Ctrl+F）。

---

## 5. 负面约束 (Negative Constraints - 重要)
*   **严禁混合**：禁止在同一个文件中 import `material.dart` 和 `fluent_ui.dart`（会导致命名冲突）。
*   **严禁移动端思维**：不要生成 `BottomNavigationBar` 或全屏抽屉菜单。
*   **严禁拉伸**：表单组件不允许横向铺满 1920px 屏幕，必须居中并限制最大宽度。
*   **严禁硬编码颜色**：必须使用 `FluentTheme.of(context).accentColor` 或 `FluentTheme.of(context).cardColor` 以适配暗黑模式。

---

## 6. Prompt 模板示例

### 生成新页面时：
> "Create a new Fluent UI settings page. Use `ScaffoldPage` with a `header`. Include a `ListView` containing `ListTile` for settings, and a `ToggleSwitch` for 'Dark Mode'. Ensure the layout is constrained to a maximum width of 800px and centered."

### 处理窗口逻辑时：
> "Implement a custom title bar using `window_manager`. It should include the app logo, a search box in the middle, and Windows-style minimize/maximize/close buttons on the right. Ensure the background uses the `Acrylic` effect."

### 响应式适配时：
> "Convert this layout to be responsive: if the window width is less than 1000px, change the `NavigationPane` to `compact` mode; otherwise, keep it `open`."

---

## 7. 代码风格规范
*   **Const Optimization**：所有静态组件必须加 `const`。
*   **Separation**：UI 与逻辑分离，复杂的 `NavigationPane` 项应抽离为独立的 `Widget` 函数或类。
*   **Assets**：Windows 图标建议优先使用 `.svg` 以保证在大屏高 DPI 下不失真。
