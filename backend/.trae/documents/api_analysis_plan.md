
# API分析与测试数据生成计划

## 项目概述

本项目是一个基于GoFrame框架的后端服务，包含认证和条目管理相关的API。

## API定义分析

### 1. 认证相关API (`api/auth/v1/auth.go`)

| API路径 | 方法 | 结构定义 | Handler |
|---------|------|----------|---------|
| `/auth/login` | POST | LoginReq / LoginRes | PortableLoginHandler |
| `/auth/logout` | POST | LogoutReq / LogoutRes | PortableLogoutHandler |

### 2. 条目管理相关API (`api/item/v1/item.go`)

| API路径 | 方法 | 结构定义 | Handler |
|---------|------|----------|---------|
| `/item/submit` | POST | ItemSubmitReq / ItemSubmitRes | PortableItemSubmit |
| `/item/download` | GET/POST | ItemDownloadReq / ItemDownloadRes | PortableItemDownload |

### 3. 不由gf工具生成的Portable Handler

根据代码分析，以下Handler是手动实现的（非gf工具生成）：

| Handler函数 | 文件路径 | 功能描述 |
|-------------|----------|----------|
| `PortableLoginHandler` | `internal/controller/auth/auth_v1_login.go` | 用户登录，支持用户名/邮箱+密码验证 |
| `PortableLogoutHandler` | `internal/controller/auth/auth_v1_logout.go` | 用户登出，清除Cookie |
| `PortableCookiePrintHandler` | `internal/controller/auth/auth_v1_debug.go` | 调试用，打印JWT Cookie信息（仅开发模式） |
| `PortableItemSubmit` | `internal/controller/item/item_v1_item_submit.go` | 提交条目，加密保存并生成份额 |
| `PortableItemDownload` | `internal/controller/item/item_v1_item_download.go` | 下载条目，需要设备份额恢复密钥 |

## 计划步骤

### 步骤1：创建测试数据目录

确认 `tests/data/` 目录存在，如不存在则创建。

### 步骤2：生成认证API测试数据

创建 `tests/data/auth/` 目录，生成：
- `login_request.json5` - 登录请求示例
- `login_response.json5` - 登录响应示例
- `logout_request.json5` - 登出请求示例
- `logout_response.json5` - 登出响应示例

### 步骤3：生成条目管理API测试数据

创建 `tests/data/item/` 目录，生成：
- `item_submit_request.json5` - 条目提交请求示例
- `item_submit_response.json5` - 条目提交响应示例
- `item_download_request.json5` - 条目下载请求示例
- `item_download_response.json5` - 条目下载响应示例

### 步骤4：生成DTO字段注释文档

在每个json5文件中添加字段含义注释。

## 预期输出

```
tests/data/
├── auth/
│   ├── login_request.json5
│   ├── login_response.json5
│   ├── logout_request.json5
│   └── logout_response.json5
└── item/
    ├── item_submit_request.json5
    ├── item_submit_response.json5
    ├── item_download_request.json5
    └── item_download_response.json5
```
