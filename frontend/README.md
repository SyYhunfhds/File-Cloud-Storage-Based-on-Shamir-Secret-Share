# CryptographyDesign - 财务条目托管终端

一个基于 **Flutter** 构建的 Windows 桌面端应用，专注于财务条目的安全上传、管理、共享与加密存储。项目采用 Feature-first 分层架构，结合 Riverpod 状态管理与 AES-256-GCM 加密，为企业内部提供端到端的敏感数据托管解决方案。

---

## 项目结构

```
lib/
├── core/                      # 基础设施层
│   ├── api_config_provider.dart       # 后端 API 地址配置
│   ├── api_utils.dart                 # 统一 API 错误处理
│   ├── auto_refresh_config_provider.dart  # 定时刷新配置
│   ├── constants.dart                 # 窗口尺寸、断点常量
│   ├── formatters.dart                # 格式化工具
│   └── theme.dart                     # Material 3 主题配置
├── desktop/                   # Windows 桌面适配
│   └── title_bar.dart                 # 自定义标题栏
├── features/                  # 业务功能模块（Feature-first）
│   ├── auth/                  #   认证（登录 / 注册 / 登出 / 快速登录）
│   ├── home/                  #   主页（条目列表 / 搜索 / 筛选 / 分页）
│   ├── items/                 #   条目数据模型 & API 服务
│   ├── upload/                #   条目上传（文件选择 / 拖拽 / 进度）
│   ├── download/              #   条目下载（ID + Share 还原）
│   ├── share/                 #   份额管理 UI（列表 / 详情 / 解密）
│   ├── shares/                #   份额核心业务（AES 加密 / Hive 持久化 / SSE 刷新）
│   ├── audit/                 #   审计管理（列表 / 通过 / 拒绝 / SSE 份额刷新）
│   └── settings/              #   设置（主题 / 账号管理 / 自动刷新 / 关于）
└── router/                    # GoRouter 路由配置（StatefulShellRoute）
```

---

## 技术栈

| 类别 | 技术选型 | 用途 |
|------|----------|------|
| **框架** | Flutter 3.11+ / Dart 3.11+ | 跨平台 UI |
| **状态管理** | Riverpod 3.3.1 | 全局状态 + 精确刷新 |
| **路由** | GoRouter 17.3.0 | 声明式路由 + 持久侧边栏 |
| **加密** | PointyCastle 3.9.1 | AES-256-GCM 份额加解密 |
| **安全存储** | flutter_secure_storage 9.2.4 | AES 密钥存储 |
| **本地存储** | Hive 2.2.3 | 加密份额本地持久化 |
| **HTTP** | dart:http 1.6.0 | REST API 通信 |
| **窗口管理** | window_manager 0.5.1 | Windows 窗口控制 |
| **文件交互** | file_picker + desktop_drop | 文件选择与拖拽上传 |
| **序列化** | json_annotation + json_serializable | 模型代码生成 |
| **系统集成** | tray_manager + url_launcher | 系统托盘 / 打开外部链接 |
| **路径工具** | path_provider + path + sanitize_filename | 文件路径 / 文件名处理 |

---

## 架构概览

### 分层设计

每个 Feature 内部遵循单向依赖链：

```
models/ → services/ → providers/ → views/ → widgets/
  (DTO)    (API封装)   (Riverpod)   (页面)   (组件)
```

### 状态管理

使用 Riverpod 3.x 的 `NotifierProvider` / `Notifier` 模式管理全局状态，关键 Provider 包括：

| Provider | 职责 |
|----------|------|
| `authProvider` | 登录态管理（Token、用户信息） |
| `entryListProvider` | 条目列表分页、筛选、搜索 |
| `uploadProvider` | 上传全生命周期 |
| `downloadProvider` | 文件下载流程 |
| `shareListProvider` | 份额列表管理 |
| `shareServiceProvider` | 份额加密存储服务 |
| `apiConfigProvider` | 后端连接配置 |

### 路由

采用 `StatefulShellRoute.indexedStack` 实现侧边导航持久化，所有页面共享 `AppShell`（NavigationRail + 内容区），切换时保持页面状态。

### 响应式布局

- **< 640px**: 紧凑模式
- **640px - 1000px**: 侧边栏仅图标
- **> 1000px**: 侧边栏扩展（图标 + 标签）
- **内容区**: 最大宽度 1400px，居中约束

---

## 功能模块

### 已完成

- [x] **用户认证**：登录 / 注册 / 登出，JWT Bearer Token 鉴权
- [x] **快速登录**：缓存最近登录账号，下次启动可一键快速登录
- [x] **条目列表**：分页查询、三种筛选范围（我的 / 公开 / 全部）、客户端文件名搜索
- [x] **条目上传**：文件选择、桌面拖拽、进度追踪、Recovery Code 展示
- [x] **条目下载**：ID + Device Share 还原文件
- [x] **份额本地加密存储**：AES-256-GCM 加密后存入 Hive，按用户隔离
- [x] **份额管理**：列表展示（掩码份额）、详情查看、开发者模式解密到剪贴板
- [x] **SSE 份额刷新**：通过 Server-Sent Events 实时接收份额刷新进度，进度百分比弹窗展示
- [x] **审计管理**：审计列表分页查询、按状态筛选（已通过 / 已拒绝 / 待审核）、通过 / 拒绝操作
- [x] **Windows 桌面适配**：自定义标题栏、窗口居中、最小尺寸 1024×700
- [x] **快捷键支持**：ESC 返回、翻页快捷键等
- [x] **亮/暗色主题**：Material 3 主题，跟随系统主题切换
- [x] **设置页**：主题切换、账号管理、定时自动刷新配置、API 地址配置、关于页面

### 进行中 / 待完成

- [ ] **条目详情页**：API 端点已定义但未接入 UI
- [ ] **条目修改功能**：API 端点已定义但未实现
- [ ] **前端大文件分片上传**：尚未实现
- [ ] **单元测试 / 集成测试**：`test/` 目录为空
- [ ] **`fluent_ui` 组件库迁移**：当前仍使用 Material Design 3
- [ ] **右键上下文菜单**：尚未实现桌面端右键交互

---

## 已知问题

### 1. 系统警告音

- 文件提交窗口的 `PopScope`/`ModalBarrier` 拦截机制在特定条件下触发 Windows `HTERROR` 警告音
- **状态**：低优先级，不影响功能

### 2. >1MB 文件份额解密错误 (@BUG-001)

- **描述**：上传大于 1MB 的文件后，Device Share 进行 AES-GCM 解密时得到的份额可能不一致
- **排查方向**：multipart 传输完整性或 HTTP 框架对大请求体的缓冲区限制
- **状态**：不影响基本功能，暂时搁置

### 3. 切换/退出登录后旧 Device Share 可能不可用 (@BUG-002)

- **描述**：用户切换或退出登录再登录回来时，旧的 Device Share 可能无法使用
- **排查方向**：可能与 `userId` 识别精度或密钥绑定策略有关
- **状态**：不影响基本功能，暂时搁置

---

## 开发指南

### 环境要求

- Flutter SDK: stable channel (>=3.11.5)
- Dart SDK: ^3.11.5

### 启动

```bash
# 安装依赖
flutter pub get

# 运行（Windows 桌面端）
flutter run -d windows
```

### 代码规范

- 项目遵循 `analysis_options.yaml` 中的 lint 规则
- Feature-first 目录结构，每个功能模块独立分层
- UI 组件使用 `ConsumerWidget` / `ConsumerStatefulWidget` 实现精确刷新
- 桌面端交互遵循 Windows 平台习惯：鼠标悬停反馈、快捷键绑定
- Commit 遵循 Conventional Commits 规范（前缀 `[Trae]`）
