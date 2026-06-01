# Flutter 路由跳转单元测试计划

## 测试目标
通过单元测试验证 front4largescreen 的路由跳转功能，定位问题所在。

## 当前代码结构
```
lib/
├── router.dart              # 路由配置
├── providers/
│   └── auth_provider.dart   # 认证状态管理
└── pages/
    ├── login_page.dart      # 登录页
    ├── home_page.dart       # 首页
    └── register_page.dart  # 注册页
```

## 测试策略
1. **Mock 认证状态**：模拟不同的认证状态
2. **测试路由重定向逻辑**：验证 redirect 函数的行为
3. **端到端测试**：测试完整的登录流程和页面跳转

## 测试场景清单

### 测试场景 1: 路由初始化测试
```dart
test('应用启动时，未认证用户应重定向到 /login', () {
  // 验证初始路由为 /login
});
```

### 测试场景 2: 受保护路由重定向测试
```dart
test('未认证用户访问 / 时应重定向到 /login', () {
  // 设置未认证状态
  // 尝试访问 /
  // 验证重定向到 /login
});
```

### 测试场景 3: 已认证用户路由测试
```dart
test('已认证用户访问 /login 应保持在 /login 或重定向到 /', () {
  // 设置已认证状态
  // 尝试访问 /login
  // 验证行为
});
```

### 测试场景 4: 登录成功跳转测试
```dart
test('登录成功后应跳转到 /', () async {
  // 模拟登录成功
  // 调用 context.go('/')
  // 验证页面跳转到 /
});
```

### 测试场景 5: 退出登录跳转测试
```dart
test('退出登录后应跳转到 /login', () async {
  // 设置已认证状态
  // 调用退出登录
  // 验证跳转到 /login
});
```

## 测试文件结构
```
test/
├── router/
│   ├── router_redirect_test.dart    # 路由重定向测试
│   └── router_initialization_test.dart # 路由初始化测试
├── auth/
│   └── auth_state_test.dart         # 认证状态测试
└── integration/
    └── login_flow_test.dart         # 登录流程集成测试
```

## 实现步骤

### 步骤 1: 创建测试目录和文件
创建以下测试文件：
- `test/router/router_redirect_test.dart`
- `test/router/router_initialization_test.dart`
- `test/auth/auth_state_test.dart`

### 步骤 2: 添加测试依赖
在 `pubspec.yaml` 中添加：
```yaml
dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^6.0.0
  mocktail: ^1.0.0  # Mock 库，无需代码生成
```

### 步骤 3: 编写路由重定向测试
测试 `router.dart` 中的 `redirect` 函数逻辑：
- 测试未认证用户访问受保护路由时的行为
- 测试已认证用户访问公开路由时的行为

### 步骤 4: 编写认证状态测试
测试 `auth_provider.dart` 中的状态管理：
- 测试 `loginSuccess` 方法是否正确更新状态
- 测试 `logout` 方法是否正确清理状态
- 测试 `isAuthenticated` 字段是否正确更新

### 步骤 5: 运行测试并分析结果
```bash
flutter test
```

根据测试结果定位问题：
- 如果测试失败，测试框架会指出具体的失败原因
- 如果测试通过但功能不工作，说明测试覆盖不够全面

## 测试工具选择
- **mocktail**: 轻量级 Mock 库，无需代码生成
- **flutter_test**: Flutter 官方测试框架
- **go_router**: 用于测试页面跳转

## 预期产出
- 至少 5 个测试用例
- 每个测试用例有清晰的断言
- 测试结果能够定位路由跳转问题
