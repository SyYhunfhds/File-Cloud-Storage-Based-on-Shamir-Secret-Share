# CryptographyDesign

基于 **Shamir 秘密共享（Shamir Secret Sharing）** 的文件加密云托管系统——一个密码学课程设计项目，目标是验证 Shamir SS 份额分布存储方案。

---

## 项目简介

本项目是一个单体文件加密云托管应用，核心流程：

```
用户上传文件 → AES-256-GCM 加密存储 → Shamir SS (2-of-3) 分割 AES 密钥
    ├── Device Share  → 返回客户端保存
    ├── Auth Share    → 存入 PostgreSQL（服务端持有）
    └── Recovery Share → 存入 PostgreSQL（用 Recovery Code 恢复）

下载时：用户提交 Device Share + 服务端 Auth Share → Lagrange 插值恢复密钥 → 解密文件 → 传输给用户
```

项目由三大部分组成：

| 部分 | 说明 | 开发模式 |
|------|------|----------|
| `exploration/` | 算法预研与验证工作区（7 个子项目） | Agent 全流程生成 |
| `backend/` | GoFrame v2 Go 后端（工程实现） | 人工开发 |
| `frontend/` | Flutter 桌面客户端 | Agent 全流程生成 |

> **课设背景**：本项目为密码学课程设计，目标是"能跑"和验证 Shamir SS 份额分布存储方案，**不面向高并发生产场景**。业务表单未使用缓存机制，直接影响是没有 TOKEN 吊销机制。

---

## 项目架构

### 目录结构

```
CryptographyDesign/
├── backend/                        # Go 后端
│   ├── main.go                     # 入口
│   ├── internal/
│   │   ├── cmd/cmd.go              # 路由注册、依赖注入、启动
│   │   ├── controller/             # 控制器层
│   │   │   ├── item/               # 文件条目 CRUD + 上传下载
│   │   │   ├── user/               # 用户管理
│   │   │   ├── auth/               # 登录、注销、JWT 中间件
│   │   │   ├── audit/              # 审计审批
│   │   │   ├── share/              # 份额刷新 / 拉取
│   │   │   ├── hello/              # 健康检查
│   │   │   └── version/            # 版本 / 关于信息
│   │   ├── logic/                  # 业务逻辑工具
│   │   │   ├── crypto-kits.go      # AES-GCM + Shamir 加解密
│   │   │   ├── file-kits.go        # 文件读写 + 加密文件管理
│   │   │   ├── hash-kits.go        # Argon2 密码哈希
│   │   │   └── cookie-kits.go      # JWT 签发 / 校验
│   │   ├── dao/                    # 数据访问层 (GoFrame gdb ORM)
│   │   ├── config/                 # 配置结构体
│   │   ├── claims/                 # JWT Claims 上下文注入
│   │   └── module/watcher/         # 文件监控模块
│   ├── pkg/shamir/v3/              # 自研 Shamir SS over GF(2^32-5)
│   │   ├── field.go                # GF(2^32-5) 域运算
│   │   ├── shamir.go               # Split / Recover / Delta / ApplyDelta
│   │   └── padding.go              # PKCS#7 风格 4 字节对齐填充
│   └── api/                        # GoFrame 接口定义
│
├── frontend/                       # Flutter 桌面端（Windows）
│   └── lib/
│       ├── main.dart               # 入口
│       ├── router/                 # GoRouter 路由
│       ├── core/                   # 主题 / 常量 / 工具
│       └── features/               # 功能模块（上传、份额、审计等）
│
├── exploration/                    # 算法预研与验证（Agent 生成）
│   ├── shamir-real/                # 实数域 SSS 教学验证
│   ├── shamir-galois/              # GF(257) 有限域 SSS + 加法同态
│   ├── shamir-image-mode/          # 灰度图像逐像素 SSS
│   ├── shamir-srgb-images-input/   # sRGB 四通道图像 SSS + msgpack
│   ├── shamir-go/                  # Go 语言 SSS 套具 + 性能基准
│   ├── shamir-share-update/        # 零值扰动多项式（PSS）验证
│   └── shamir-uint32-array-input/  # GF(2^32-5) 完整 PSS（最终参考实现）
│
└── manifest/                       # K8s/Docker 部署清单
```

### 算法演进路线

`exploration/` 中的 7 个子项目形成一条从理论到工程的递进路径，最终收敛到 `backend/pkg/shamir/v3/`：

```
实数域 → GF(257) → 同态特性 → 图像输入 → sRGB 四通道 → Go 套具 → PSS(GF257) → GF(2^32-5)+完整 PSS
                                                                                ↓
                                                                     backend/pkg/shamir/v3/
```

### 后端分层架构

```
┌──────────────────────────────────────────────────┐
│  api/         接口定义（GoFrame I/F 风格）         │
├──────────────────────────────────────────────────┤
│  controller/  控制器                             │
│  ├── 参数解析 & 校验                              │
│  ├── 调用 logic 层工具                            │
│  └── 调用 DAO 层操作数据库                         │
├──────────────────────────────────────────────────┤
│  logic/       业务逻辑工具（无状态）               │
│  ├── CryptoUtils  AES-GCM + Shamir Split/Recover │
│  ├── FileUtils    文件加密读写                    │
│  ├── HashUtils    Argon2 密码哈希                │
│  └── CookieUtils  JWT 签发/校验                  │
├──────────────────────────────────────────────────┤
│  dao/         数据访问（GoFrame gdb）             │
│  ├── Users / Items / Shares / ItemMembers        │
│  ├── Employees / Jobs / Tasks / Secrets           │
│  └── 内部生成 entity/do 模型                      │
├──────────────────────────────────────────────────┤
│  pkg/shamir/v3/  Shamir SS 密码学基元             │
└──────────────────────────────────────────────────┘
```

### 路由结构

| 路由组 | 前缀 | 中间件 | 功能 |
|--------|------|--------|------|
| `/v1` | 公开 API | `MiddlewareHandlerResponse` | 用户注册、登录、健康检查、关于信息 |
| `/v1/protected` | 受保护 API | `InjectCookieIntoContext` + `MiddlewareHandlerResponse` | 文件 CRUD、份额管理、审计审批、注销 |
| `/v2` | 用户管理 V2 | `MiddlewareHandlerResponse` | 用户列表查询 |

### 核心数据流

**上传流程：**

```
1. 客户端上传文件
2. 服务端生成随机 AES-256 密钥
3. AES-GCM 加密文件 → 保存 .enc 文件
4. Shamir SS 将密钥分割为 3 份份额（2-of-3 门限）：
   - Device Share（坐标 = murmur3(userId)）→ 返回客户端
   - Auth Share（随机坐标）→ Base64 编码后存入 shares 表
   - Recovery Share（随机坐标）→ 生成 Recovery Code 并哈希存储
5. 加密文件记录写入 items 表
```

**下载流程：**

```
1. 客户端提交 Device Share
2. 服务端查询 shares 表获取 Auth Share
3. Base64 解码 → Lagrange 插值恢复 AES 密钥
4. AES-GCM 解密 .enc 文件到临时目录
5. 流式传输给客户端
6. defer 删除临时文件 + 置零密钥内存
```

**份额刷新流程（PSS）：**

```
1. 用户提交 Device Share（或 Recovery Code）
2. 服务端取出 Auth Share
3. 重建原始 AES 密钥
4. 用新的随机坐标重新分割密钥
5. 更新 Auth Share + Recovery Share → 存入数据库
6. 为已审批的成员缓存新的 Device Share
7. 通过 SSE 向前端推送进度
```

---

## 技术栈

### 后端

| 层面 | 技术 | 说明 |
|------|------|------|
| Web 框架 | **GoFrame v2.10.2** | 国产全栈框架，提供路由、ORM、配置、OpenAPI |
| 数据库 | **PostgreSQL** | 通过 `gogf/gf/contrib/drivers/pgsql/v2` 驱动 |
| ORM | **GoFrame gdb** | 内置于 GoFrame，DAO 模式 |
| 认证 | **JWT** (`golang-jwt/jwt/v5`) | 支持 HS256/HS384/HS512/RS256/RS384/RS512/ES256/ES384/ES512/PS256/PS384/PS512/EdDSA |
| 密码哈希 | **Argon2** (`pilinux/argon2`) | 64MB Memory, 3 Iterations, 2 Parallelism |
| 对称加密 | **AES-256-GCM** (Go 标准库 `crypto/aes`) | 随机 Nonce (12B) + 随机 AAD (32B) |
| 秘密共享 | **自研 Shamir SS** (`pkg/shamir/v3`) | GF(2^32-5) 有限域 |
| 用户坐标 | **Murmur3** (`spaolacci/murmur3`) | 从用户 ID 生成确定性的份额坐标 |
| 可观测性 | **OpenTelemetry OTLP** | 分布式追踪 |
| 语言版本 | **Go 1.25** | |

### 算法验证（exploration）

| 层面 | 技术 |
|------|------|
| 语言 | Python 3.x |
| 包管理 | UV (`uv.toml`) |
| 图形处理 | Pillow |
| 科学计算 | NumPy |

### 前端

| 层面 | 技术 |
|------|------|
| 框架 | **Flutter 3.11** |
| 状态管理 | `flutter_riverpod ^3.3.1` |
| 路由 | `go_router ^17.3.0` |
| HTTP | `http ^1.6.0` |
| 桌面支持 | `window_manager`, `desktop_drop`, `file_picker`, `tray_manager` |
| 本地存储 | `hive`, `flutter_secure_storage` |
| 密码学 | `pointycastle ^3.9.1` |

> **选型原因**：不采用 Vue3/React 技术栈，避免开发时携带至少 1.5GB 的 Node.js 运行时占用。抛开 Material Design 与 Windows 原生 UI 风格差异不提，Flutter 本身也具有跨端开发的便利性。

### 密码学算法

| 算法 | 用途 | 参数 |
|------|------|------|
| Shamir SS | AES 密钥分割与恢复 | GF(2^32-5) 域，Prime = 4294967291，2-of-3 门限 |
| AES-256-GCM | 文件对称加密 | 32 字节随机密钥，12 字节 Nonce，32 字节 AAD |
| Argon2 | 用户密码哈希 | Memory 64MB, Iterations 3, Parallelism 2, Salt 16B, Key 32B |
| Murmur3 | 从用户 ID 生成份额坐标 | 32-bit FNV-1a 变体 |
| JWT | 用户认证 | 默认 HS256，可配置非对称算法 |

---

## 项目不足

### 1. ORM 与原生 SQL 混用

简单 CRUD 使用 GoFrame DAO Model 风格，复杂多表 JOIN 和聚合查询则直接手写原始 SQL，两种风格交织在同一控制器文件中。

**示例**：[item_v1_item_list.go](backend/internal/controller/item/item_v1_item_list.go)

```go
// ORM 风格 —— 简单查询
err = dao.Users.Ctx(ctx).Where("username", req.Username).Scan(user)

// 原生 SQL —— 在同一个文件中不远处
const queryAll = `select distinct ... from public.items i
    left join public.users ownerUsers on ownerUsers.id = i.owner_id
    left join public.users uploaderUsers on uploaderUsers.id = i.uploader_id
    left join public.item_members im on i.id = im.item_id
    where i.is_public=true or i.owner_id = ?`
```

**影响**：维护时需要同时理解 ORM 和 SQL 两种查询方式，新增查询时缺乏统一的风格指导。虽然所有原生 SQL 均使用 `?` 占位参数化（无 SQL 注入风险），但风格不一致降低了代码可读性。

### 2. 多表 JOIN 手写查询

多处业务查询手写复杂的多表 LEFT JOIN + 聚合 SQL，不仅冗长，而且跨表关联逻辑分散在控制器中难以复用。

| 位置 | SQL 复杂度 |
|------|-----------|
| [user_v2_get_list.go](backend/internal/controller/user/user_v2_get_list.go#L20-L27) | `employees` ⟕ `users` ⟕ `jobs` |
| [audit_v1_audit_list.go](backend/internal/controller/audit/audit_v1_audit_list.go#L22-L41) | `items` ⟕ `item_members` ⟕ `users` ⟕ `shares` + `array_agg` × 10 |
| [share_v1_share_refresh.go](backend/internal/controller/share/share_v1_share_refresh.go#L44-L78) | 3 条独立的多表 LEFT JOIN 查询 |
| [item_v1_item_list.go](backend/internal/controller/item/item_v1_item_list.go#L19-L54) | `items` ⟕ `users` × 2 ⟕ `item_members`（含计数查询） |

**影响**：SQL 字符串硬编码在控制器层，无法被 DAO 层封装复用，也难以进行单元测试。

### 3. 手工依赖注入

通过函数式选项模式（Functional Options）在 [cmd.go](backend/internal/cmd/cmd.go#L103-L156) 路由注册处手动装配所有依赖：

```go
// 每个控制器手动传入配置选项
itemctrl.NewV1(
    itemctrl.WithArgonConfig(&mainConfig.Server.Argon),
    itemctrl.WithCryptoConfig(&mainConfig.Item),
)
auditctrl.NewV1(
    auditctrl.WithArgonConfig(&mainConfig.Server.Argon),
    auditctrl.WithCryptoConfig(&mainConfig.Item),
)
share.NewV1(
    share.WithArgonConfig(&mainConfig.Server.Argon),
    share.WithCryptoConfig(&mainConfig.Item),
)
```

认证 Handler 和中间件同样在路由注册处手动构建工具实例：

```go
group.POST("/auth/login", auth.PortableLoginHandler(
    &mainConfig.Server.Argon, &mainConfig.Server.Cookie,
))
group.Middleware(auth.InjectCookieIntoContext(&mainConfig.Server.Cookie))
```

**影响**：没有统一的 DI 容器，依赖关系散落在路由注册代码中。随着控制器增多，`cmd.go` 持续膨胀，新增控制器需要同时修改路由注册和依赖注入两处。

### 4. 无 TOKEN 吊销机制

JWT 签发后无法强制撤销。`/auth/logout` 仅清除客户端 Cookie，服务端不维护任何黑名单或白名单。

**影响**：即使用户主动注销，已签发的 Token 在过期前仍然有效。这是课设项目的已知局限——业务表单没有引入任何缓存机制（如 Redis），无法实现 Token 黑名单。

### 5. Auth Share 加密被注释

[crypto-kits.go](backend/internal/logic/crypto-kits.go) 中 4 个份额加解密方法均直接返回原文，**这是一个因底层基础设施限制的无奈之举**。

[issues.md](backend/issues.md) 记录了一个严重的 PostgreSQL bytea 静默比特翻转问题：

```
存盘前（校验和 3419171272）：
    ...92e...
存入 bytea 后再取出（校验和 2977688604）：
    ...92W...   ← Base64 字符 e(011110) 被静默翻转成 W(010110)
```

多组数据测试表明，PostgreSQL bytea 列在写入/读取过程中存在偶发的物理比特翻转，导致 AES-GCM 密文无法通过认证标签校验（GCM 对任何比特篡改都会拒绝解密）。因此不得不**暂时移除 AES-GCM 加密流程**，将 Auth Share 以明文 Base64 形式存储。

```go
func (cu *CryptoUtils) EncryptAuthShare(ctx context.Context, share []byte, autoClear ...bool) ([]byte, error) {
    // 先给注销了 — 因为 bytea 比特翻转导致 GCM 认证失败
    return share, nil   // ← L317: 返回明文
}
```

**当前规划**：改用 [XXTEA](https://en.wikipedia.org/wiki/XXTEA) 对份额坐标本身进行加密（XXTEA 对单比特翻转不敏感 —— 解密后只有对应比特位出错而非整块拒绝），而不是用 AES-GCM 包裹整个份额载荷。但由于项目已停止维护，上述方案停留在规划阶段未落地。

### 6. 部分功能未实现

- `ItemRecover`（通过 Recovery Code 恢复份额）：返回 `NotImplemented`
- `ItemSubmissionConfirm`（审核通过申请）：返回 `NotImplemented`

### 7. 安全审查发现

通过 TRAE-security-review 审计 `backend/` 代码，发现以下补充问题：

| 问题 | 严重度 | 位置 | 说明 |
|------|--------|------|------|
| JWT `none` 算法入口 | 中 | [cookie-kits.go:L94](backend/internal/logic/cookie-kits.go#L94) | `str2SigningMethod` 映射表包含 `"none": jwt.SigningMethodNone`，若配置文件误设为 `none` 将导致 Token 可被伪造 |
| Span 泄露主密钥 | 中 | [crypto-kits.go:L335](backend/internal/logic/crypto-kits.go#L335) | OpenTelemetry Span 属性中包含 `master_key.base64encode`，代码中已标注 TODO 移除 |

---

## 项目值得肯定的地方

### 1. 系统的算法预研体系

`exploration/` 目录包含 7 个层层递进的算法验证子项目，从实数域教学验证到 GF(2^32-5) 生产级参考实现，每个阶段都有完整的代码实现、测试套件和 `development.md` 开发文档。这种"先验证、后工程"的方法为后端 Shamir SS 实现提供了坚实的理论基础。

### 2. 自研 Shamir SS 算法实现

[`pkg/shamir/v3/`](backend/pkg/shamir/v3/) 在 GF(2^32-5) 有限域上完整实现了：

- `Split()` — 秘密分割（`crypto/rand` 安全随机多项式系数）
- `Recover()` — Lagrange 插值恢复
- `GenerateDelta()` / `ApplyDelta()` — 零值扰动向量（Proactive Secret Sharing）
- `GenerateSingleShare()` — 单份额动态生成
- PKCS#7 风格 4 字节对齐 Padding

所有运算使用 `uint64` 中间存储避免溢出，模运算采用安全素数 `2^32 - 5 = 4294967291`。

### 3. AES-GCM 加密流程与内存安全

- 每次加密使用**随机 Nonce + 随机 AAD**，避免 Nonce 重用
- 加密文件以 UUID 命名，附带魔术字符 `"ITEM"`
- 解密后的临时文件通过 `defer` 闭包保证**立即删除**
- 密钥和明文使用后通过 `Memclr()` 主动**遍历置零**，降低内存泄露风险

### 4. JWT 认证与水平越权防护

- 中间件 `InjectCookieIntoContext` 统一校验 JWT，解析 Claims 注入请求 Context
- 所有数据修改操作在 SQL 中通过 `owner_id = ?` 参数进行**水平越权检查**
- 支持 Cookie 和 Authorization Header 两种 Token 携带方式
- 注销操作清除服务端 Cookie

### 5. 数据库事务操作

关键数据写入（如文件上传、份额刷新）使用 `g.DB().Transaction()` 保证原子性：

```go
// 上传文件 → 写入条目 + 写入 Auth Share + 写入 Recovery Share 在同一事务中
err = g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error { ... })
```

### 6. 可观测性

OpenTelemetry Span 覆盖了几乎所有关键操作路径——从 HTTP 请求入口、数据库查询、加密操作到文件 I/O，每个阶段都有明确的 Span Event 和状态记录。

### 7. SSE 流式响应

份额刷新接口（ShareRefresh）使用 **Server-Sent Events** 向前端实时推送操作进度（百分比 + 阶段描述），提升了长耗时操作的用户体验。

### 8. Agent 协作开发模式

项目中 `exploration/` 和 `frontend/` 均采用**"人工制定 Goal + Agent 全流程编写、维护和审查"**的开发模式——人类负责架构决策和技术方向，Agent 负责具体代码实现、测试和文档编写。这种模式在课程设计场景下高效地完成了算法验证和前端开发工作。

### 9. Flutter 轻量级跨端方案

前端选用 Flutter 而非 Vue3/React，避免了至少 1.5GB 的 Node.js 运行时占用，同时具备了 Windows / macOS / Linux / Web 跨端部署的便利性。

---

## 快速开始

### 环境要求

| 依赖 | 版本要求 |
|------|---------|
| Go | ≥ 1.25 |
| Python | ≥ 3.x（仅 exploration 需要） |
| PostgreSQL | 任意版本 |
| Flutter SDK | ^3.11.5 |
| UV (Python 包管理) | 最新版（仅 exploration 需要） |

### 启动 exploration 算法验证（可选）

```bash
# 进入任一算法验证子项目
cd exploration/shamir-real          # 实数域 SSS
cd exploration/shamir-galois        # GF(257) 有限域
cd exploration/shamir-uint32-array-input  # GF(2^32-5) 完整实现

# 安装依赖并运行
uv run main.py
```

### 启动后端

1. 创建 PostgreSQL 数据库并初始化表结构

2. 配置 `backend/.env` 文件（参考 `.env.develop`）：

```env
# 数据库连接
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=your_password
DB_NAME=crypto_course
```

3. 配置 `backend/manifest/deploy/kustomize/overlays/develop/configmap.yaml` 中的运行时参数

4. 启动服务：

```bash
cd backend
go run main.go
```

服务默认监听 `:8000`，Swagger 文档位于 `/swagger`。

### 启动前端

```bash
cd frontend
flutter pub get
flutter run -d windows
```

---

## 开发模式

| 模块 | 开发方式 |
|------|----------|
| `exploration/` | 人工制定 Goal + Agent 全流程编写、验证和维护 |
| `frontend/` | 人工制定 Goal + Agent 全流程编写、维护和审查 |
| `backend/` | 人工开发（唯一不由 Agent 负责的部分） |

每个 Agent 生成的子项目都包含 `development.md`（开发文档）和 `task.md`（Agent 任务指令），记录了完整的需求到实现链路。

---

## 许可证

本项目为密码学课程设计作品。
