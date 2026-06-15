# 安全审查：AES-GCM 份额加密实现

## 摘要

审查 `crypto-kits.go` 中四个 AES-GCM 实现（Auth Share / Recovery Share 的加密和解密）以及两个委托方法（EncryptMemberShare / DecryptMemberShare）。

**结论：`EncryptAuthShare ↔ DecryptAuthShare` 和 `EncryptRecoveryShare ↔ DecryptRecoveryShare` 两对存在相同的结构化缺陷，导致解密必定失败。**

---

## 数据格式对比

### `SymmetricEncrypt` / `SymmetricDecrypt`（正确的参考实现）

```
加密: nonce(12) + gcm.Seal(nonce, nonce, plaintext, cu.ao.additionalBytes)
      ↓ 输出: nonce || encrypt(plaintext) || tag(16)
      (AAD = cu.ao.additionalBytes，固定不变，不写入输出)

解密: nonce = ciphertext[:12]
      plaintext = gcm.Open(nil, nonce, ciphertext[12:], cu.ao.additionalBytes)
      (使用相同的固定 AAD)
```

**正确。** AAD 是固定值（`cu.ao.additionalBytes`），加解密两端一致。

### `EncryptAuthShare` / `DecryptAuthShare`（有缺陷）

```
加密: nonce = grand.B(12)
      aad = grand.B(32)   ← 随机生成
      ciphertext = gcm.Seal(nonce_copy, nonce, share, aad)
      ↓ 输出: nonce || encrypt(share) || tag(16)
      (AAD 未写入输出)
      Memclr(aad)          ← AAD 被清零

解密: nonce = ciphertext[:12]
      aad = ciphertext[len-32 : len]   ← 期望从尾部读取 AAD
      data = ciphertext[12 : len-32]
      share = gcm.Open(nil, nonce, data, aad)
```

**缺陷：加密端 `gcm.Seal` 只输出 `nonce || encrypted || tag`，AAD 未拼接到输出尾部。但解密端从 `ciphertext[len-32:len]` 读取 AAD——这个位置实际是 encrypted data 末尾的 32 字节。**

结果：解密时传给 `gcm.Open` 的 AAD 与加密时使用的随机 AAD 不一致，GCM 认证必然失败。

### `EncryptRecoveryShare` / `DecryptRecoveryShare`（相同缺陷）

逻辑与 Auth Share 对完全一致，存在相同问题。

### `EncryptMemberShare` / `DecryptMemberShare`

两者只是循环委托 `EncryptAuthShare` / `DecryptAuthShare`，继承相同的缺陷。

---

## 发现的问题

### 🔴 问题 1: Auth Share 加解密 AAD 存储不一致

| 项目 | 值 |
|------|-----|
| **类别** | `weak_crypto` |
| **严重程度** | HIGH |
| **置信度** | 0.95 |
| **位置** | [`crypto-kits.go#L218-L224`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L218-L224) (加密) |
| | [`crypto-kits.go#L255-L268`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L255-L268) (解密) |

**证据:**

加密端（[L218-L224](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L218-L224)）：
1. 生成随机 AAD 32 字节
2. `gcm.Seal(copiedNonce, nonce, share, aad)` — AAD 仅用于认证计算，**不写入输出**
3. 输出格式：`nonce(12) || encrypted || tag(16)`，共 28 + len(share) 字节
4. `Memclr(aad)` — 清空 AAD

解密端（[L255-L268](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L255-L268)）：
1. 假设 `ciphertext` 格式为 `nonce(12) || encrypted || tag(16) || aad(32)`
2. 从 `ciphertext[len-32:len]` 截取「AAD」
3. 这个位置实际是 encrypted data 末尾的字节，**不是真正的 AAD**

**影响:** 传入 `gcm.Open` 的 AAD 与加密时使用的 AAD 不同，GCM 认证标签校验失败，解密必定返回错误。

**建议:** 加密时 `Seal` 完成后，将 AAD 拼接到输出尾部；或改用固定 AAD 方案（如 `SymmetricEncrypt` 的做法）。

### 🔴 问题 2: Recovery Share 加解密 AAD 存储不一致

| 项目 | 值 |
|------|-----|
| **类别** | `weak_crypto` |
| **严重程度** | HIGH |
| **置信度** | 0.95 |
| **位置** | [`crypto-kits.go#L304-L309`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L304-L309) (加密) |
| | [`crypto-kits.go#L338-L351`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L338-L351) (解密) |

与问题 1 完全相同。加密时生成随机 AAD 但不写入输出，解密时从密文尾部错误地读取 AAD。

**建议:** 同上。

### 🟡 问题 3: `EncryptRecoveryShare` 参数 `key` 的无条件 Write-back

| 项目 | 值 |
|------|-----|
| **类别** | `crypto & secret handling` |
| **严重程度** | LOW |
| **置信度** | 0.85 |
| **位置** | [`crypto-kits.go#L279-L285`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L279-L285) |

```go
func (cu *CryptoUtils) EncryptRecoveryShare(share []byte, key []byte, autoClear ...bool) (ciphertext []byte, nk string, err error) {
    if len(key) < cu.ao.keyLen {
        nk = grand.S(cu.ao.keyLen)
        key = []byte(nk)  // 重新赋值局部变量 key
    } else {
        nk = string(key)  // 无条件将 key 转换为字符串
    }
```

当 `key` 长度满足要求时（例如调用方 `EncryptRecoveryShare(rs, nil)` 中的 `nil` 触发了 `len(nil) == 0 < 32` 分支），`nk = string(key)` 将字节切片转为字符串。此操作**复制**底层字节，不修改原数据，安全风险低。但如果未来调用方传入非 nil 的用户密钥，`string(key)` 会创建字符串驻留，可能延长密钥在内存中的存活时间。

**建议:** 只在需要返回新密钥时才做 `string(key)` 转换。

---

## 影响范围

问题 1 和 2 直接影响 `share_v1_share_refresh.go` 中的调用链路：

```
EncryptAuthShare  (L248) → 存入 DB
DecryptAuthShare  (L182) → 从 DB 读取后解密  ← 必定失败
EncryptRecoveryShare (L249) → 存入 DB
DecryptRecoveryShare (L135) → 从 DB 读取后解密  ← 必定失败
```

`EncryptMemberShare` / `DecryptMemberShare` 继承同样缺陷。

**后果：** 当前生产数据中所有通过 `EncryptAuthShare` / `EncryptRecoveryShare` 加密的份额均无法通过 `DecryptAuthShare` / `DecryptRecoveryShare` 解密。修复后需要重新加密已有数据。

---

## 参考：正确的对称加解密实现

`SymmetricEncrypt` / `SymmetricDecrypt`（[L128-L198](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/crypto-kits.go#L128-L198)）是正确的参考实现，原因：
- 使用固定的 `cu.ao.additionalBytes` 作为 AAD
- 加解密两端使用相同的 AAD
- AAD 不写入输出，解密时从配置读取
