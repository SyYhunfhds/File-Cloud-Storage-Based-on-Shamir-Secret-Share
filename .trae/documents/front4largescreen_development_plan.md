# Front4LargeScreen 开发计划

## 项目概述
为企业财务审计条目存储系统开发大屏桌面端应用，使用 Flutter + Material Design 风格。

## 技术栈
- **框架**: Flutter (Material Design)
- **状态管理**: flutter_riverpod
- **路由管理**: go_router
- **网络请求**: dio
- **JSON 序列化**: json_annotation + json_serializable
- **安全存储**: flutter_secure_storage
- **本地存储**: shared_preferences
- **文件选择**: file_picker

## 开发步骤

### 1. 项目初始化与依赖配置
- 更新 pubspec.yaml，添加所有必要依赖
- 配置项目基础结构

### 2. 创建数据模型 (Models)
- 通用响应模型
- 认证相关模型 (登录、注册)
- 用户信息模型
- 条目管理模型
- 关于信息模型

### 3. 创建服务层 (Services)
- API 服务封装 (使用 Dio)
- 本地存储服务
- 认证服务

### 4. 创建状态管理 (Providers)
- 认证状态提供者
- 条目列表状态提供者
- 用户信息状态提供者

### 5. 创建工具类 (Utils)
- 常量定义
- 验证器

### 6. 创建页面 (Pages)
- 登录页 (大屏布局)
- 注册页 (大屏布局)
- 首页 (条目列表)
- 条目提交页
- 关于页
- 用户信息页

### 7. 路由配置
- 使用 go_router 配置路由

### 8. 主应用配置
- 主题配置
- 初始化 ProviderScope

## 关键特性

### 大屏布局设计
- 使用 Row/Column 布局实现桌面端专业软件风格
- 适当的间距和卡片设计
- 响应式布局（适配不同大屏尺寸）

### 认证流程
- 登录/注册界面
- JWT Token 管理
- 自动登录

### 条目管理
- 条目列表展示（分页）
- 条目上传（支持文件选择）
- 份额和恢复码展示
- 条目更新
- 条目删除

### 用户体验
- 加载状态提示
- 错误提示
- 成功提示
- 恢复码掩码显示 + 复制功能

## 文件结构
```
lib/
├── main.dart
├── router.dart
├── models/
├── pages/
├── providers/
├── services/
└── utils/
```
