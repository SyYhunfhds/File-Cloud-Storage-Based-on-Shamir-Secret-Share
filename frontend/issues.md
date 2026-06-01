## 须知
- 以下测试结果均在Windows热重载模式下测出
- 后端有Jaeger链路追踪, 所以用户说请求成功了就是成功了, 失败了就是失败了

### UI问题
- 登出按钮点了没反应, 显示请求成功, 但页面并没有退出登录, 而且可以继续访问其他页面, 也能拉取条目且不会被拦截
- 部分问题在于后端并没有实现基于Redis的JWT吊销机制, 但前端是主责
```
[DEBUG] parseApiResponse - code: 0, message: 登出成功, isSuccess: true
[DEBUG] parseApiResponse - 原始响应体: {"code":0,"message":"登出成功"}
[DEBUG] parseApiResponse - 解析后的JSON: {code: 0, message: 登出成功}
[DEBUG] parseApiResponse - code: 0, message: 登出成功, isSuccess: true
[ERROR:flutter/runtime/dart_vm_initializer.cc(40)] Unhandled Exception: CircularDependencyError: Circular dependency detected.
This happens when a provider somehow depends on itself.

The circular dependency chain is as follows:
  NotifierProvider<ShareListNotifier, ShareListState>#e206e


#0      Ref._debugAssertCanDependOn (package:riverpod/src/core/ref.dart:183:9)
#1      Ref.invalidate (package:riverpod/src/core/ref.dart:324:21)
#2      AuthNotifier.logout (package:frontend/features/auth/providers/auth_provider.dart:179:9)
<asynchronous suspension>
#3      HomeToolbar._buildUserArea.<anonymous closure> (package:frontend/features/home/widgets/toolbar.dart:152:27)
<asynchronous suspension>
```

