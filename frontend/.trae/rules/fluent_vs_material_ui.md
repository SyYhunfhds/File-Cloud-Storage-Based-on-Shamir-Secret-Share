---
alwaysApply: false
---
# 🛑 强制性规则：禁止使用 Material 库
1. **禁止导入**：严禁在任何文件中使用 `import 'package:flutter/material.dart';`。
2. **唯一导入**：必须且仅能使用 `import 'package:fluent_ui/fluent_ui.dart';`。
3. **报错修正逻辑**：如果编译器提示“Widget not defined”，Agent 必须优先检查是否误用了 Material 组件，并对照下表进行重写。

---

# 📑 Fluent UI 深度组件替换指南

### 1. 基础容器与结构 (Structure)
| Material 组件 (禁用) | Fluent UI 组件 (替代) | 关键差异备注 |
| :--- | :--- | :--- |
| `Scaffold` | **`NavigationView`** | 核心架构。使用 `pane` 属性定义侧边栏，`content` 定义主体。 |
| `AppBar` | **`PageHeader` / `CommandBar`** | `PageHeader` 用于页面标题；`CommandBar` 用于操作按钮。 |
| `BottomNavigationBar` | **`NavigationPane` (top/left)** | 桌面端通常使用侧边模式。 |
| `Drawer` | **`NavigationPane` (open)** | 侧边栏通过 `displayMode` 切换状态。 |
| `Divider` | **`Divider`** | 名字相同，但确保来自 `fluent_ui`。 |
| `VerticalDivider` | **`Divider(direction: Axis.vertical)`** | |

### 2. 按钮与交互 (Buttons)
| Material 组件 (禁用) | Fluent UI 组件 (替代) | 关键差异备注 |
| :--- | :--- | :--- |
| `ElevatedButton` | **`FilledButton`** | 强调色背景按钮。 |
| `TextButton` | **`Button`** | 普通样式按钮。 |
| `OutlinedButton` | **`Button`** | Fluent UI 的普通 Button 自带轻微边框。 |
| `IconButton` | **`IconButton`** | 属性类似，但注意 `Icon` 需要用 `FluentIcons`。 |
| `FloatingActionButton` | **禁止使用** | 桌面端不应有悬浮按钮，改为 `CommandBar`。 |
| `TextButton` (超链接) | **`HyperlinkButton`** | 专门用于链接跳转。 |

### 3. 表单与输入 (Input)
| Material 组件 (禁用) | Fluent UI 组件 (替代) | 关键差异备注 |
| :--- | :--- | :--- |
| `TextField` | **`TextBox`** | 核心输入框，支持 `prefix` 和 `suffix`。 |
| `Switch` | **`ToggleSwitch`** | 风格为 Windows 11 开关。 |
| `Checkbox` | **`Checkbox`** | 注意：`checked` 属性代替了 `value`。 |
| `Radio` | **`RadioButton`** | 使用方式基本一致。 |
| `Slider` | **`Slider`** | |
| `DropdownButton` | **`ComboBox` / `DropDownButton`** | `ComboBox` 用于选择，`DropDownButton` 用于弹出菜单。 |
| `SearchDelegate` | **`AutoSuggestBox`** | 桌面端标准的搜索建议框。 |

### 4. 数据展示 (Data & Lists)
| Material 组件 (禁用) | Fluent UI 组件 (替代) | 关键差异备注 |
| :--- | :--- | :--- |
| `ListTile` | **`ListTile`** | 名字相同，但 `fluent_ui` 的 `ListTile` 不支持 `subtitle`（请使用 `subtitle` 的替代方案或嵌套 Column）。 |
| `Card` | **`Card`** | 建议使用 `Card` 或直接用 `Container` 配合 `FluentTheme.of(context).cardColor`。 |
| `DataTable` | **`TreeLayout` / `HorizontalScrollView`** | Fluent UI 官方暂无复杂的 Table，通常用 `ListView` 自行构建行。 |
| `ExpansionTile` | **`Expander`** | Windows 标准的可折叠面板。 |

### 5. 弹窗与通知 (Feedback)
| Material 组件 (禁用) | Fluent UI 组件 (替代) | 关键差异备注 |
| :--- | :--- | :--- |
| `AlertDialog` | **`ContentDialog`** | 必须调用 `showDialog`。支持 `primaryButton` 和 `secondaryButton`。 |
| `SnackBar` | **`InfoBar`** | 非常重要！Windows 风格通知应显示在顶部或内联，使用 `InfoBar`。 |
| `Tooltip` | **`Tooltip`** | 基本一致。 |
| `CircularProgress` | **`ProgressRing`** | Windows 的圆圈进度条。 |
| `LinearProgress` | **`ProgressBar`** | |

### 6. 视觉符号 (Icons & Typography)
| Material 组件 (禁用) | 必须使用 (Fluent UI) |
| :--- | :--- |
| `Icons.xxx` | **`FluentIcons.xxx`** (需要安装 `fluentui_system_icons`) |
| `TextTheme` | **`Typography`** (例如 `FluentTheme.of(context).typography.title`) |
| `Colors.blue` | **`AccentColor.swatch(options)`** (Windows 使用强调色系统) |

