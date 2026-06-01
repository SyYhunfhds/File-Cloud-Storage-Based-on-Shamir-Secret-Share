# 帮助信息页面实现计划

## 需求分析

根据后端 API 文档 `/v1/about`，实现一个帮助信息页面，包含版本号、项目领导人、开发者列表等信息。

### 功能要求
- 展示版本号
- 展示项目领导人
- 展示开发者列表
- 支持数据缓存
- 支持手动刷新
- 无需登录即可访问

## 实现方案

### 1. 创建 About DTO 模型
- 新建 `lib/models/about.dart`
- 使用 json_serializable 注解
- 字段：`version`, `leader`, `developers`

### 2. 更新 API 服务
- 在 `lib/services/api_service.dart` 添加 `getAboutInfo()` 方法
- 调用 `/v1/about` 接口
- 返回 `ApiResponse<About>`

### 3. 更新常量配置
- 在 `lib/utils/constants.dart` 添加 API 路径常量

### 4. 创建帮助页面
- 新建 `lib/pages/about_page.dart`
- 使用 Riverpod 管理状态
- 显示信息卡片布局
- 添加刷新按钮
- 实现缓存机制（使用 shared_preferences）

### 5. 更新路由配置
- 在 `lib/router.dart` 添加 `/about` 路由
- 无需登录验证

### 6. 添加导航入口
- 在 `lib/pages/home_page.dart` 的侧边栏和功能卡片中添加帮助信息入口
- 在其他页面的侧边栏也添加统一入口

## 修改文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `lib/models/about.dart` | 新建 | About 数据模型 |
| `lib/utils/constants.dart` | 修改 | 添加 API 路径常量 |
| `lib/services/api_service.dart` | 修改 | 添加获取帮助信息方法 |
| `lib/pages/about_page.dart` | 新建 | 帮助信息页面 |
| `lib/router.dart` | 修改 | 添加路由配置 |
| `lib/pages/home_page.dart` | 修改 | 添加导航入口 |
| `lib/pages/share_page.dart` | 修改 | 添加导航入口 |
| `lib/pages/audit_page.dart` | 修改 | 添加导航入口 |

## 步骤详情

### 步骤 1: 创建 About DTO
- 使用 `@JsonSerializable()` 注解
- 包含 `String version`, `String leader`, `List<String> developers`
- 运行 `flutter pub run build_runner build`

### 步骤 2: 扩展 API 服务
- 添加 `getAboutInfo()` 方法
- 返回 `Future<ApiResponse<About>>`

### 步骤 3: 创建帮助页面
- 使用 ConsumerStatefulWidget
- 实现数据缓存逻辑
- 实现刷新功能
- 美观的 UI 设计（卡片式布局）

### 步骤 4: 更新所有页面导航
- 在所有包含侧边栏的页面添加"帮助信息"菜单项

## 技术要点

### 缓存策略
- 使用 shared_preferences 缓存 About 数据
- 记录缓存时间（可选）
- 提供手动刷新功能

### 路由访问权限
- 帮助页面不受登录限制，可直接访问
- 不添加到路由守卫的受保护列表中

### UI 设计
- 顶部显示应用图标和标题
- 中间卡片显示版本号、领导人
- 下方列表显示开发者
- 右上角刷新按钮
