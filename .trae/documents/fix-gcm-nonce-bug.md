# 修复：GCM 加密时 Nonce 被覆盖的问题

## 问题分析

在 `EncryptAuthShare` 和 `EncryptRecoveryShare` 中，`gcm.Seal` 的调用方式存在严重缺陷：

```go
ciphertext = gcm.Seal(copiedNonce, nonce, share, aad)
```

**问题原因：**
- `copiedNonce` 是长度为12的字节切片
- `gcm.Seal(dst, nonce, plaintext, aad)` 会将加密结果写入 `dst`，**覆盖**其原有内容
- 由于 `copiedNonce` 长度只有12，加密后的数据会覆盖掉 nonce 的值
- 最终 ciphertext 的前12字节不再是原始 nonce，而是加密数据的一部分

**后果：**
- 解密时从 ciphertext 前端提取的 "nonce" 实际上是加密数据的一部分
- 解密时使用的 nonce 与加密时使用的 nonce 完全不同
- GCM 认证必定失败：`cipher: message authentication failed`

## 修复方案

将 `Seal` 的调用改为：
```go
encrypted := gcm.Seal(nil, nonce, share, aad)  // 让 Seal 分配新空间
ciphertext = append(copiedNonce, encrypted...)   // 将加密数据追加到 nonce 后面
ciphertext = append(ciphertext, aad...)
```

## 影响范围

- `EncryptAuthShare`
- `EncryptRecoveryShare`

## 修复步骤

1. 修改 `crypto-kits.go` 中两个加密函数
2. 运行测试验证修复
3. 提交修复
