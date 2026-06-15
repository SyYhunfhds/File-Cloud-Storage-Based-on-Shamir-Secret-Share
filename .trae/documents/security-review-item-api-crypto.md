# 安全审查：条目提交/下载 API 对 CryptoUtils V2 的调用封装

## 审查范围

- `item_v1_item_submit.go` — 文件上传加密 + 份额生成
- `item_v1_item_download.go` — 份额恢复 + 文件解密下载
- `crypto-kits.go`（V2 版本，含 `ctx` 参数和追踪埋点）— 底层加解密实现

## 摘要

底层 AES-GCM 加解密逻辑（`EncryptAuthShare`/`DecryptAuthShare`）本身正确，两对函数的输入输出格式一致（`nonce || encrypted || tag || aad`）。但 API 层封装存在若干问题，其中 **未检查 `EncryptAuthShare` 的返回值** 是导致数据入库损坏、解密端无法恢复的根因。

---

## 发现的问题

### 🔴 问题 1：`EncryptAuthShare` 的错误被 `EncryptRecoveryShare` 静默覆盖

| 项目 | 值 |
|------|-----|
| **类别** | `crypto & secret handling` |
| **严重程度** | HIGH |
| **置信度** | 0.95 |
| **位置** | [`item_v1_item_submit.go#L108-L110`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/controller/item/item_v1_item_submit.go#L108-L110) |

**证据：**

```go
encryptedAuthShare, err := crypU.EncryptAuthShare(ctx, authShare, false)   // L108: err 被声明
// 随机密钥加密恢复份额, 并返回密文和32字节密钥
encryptedRecovery, recoveryCode, err := crypU.EncryptRecoveryShare(recoveryShare, nil, false)  // L110: err 被覆盖
recoveryCodeHash, err := hu.HashGen(recoveryCode)                            // L111: 再次被覆盖
if err != nil { // 份额生成失败                                              // L112: 只检查了 HashGen 的错误
```

`EncryptAuthShare` 的返回值 `err` 在 L110 被 `:=` 覆盖。Go 的 `:=` 在至少声明一个新变量（此处为 `recoveryCode`）时，会复用作用域内的已有变量 `err`。因此 L108 的 `err` 被 L110 的结果无条件覆盖。如果 `EncryptAuthShare` 加密失败（无论何种原因），`err` 会被 `EncryptRecoveryShare` 的结果擦除，错误被静默吞没。

同时，`encryptedAuthShare` 在 `EncryptAuthShare` 失败时可能为空或部分数据，但 L150 直接入库：
```go
AuthShare: encryptedAuthShare,
```

**影响：** 数据库中存储了损坏的 Auth Share 密文。下载端从 DB 读取后调用 `DecryptAuthShare` 必然 GCM 认证失败，且追踪日志中「加密/解密两侧的 ciphertext 一模一样（都是损坏的数据）」，与用户描述的「哪怕 Span属性记录得一模一样，这个Auth Share 也解密不出来」完全吻合。

**建议：** 每个 `Encrypt*` 调用后立即检查 `err`，失败时回滚事务并返回错误，不应继续执行。

---

### 🔴 问题 2：服务端主密钥以 base64 明文写入追踪 Span

| 项目 | 值 |
|------|-----|
| **类别** | `sensitive_data_exposure` |
| **严重程度** | HIGH |
| **置信度** | 0.95 |
| **位置** | [`crypto-kits.go#L225-L227`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L225-L227) |
| | [`crypto-kits.go#L289-L291`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L289-L291) |

**证据：**

```go
// EncryptAuthShare (L225-227):
attribute.Int("crypto_utils.master_key.length", len(cu.ao.key)),
attribute.String("crypto_utils.master_key.base64encode", gbase64.EncodeToString(cu.ao.key)),

// DecryptAuthShare (L289-291):
attribute.Int("crypto_utils.master_key.length", len(cu.ao.key)),
attribute.String("crypto_utils.master_key.base64encode", gbase64.EncodeToString(cu.ao.key)),
```

`cu.ao.key` 是 AES-GCM 对称主密钥，用于加密/解密所有 Auth Share。两端完整写入 Span 属性。任何能访问可观测性系统（Jaeger、Datadog、Cloud Trace 等）的人员均可直接获取该密钥，进而解密全库 Auth Share。

**影响：** 全库 Auth Share 机密性丧失，所有用户条目的文件密钥均可被恢复。

**建议：** 从生产环境的 Span 中移除 `crypto_utils.master_key.*` 属性。已有的 TODO 注释（`// TODO: 移除高危Span`）应尽快落实。

---

### 🟡 问题 3：`DecryptAuthShare` 初始 Span 属性引用了未赋值的返回值

| 项目 | 值 |
|------|-----|
| **类别** | `sensitive_data_exposure`（低危，主要导致日志误导） |
| **严重程度** | LOW |
| **置信度** | 0.95 |
| **位置** | [`crypto-kits.go#L287-L288`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L287-L288) |

**证据：**

```go
func (cu *CryptoUtils) DecryptAuthShare(ctx context.Context, ciphertext []byte, autoClear ...bool) (share []byte, err error) {
    // ...span 创建...
    span.SetAttributes(
        attribute.Int("auth_share.length", len(share)),         // share 是 nil!
        attribute.String("auth_share.base64encode", gbase64.EncodeToString(share)),  // 空字符串!
```

`share` 是函数的**命名返回值**，函数入口处值为 `nil`。此处应使用输入参数 `ciphertext`。该问题不影响解密逻辑，但导致 Span 中 `auth_share.length` 恒为 0、`auth_share.base64encode` 恒为空，给调试带来误导。

**建议：** 将前两个 `attribute` 的取值源从 `share` 改为 `ciphertext`。

---

### 🟡 问题 4：DeviceShare 前 4 字符的越界风险

| 项目 | 值 |
|------|-----|
| **类别** | `untrusted_input_handling` |
| **严重程度** | LOW |
| **置信度** | 0.85 |
| **位置** | [`item_v1_item_download.go#L96`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/controller/item/item_v1_item_download.go#L96) |

**证据：**

```go
attribute.String("request.parameter.share.first_4_chars", req.DeviceShare[:4]),
```

如果 `req.DeviceShare` 长度小于 4（恶意请求或校验疏忽），此处会 panic: `slice bounds out of range`。虽然 `v1.ItemDownloadReq` 的 `DeviceShare` 应该有 `v:"required"` 校验，但 GoFrame 的参数校验在解析阶段执行，理论上不会让空值到达此处。但长度不足 4 的 base64 字符串（如 `"ab"`）仍可通过校验。

**建议：** 切取前使用 `len(req.DeviceShare)` >= 4 判读，或使用 `req.DeviceShare[:min(4, len(req.DeviceShare))]`。

---

## 影响链路分析

```
提交API:                                          下载API:
───────                                           ───────
SplitShare(key)                                   
  ↓ authShare (Base64 JSON)                       DB → auth_share (损坏的密文)
  ↓                                                 ↓
EncryptAuthShare(ctx, authShare, false)           DecryptAuthShare(ctx, dbBytes, false)
  ↓ error ❌ ← 被 L110 覆盖                         ↓ GCM认证失败 ❌
  ↓ encryptedAuthShare (可能为空/损坏)               ↓ 返回错误
  ↓ DB → Save(auth_share=encryptedAuthShare)       ↓ "服务端认证份额已损坏"
```

关键断裂点：[问题1](item_v1_item_submit.go#L108-L110)。`EncryptAuthShare` 的返回值错误未传播，导致后续逻辑使用不可靠的 `encryptedAuthShare`。

---

## 无关本报告的已修复问题

- `EncryptAuthShare`/`EncryptRecoveryShare` 的 AAD 未追加到密文尾部 —— 上一轮审查已修复。
