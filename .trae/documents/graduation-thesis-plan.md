# 毕设论文撰写计划

## 任务概述

根据 `task.md` 的要求，撰写一篇以**毕设**为背景的 Markdown 文档，覆盖完整的论文结构，并满足附录中的所有补充要求。

## 输出文件

在项目根目录创建 `毕业论文.md`。

---

## 文档结构（来自 task.md）

```
## 绪论
### 国内外研究现状
### 主要内容
### 组织结构
## 相关技术
## 系统分析
## 需求分析
### 可行性分析
### 系统功能模块分析
## 总体设计
## 具体实现
## 软件测试
## 总结与展望
### 总结
### 展望
```

---

## Phase 1: 素材搜集（已完成大部分）

### 已读取的关键文件

| 文件 | 用途 |
|---|---|
| `task.md` | 文档结构、附录要求 |
| `backend/issues.md` | 比特翻转BUG → 移除AES加密的原因 |
| `frontend/.trae/documents/书写风格.md` | 写作风格规范（12条核心规则+9条观察规则） |
| `backend/internal/cmd/cmd.go` | 路由注册、中间件、JWT认证、OTLP链路追踪 |
| `backend/pkg/shamir/v3/shamir.go` | v3 简化版 Shamir 实现（Split/Recover/Delta） |
| `backend/pkg/shamir/v2/shamir.go` | v2 完整版 Shamir 实现（需外部seed） |
| `backend/pkg/shamir/v3/field.go` | GF(2^32-5) 域运算（add/sub/mul/pow/inv/div/lagrange） |
| `exploration/自拟题目.md` | 题目定义、8个具体要求 |

### 仍需读取的关键文件

| 文件 | 用途 |
|---|---|
| `backend/internal/logic/crypto/v2/crypto-kits-v2.go` | 具体实现节：加密工具 |
| `backend/internal/logic/item-kits.go` | 具体实现节：条目业务逻辑 |
| `backend/internal/logic/crypto-kits.go` | 具体实现节：主加密逻辑 |
| `backend/internal/controller/item/*.go` | 具体实现节：条目控制器 |
| `backend/internal/controller/auth/*.go` | 具体实现节：认证控制器 |
| `backend/internal/controller/share/*.go` | 具体实现节：份额控制器 |
| `backend/internal/controller/audit/*.go` | 具体实现节：审计控制器 |
| `backend/pkg/shamir/v2/field.go` | 具体实现节：v2域运算对比 |
| `frontend/lib/features/shares/` 相关文件 | 具体实现节：前端份额加密/存储 |
| `frontend/lib/features/items/` 相关文件 | 具体实现节：前端条目模型 |
| `exploration/shamir-go/` 关键文件 | 相关技术节：Shamir算法探索历程 |
| `exploration/shamir-galois/` 关键文件 | 相关技术节：伽罗瓦域验证 |
| `.trae/documents/security-review-*.md` | 软件测试节：安全审查发现 |
| `项目详述.md` | 各节补充信息 |

---

## Phase 2: 各节撰写计划

### 2.1 绪论

**国内外研究现状**：
- 秘密共享概念（Shamir 1979, Blakley 1979）
- 门限密码学在企业场景的应用
- privy.io 的 2-of-3 架构参考
- 移动对手攻击与主动安全

**主要内容**：
- 基于 `自拟题目.md` 的8个具体要求提炼

**组织结构**：
- 按章节简述

### 2.2 相关技术

- GoFrame 框架
- Flutter 跨平台桌面开发
- Shamir 秘密共享算法（GF(2^32-5) 有限域）
- 拉格朗日插值
- JWT 认证
- PostgreSQL
- OTLP 链路追踪
- 零值扰动多项式（动态成员管理）
- FNV-1a 哈希（用户X坐标生成）
- 探索历程：`exploration/` 目录中的算法演进

### 2.3 需求分析 & 系统分析

**可行性分析**：
- 技术可行性（Go + Flutter 成熟技术栈）
- 算法可行性（Shamir 在 GF(2^32-5) 上已验证）
- 操作可行性

**系统功能模块分析**：
- 用户管理模块：`backend/internal/controller/user/`
- 认证模块：`backend/internal/controller/auth/`
- 条目管理模块：`backend/internal/controller/item/`
- 份额管理模块：`backend/internal/controller/share/`
- 审计模块：`backend/internal/controller/audit/`
- 文件监控模块：`backend/internal/module/watcher/`
- 前端对应 features：`frontend/lib/features/`

### 2.4 总体设计

- 系统架构图（文字描述）
- 技术选型理由
- 模块划分
- 数据流设计
- API 路由设计（来自 `cmd.go`）：`/v1`、`/v1/protected`、`/v2`

### 2.5 具体实现（重点章节）

**必须包含源代码引用**，从以下文件中提取：

1. **Shamir 有限域实现**
   - `backend/pkg/shamir/v3/field.go`：add/sub/mul/pow/inv/div、evalPolynomial、lagrangeInterpolation
   - `backend/pkg/shamir/v2/field.go`：对比 v2 与 v3 的差异

2. **秘密拆分与恢复**
   - `backend/pkg/shamir/v3/shamir.go`：Split/Recover
   - `backend/pkg/shamir/v2/shamir.go`：v2 的 Split（需seed）对比

3. **零值扰动与动态成员管理**
   - `backend/pkg/shamir/v3/shamir.go`：GenerateDelta/ApplyDelta
   - `backend/pkg/shamir/v2/shamir.go`：CalculateSingleDelta

4. **后端控制器层**
   - `backend/internal/controller/item/`：条目CRUD、提交、审批
   - `backend/internal/controller/auth/`：JWT登录/注销/中间件
   - `backend/internal/controller/share/`：份额列表/拉取/刷新

5. **前端核心实现**
   - `frontend/lib/features/shares/`：份额存储/加密（Hive）
   - `frontend/lib/features/items/`：条目数据模型
   - `frontend/lib/router/app_router.dart`：GoRouter 路由设计

6. **安全检查：Sensitive Span**
   - 从链路追踪代码中找出敏感 Span（在 controller 或 logic 层的调试日志）

### 2.6 软件测试

- **链路追踪**：OTLP HTTP 集成（`cmd.go` 中的 otlphttp.Init）
- **单元测试**：`backend/pkg/shamir/v2/field_test.go`、`backend/pkg/shamir/v3/shamir_test.go`
- **集成测试**：`backend/tests/` 目录
- **安全测试**：结合 `.trae/documents/security-review-*.md` 中的安全审查发现

### 2.7 总结与展望

**总结**：
- 项目实现了什么
- 技术亮点
- 项目不足：
  - 前端序列化不一致（json_serializable vs switch-case）
  - 链路追踪中未移除的调试 Span（敏感信息泄露风险）
  - AES 加密移除后的安全降级

**展望**：
- XXTEA 保格式混淆（对份额坐标进行格式保留加密）
- 完善前端序列化统一方案
- 移除敏感调试 Span
- 加入真正的 AES-GCM 份额加密

---

## Phase 3: 安全审查（使用 TRAE-security-review 技能）

虽然 TRAE-security-review 技能主要面向 diff 审查，但我们可以对以下关键安全面进行分析并融入论文：

1. **硬编码凭据检查**：搜索代码中的密钥、token 是否硬编码
2. **CSPRNG 使用**：确认 v3 的 `crypto/rand` 使用正确、v2 的 HMAC-SHA256 种子派生安全
3. **JWT 安全**：检查 JWT secret 来源、过期时间、中间件覆盖范围
4. **输入验证**：检查控制器层的参数校验
5. **日志安全**：检查链路追踪中的敏感 Span 是否泄露密钥信息
6. **安全降级评估**：AES 加密移除后的风险分析

---

## Phase 4: 执行步骤

1. 读取所有仍然需要的源文件（约15个文件）
2. 对项目执行安全审查（Grep 搜索硬编码、CSPRNG、JWT、日志等模式）
3. 按节撰写文档，每节完成后检查书写风格
4. 整合安全审查发现到"软件测试"和"项目不足"
5. 最终通读检查

---

## 书写风格注意事项（来自书写风格.md）

- **RULE-01**：不假定读者知识 → 解释 Shamir、GF(2^32-5)、拉格朗日插值等概念
- **RULE-02**：用主动语态 → "我们设计了…"而非"设计被…"
- **RULE-03**：用具体术语 → 引用具体文件路径和函数名
- **RULE-04**：删除多余词 → 避免"值得注意的是"等
- **RULE-12**：拆分长句 → 每句不超过30词
- **RULE-H**：事实性声明需引用 → 引用项目中具体代码行

---

## 关键决策

1. **输出文件**：项目根目录 `毕业论文.md`
2. **语言**：中文（符合毕设惯例）
3. **安全审查范围**：全项目扫描（非diff），融入论文而非独立报告
4. **代码引用格式**：使用 `file:///` 链接引用源码位置
5. **展望内容**：XXTEA 保格式混淆（项目无代码体现，仅放展望）
