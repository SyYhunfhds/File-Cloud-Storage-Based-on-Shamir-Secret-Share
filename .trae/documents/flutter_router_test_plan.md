# Flutter 路由跳转测试计划

## 测试目标
验证 front4largescreen 应用的路由跳转功能是否正常工作。

## 当前代码结构
- 路由配置: `lib/router.dart`
- 认证状态: `lib/providers/auth_provider.dart`
- 登录页面: `lib/pages/login_page.dart`
- 首页: `lib/pages/home_page.dart`

## 测试数据
### 用户名+密码
```json
{
    "username": "Christ",
    "password": "123456abcABC"
}
```
### 邮箱+密码
```json
{
    "email": "Kacey_Wolff@yahoo.com",
    "password": "123456abcABC"
}
```

## 测试场景

### 1. 初始化路由测试
- **场景 1.1**: 应用启动时，未认证用户应该被重定向到 `/login`
- **预期**: 初始页面为 `/login`

### 2. 认证保护测试
- **场景 2.1**: 未认证用户访问 `/` 时，应该重定向到 `/login`
- **场景 2.2**: 未认证用户访问 `/about` 时，应该重定向到 `/login`
- **预期**: 访问受保护路由时自动跳转到登录页

### 3. 已认证用户路由测试
- **场景 3.1**: 已认证用户访问 `/login` 时，应该保持在 `/login` 或重定向到 `/`
- **预期**: 不应重复登录页面

### 4. 登录跳转测试
- **场景 4.1**: 登录成功后跳转到 `/`
- **场景 4.2**: 登录失败时保持当前页面
- **预期**: 登录成功跳转到首页

### 5. 退出登录测试
- **场景 5.1**: 退出登录后应该跳转到 `/login`
- **预期**: 退出登录后返回登录页

## 测试实现步骤

### 步骤 1: 创建测试目录结构
```
test/
├── router/
│   └── router_test.dart
└── widget/
    └── login_page_test.dart
```

### 步骤 2: 添加测试依赖
在 `pubspec.yaml` 的 `dev_dependencies` 中添加:
```yaml
mocktail: ^1.0.0
```

### 步骤 3: 创建测试文件

#### 3.1 创建路由测试文件 `test/router/router_test.dart`
- 使用 `go_router_test` 包进行路由测试
- 创建 mock `AuthNotifier` 来模拟认证状态
- 测试路由重定向逻辑

#### 3.2 创建认证状态测试 `test/auth/auth_provider_test.dart`
- 测试认证状态初始化
- 测试登录成功状态更新
- 测试登出状态清理

### 步骤 4: 运行测试
```bash
flutter test
```

## 测试技术栈
- `flutter_test`: Flutter 测试框架
- `mocktail`: Mock 对象库（轻量级，无需代码生成）
- `go_router_test`: GoRouter 官方测试助手（可选）

## 注意事项
- 测试应该独立运行，不依赖外部服务
- 使用 mock 模拟 API 调用
- 确保测试覆盖所有路由跳转场景
- 测试应该在 CI/CD 环境中可重复执行
