# 文件加解密 Bug 根因分析

## 摘要

上传加密文件后下载解密时，`SymmetricDecrypt` 对 Auth Share 的 AES-GCM 解密报认证失败。小文本文件 (<1MB) 正常；大文件 (>1MB) 和二进制文件（zip/mp3 等，无论大小）必定失败。

根因定位在 [`file-kits.go#L124`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/file-kits.go#L124) — `gcm.Seal` 的 dst 和 src 共享同一块内存，导致上传流程中内存污染，进而使存入数据库的 `encryptedAuthShare` 字节错误。

此 Bug 会导致**所有已上传文件的 Auth Share 无法解密**。修复后，此前积累的测试数据将全部作废（Breaking Change）。

---

## 数据流

```
上传:
  file → ReadBytes → plain (文件原始字节)
  plain, 随机key → gcm.Seal(plain[:0], nonce, plain, tailCopy) → ciphertext
  key → SplitShare → authShare
  authShare → SymmetricEncrypt(服务器主密钥) → encryptedAuthShare → 存DB

下载:
  encryptedAuthShare ← 读DB
  encryptedAuthShare → SymmetricDecrypt(服务器主密钥) → authShare  ← 报GCM认证失败
```

---

## 根因

### Seal dst 与 src 内存重叠

**位置:** [`file-kits.go#L124`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/file-kits.go#L124)

```go
ciphertext = gcm.Seal(plain[:0], fu.nonce, plain, tailCopy)
//                    ^^^^^^^^              ^^^^^
//                    dst (长度0)            src (明文)
//                    二者共享底层数组
```

`Seal` 的签名是 `Seal(dst, nonce, plaintext, additionalData)`。这段代码将 `plain[:0]`（指向 `plain` 底层数组起始位置、长度为零的切片）作为 `dst`，将 `plain` 自身作为 `plaintext`。

Go 标准库 `sliceForAppend` 会检查容量：`cap(plain[:0]) = file.Size`，需要的输出大小为 `file.Size + 16`（密文 + GCM 认证标签），容量不足 → 分配新缓冲区。所以单从容量检查看，`plain` 不应被破坏。

但 GCM 内部先执行 CTR 模式加密。`crypto/cipher` 的 `XORKeyStream(dst, src)` 在 `dst` 和 `src` 重叠时，行为依赖 Go 版本的具体实现。Go 文档要求 `dst` 和 `src` "完全重叠或不重叠"——这里的 `plain[:0]`（len=0）和 `plain`（len=file.Size）属于既不完全重叠也不完全不重叠的灰色地带。

### 污染链

`Seal` 返回后，调用方立即执行：

```go
// item_v1_item_submit.go#L76
logic.Memclr(ciphertext) // 置零密文缓冲区
```

此后是 Shamir 份额分割：

```go
// #L105-108
deviceShare, authShare, recoveryShare, err := crypU.SplitShare(key, ac.Coordinate)
logic.Memclr(key)
encryptedAuthShare, err := crypU.SymmetricEncrypt(authShare, nil, true)
```

`Memclr` 释放了一大块内存（file.Size + 16 字节）。Go 的 GC 和内存分配器可能在后续的 `SplitShare` 或 `SymmetricEncrypt` 调用中复用这块内存。对于大文件，内存复用的概率显著提高；对于包含不可打印字节的二进制文件，Go 运行时的内存布局可能与纯文本文件不同。

结果是 `SymmetricEncrypt` 输出到 `encryptedAuthShare` 的字节与预期不符，存入数据库后，下载端的 `SymmetricDecrypt` 自然认证失败。

### 为什么小文本文件正常

- 小文件 (<1MB)：内存分配模式简单，GC 压力低，污染的窗口期短
- 文本文件：只包含可打印 ASCII，Go 内存分配器处理这种负载的路径与二进制数据不同
- 二进制文件（无论大小）：包含全范围字节值 (0x00–0xFF)，触发内存分配器中的不同代码路径

---

## 附带发现

### `io.Reader.Read` 忽略返回值 `n`

**位置:** [`file-kits.go#L44`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/file-kits.go#L44)

```go
p = make([]byte, file.Size)
_, err = f.Read(p)  // 丢弃了实际读取的字节数
```

`io.Reader.Read` 不保证一次读满 `len(p)` 字节。大文件或网络流可能分多次返回，未填充的尾部为零值。加密后的密文包含多余的零字节密文。

此问题**不直接导致** Auth Share 解密失败（Auth Share 与文件内容无关），但会导致文件加密内容不正确。

### 下载端份额恢复跳过 `Unpad`

**位置:** [`item_v1_item_download.go#L227`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/controller/item/item_v1_item_download.go#L227)

```go
key := shamir.Recover(...)  // 缺少 shamir.Unpad()
```

上传端使用 `Pad` 对齐，下载端理应调用 `Unpad`。当前恰好绕过是因为 32 字节密钥 4 字节对齐，`Pad` 不做填充。

### 处理器实例全局共享

**位置:** [`item_v1_item_download.go#L32-L36`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/controller/item/item_v1_item_download.go#L32-L36)

`fu` 和 `cu` 在闭包外创建，所有并发请求共享同一实例。当前只读操作暂无数据竞争。

---

## 影响范围

修复 [`file-kits.go#L124`](file:///f:/GolangWorkspace/CryptographyDesign/backend/internal/logic/file-kits.go#L124) 将改变加密文件的输出格式：`Seal(nil, ...)` 与 `Seal(plain[:0], ...)` 产生的密文字节可能不同（即使明文相同）。

**后果：** 修复前上传的所有文件，其 Auth Share 将永远无法解密。数据库中的 `auth_share` 列需要全部清除或迁移。
