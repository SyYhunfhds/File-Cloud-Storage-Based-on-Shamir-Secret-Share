# 路由跳转修复计划

## 问题分析
在 `router.dart` 的 `redirect` 方法中使用 `ref.read(authProvider)` 只会在路由初始化时读取一次状态，之后登录成功更新状态后，`ref.read` 不会重新读取，导致路由重定向逻辑不生效。

## 修复方案
在登录成功后，使用 `refreshListenable` 让路由监听认证状态变化。

## 具体步骤

### 1. 修改 `lib/router.dart`
- 创建 `AuthNotifier` 类继承 `ChangeNotifier`
- 使用 `ref.listen` 监听 `authProvider` 的状态变化
- 在 `GoRouter` 中设置 `refreshListenable`

### 2. 确保登录页面正确跳转
- 登录成功后调用 `context.go('/')`

### 3. 验证修复
- 测试未认证用户访问 `/` 时重定向到 `/login`
- 测试登录成功后跳转到 `/` 并正常显示
- 测试退出登录后重定向到 `/login`
