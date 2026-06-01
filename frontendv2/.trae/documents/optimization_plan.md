# Flutter 前端项目优化规划

## 一、项目概述

本项目是一个企业财务审计条目存储系统的 Flutter 桌面端应用，基于 Riverpod + GoRouter + Dio 技术栈构建。

## 二、待优化点分析

### 优化点 1：路由守卫（登录拦截）

**问题描述**：
当前路由配置未实现登录拦截机制，未登录用户可以直接访问受保护的页面（首页、份额管理、审计条目）。

**影响范围**：
- `lib/router.dart`

**优化方案**：
1. 使用 GoRouter 的 `redirect` 机制实现全局路由守卫
2. 在跳转前检查用户登录状态（通过 AuthProvider）
3. 未登录时自动重定向到登录页

---

### 优化点 2：Token 过期自动处理

**问题描述**：
当前 `ApiService` 中检测到 Token 过期或无效时，仅打印日志，未执行自动跳转登录页的操作。

**影响范围**：
- `lib/services/api_service.dart`

**优化方案**：
1. 在响应拦截器中检测 Token 过期错误码（401）
2. 清除本地存储的 Token
3. 触发全局状态登出（通过 AuthNotifier）
4. 跳转到登录页

---

### 优化点 3：状态持久化

**问题描述**：
Riverpod Provider 状态在应用重启后会丢失，用户需要重新登录。

**影响范围**：
- `lib/providers/auth_provider.dart`

**优化方案**：
1. 在 `AuthNotifier` 初始化时读取本地存储的用户信息
2. 登录成功后持久化用户信息到 `shared_preferences`

---

### 优化点 4：代码复用 - 服务初始化

**问题描述**：
每个页面都重复初始化 `ApiService` 和 `StorageService`，代码冗余。

**影响范围**：
- `lib/pages/login_page.dart`
- `lib/pages/home_page.dart`
- `lib/pages/share_page.dart`
- `lib/pages/audit_page.dart`

**优化方案**：
1. 使用 Riverpod 提供全局单例服务
2. 创建 `apiServiceProvider` 和 `storageServiceProvider`
3. 通过 `ref.watch()` 在页面中获取服务实例

---

## 三、JSON Serializable 重构分析

### 3.1 当前已使用 json_serializable 的 DTO

| 文件 | 类名 | 状态 |
|------|------|------|
| `lib/models/user.dart` | `User` | ✅ 已使用 |
| `lib/models/item.dart` | `ItemSubmitResponse`, `ItemInfo`, `ItemListResponse` | ✅ 已使用 |
| `lib/models/share.dart` | `Share` | ✅ 已使用 |
| `lib/models/audit_item.dart` | `AuditItem` | ✅ 已使用 |

### 3.2 需要重构为 json_serializable 的 DTO

| 文件 | 类名 | 原因 |
|------|------|------|
| `lib/models/response.dart` | `ApiResponse<T>` | 手动实现 JSON 解析，需改为 json_serializable |

### 3.3 重构优先级

| 优先级 | DTO | 说明 |
|--------|-----|------|
| 高 | `ApiResponse<T>` | 通用响应模型，影响所有 API 调用 |

---

## 四、实施计划

### 阶段一：JSON Serializable 重构

| 步骤 | 任务 | 文件 | 预估时间 |
|------|------|------|----------|
| 1 | 重构 `ApiResponse<T>` 支持 json_serializable | `lib/models/response.dart` | 30分钟 |
| 2 | 运行 build_runner 生成序列化代码 | - | 10分钟 |

### 阶段二：路由守卫实现

| 步骤 | 任务 | 文件 | 预估时间 |
|------|------|------|----------|
| 1 | 修改 `router.dart` 添加 redirect 逻辑 | `lib/router.dart` | 30分钟 |

### 阶段三：Token 过期处理

| 步骤 | 任务 | 文件 | 预估时间 |
|------|------|------|----------|
| 1 | 修改 ApiService 添加 Token 过期处理 | `lib/services/api_service.dart` | 30分钟 |
| 2 | 集成全局状态管理 | `lib/providers/auth_provider.dart` | 20分钟 |

### 阶段四：状态持久化

| 步骤 | 任务 | 文件 | 预估时间 |
|------|------|------|----------|
| 1 | 修改 AuthNotifier 添加持久化逻辑 | `lib/providers/auth_provider.dart` | 30分钟 |
| 2 | 测试状态恢复功能 | - | 10分钟 |

### 阶段五：服务复用优化

| 步骤 | 任务 | 文件 | 预估时间 |
|------|------|------|----------|
| 1 | 创建服务 Provider | `lib/providers/service_providers.dart` | 30分钟 |
| 2 | 重构各页面使用 Provider 获取服务 | `lib/pages/*.dart` | 60分钟 |

---

## 五、依赖与风险

### 依赖关系

```
路由守卫 → AuthProvider
Token过期处理 → ApiService + AuthProvider
状态持久化 → AuthProvider + StorageService
服务复用 → 所有页面
```

### 风险点

| 风险 | 描述 | 缓解措施 |
|------|------|----------|
| 路由循环 | redirect 可能导致无限循环 | 添加 `redirectLimit` 检查 |
| 状态不一致 | 持久化与内存状态不同步 | 使用 `AsyncNotifier` 管理异步状态 |
| 代码冲突 | 多人开发可能产生冲突 | 分阶段实施，及时提交 |

---

## 六、测试计划

| 测试项 | 测试方法 | 预期结果 |
|--------|----------|----------|
| 路由守卫 | 直接访问 `/` 路径 | 重定向到 `/login` |
| Token过期 | 修改Token后调用API | 自动跳转到登录页 |
| 状态持久化 | 登录后重启应用 | 保持登录状态 |
| 服务复用 | 检查各页面服务实例 | 共享同一实例 |

---

## 七、代码规范

1. **命名规范**：
   - Provider 使用 `xxxProvider` 命名
   - Notifier 使用 `xxxNotifier` 命名
   - DTO 使用 PascalCase

2. **文件结构**：
   - Providers 统一放在 `lib/providers/`
   - Models 统一放在 `lib/models/`
   - Utils 统一放在 `lib/utils/`

3. **JSON 序列化**：
   - 所有 DTO 必须使用 `json_serializable`
   - 使用 `@JsonKey` 处理特殊字段名
   - 生成的 `.g.dart` 文件不应手动修改

---

## 八、验收标准

1. ✅ 未登录用户无法访问受保护页面
2. ✅ Token 过期自动跳转登录页
3. ✅ 登录状态在重启后保持
4. ✅ 所有服务通过 Provider 提供
5. ✅ 所有 DTO 使用 json_serializable
6. ✅ 代码通过 `flutter analyze` 检查
7. ✅ 代码通过 `flutter test` 测试