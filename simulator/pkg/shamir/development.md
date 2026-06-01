# Shamir 秘密共享协议 Go 实现 - 开发文档

## 1. 项目概述

本项目实现了基于有限域的 Shamir 秘密共享协议，支持基本的秘密分割和恢复功能，以及加法同态和常数乘法同态特性。

### 1.1 核心功能
- 秘密分割（Split）：将单个秘密分割为多个份额
- 秘密恢复（Recover）：从阈值数量的份额中恢复原始秘密
- 动态重构（Reshare）：在不泄露原始秘密的情况下重新生成份额
- 加法同态（AddShares）：支持份额的加法操作
- 常数乘法同态（MultiplySharesByConstant）：支持份额与常数的乘法操作

## 2. 目录结构

```
simulator/
└── pkg/
    └── shamir/
        ├── doc.go          # 包级文档
        ├── shamir.go       # 核心实现
        └── shamir_test.go  # 单元测试
```

## 3. 设计决策

### 3.1 有限域选择
- 使用模素数的有限域 `GF(p)`，其中 `p` 是用户指定的素数
- 默认推荐素数为 257，确保足够的安全性和计算效率

### 3.2 多项式构造
- 使用随机系数生成 `t-1` 次多项式，其中 `t` 是门限值
- 多项式形式：`f(x) = secret + a_1 x + a_2 x^2 + ... + a_{t-1} x^{t-1} mod p`

### 3.3 拉格朗日插值
- 使用拉格朗日插值法从份额中恢复秘密
- 时间复杂度：O(t^2)，其中 `t` 是门限值

### 3.4 同态计算
- 加法同态：`f(x) + g(x) = (f+g)(x) mod p`
- 常数乘法同态：`c * f(x) = (c*f)(x) mod p`

## 4. 实现细节

### 4.1 核心结构体

#### `Share` 结构体
```go
type Share struct {
    X int64 // 份额的 x 坐标
    Y int64 // 份额的 y 坐标
}
```

#### `Config` 结构体
```go
type Config struct {
    Threshold int   // 门限值
    NumShares int   // 份额数量
    Prime     int64 // 有限域素数
}
```

#### `Shamir` 结构体
```go
type Shamir struct {
    field     *Field  // 有限域实例
    threshold int     // 门限值
    numShares int     // 份额数量
    coeffs    []int64 // 多项式系数
    version   int     // 版本号，用于跟踪重构
}
```

### 4.2 主要方法

#### `NewShamir` - 创建 Shamir 实例
- 验证配置参数的有效性
- 创建有限域实例
- 初始化 Shamir 结构体

#### `Split` - 分割秘密
- 生成随机多项式系数
- 计算每个 x 对应的 y 值
- 返回生成的份额

#### `Recover` - 恢复秘密
- 使用拉格朗日插值法计算 f(0)
- 返回恢复的秘密

#### `Reshare` - 动态重构
- 先恢复原始秘密
- 然后重新分割秘密
- 返回新的份额

#### `AddShares` - 加法同态
- 检查份额长度和 x 坐标是否匹配
- 对每个份额的 y 值进行加法操作
- 返回和的份额

#### `MultiplySharesByConstant` - 常数乘法同态
- 对每个份额的 y 值与常数相乘
- 返回乘积的份额

## 5. 同态计算特性

### 5.1 加法同态
- **原理**：两个秘密的份额相加，结果对应于两个原始秘密的和
- **应用场景**：安全多方计算、隐私保护数据聚合
- **使用示例**：
  ```go
  sumShares, err := shamir.AddShares(shares1, shares2)
  recoveredSum, err := shamir.Recover(sumShares[:3])
  ```

### 5.2 常数乘法同态
- **原理**：份额与常数相乘，结果对应于原始秘密与该常数的乘积
- **应用场景**：标量乘法、权重计算
- **使用示例**：
  ```go
  productShares, err := shamir.MultiplySharesByConstant(shares, 3)
  recoveredProduct, err := shamir.Recover(productShares[:3])
  ```

## 6. 测试策略

### 6.1 单元测试
- **TestNewShamir**：测试实例创建和参数验证
- **TestSplitAndRecover**：测试秘密分割和恢复
- **TestReshare**：测试动态重构功能
- **TestMultipleSecrets**：测试多个秘密的处理
- **TestLargePrime**：测试大素数支持
- **TestAdditiveHomomorphism**：测试加法同态
- **TestConstantMultiplicationHomomorphism**：测试常数乘法同态
- **TestAddSharesErrorCases**：测试错误处理

### 6.2 测试覆盖
- 核心功能测试覆盖
- 边界条件测试
- 错误处理测试
- 同态计算特性测试

## 7. 性能优化

### 7.1 计算优化
- 使用预计算的有限域操作
- 优化拉格朗日插值的计算
- 减少内存分配和复制

### 7.2 安全性考虑
- 使用密码学安全的随机数生成
- 确保有限域操作的正确性
- 验证输入参数的有效性

## 8. 应用场景

### 8.1 企业资金安全
- 多人审批机制
- 密钥分散保管
- 防止单点故障

### 8.2 安全多方计算
- 隐私保护数据聚合
- 安全投票系统
- 隐私保护机器学习

### 8.3 区块链技术
- 分布式密钥管理
- 多方签名
- 安全智能合约

## 9. 版本历史

### v1.0.0
- 基本秘密分割和恢复功能
- 动态重构支持
- 单元测试覆盖

### v1.1.0
- 添加加法同态特性
- 添加常数乘法同态特性
- 完善错误处理
- 增强测试覆盖

## 10. 未来发展方向

### 10.1 功能扩展
- 支持乘法同态
- 添加份额验证机制
- 实现阈值签名

### 10.2 性能优化
- 实现批量处理
- 优化大素数运算
- 并行计算支持

### 10.3 安全增强
- 添加份额验证
- 实现防欺诈机制
- 支持安全多方计算协议

## 11. 开发指南

### 11.1 环境要求
- Go 1.18+
- 无外部依赖

### 11.2 安装和使用

1. **安装**
   ```bash
   go get github.com/yourusername/shamir
   ```

2. **基本使用**
   ```go
   import "github.com/yourusername/shamir"

   // 创建实例
   s, err := shamir.New(
       shamir.WithThreshold(3),
       shamir.WithNumShares(5),
       shamir.WithPrime(257),
   )

   // 分割秘密
   shares, err := s.Split(42)

   // 恢复秘密
   recovered, err := s.Recover(shares[:3])
   ```

3. **同态计算**
   ```go
   // 加法同态
   sumShares, err := s.AddShares(shares1, shares2)
   sum, err := s.Recover(sumShares[:3])

   // 常数乘法同态
   productShares, err := s.MultiplySharesByConstant(shares, 3)
   product, err := s.Recover(productShares[:3])
   ```

## 12. 故障排除

### 12.1 常见问题

1. **错误："secret must be less than prime"**
   - 原因：秘密值大于或等于素数
   - 解决：选择更大的素数或减小秘密值

2. **错误："need at least threshold shares to recover"**
   - 原因：提供的份额数量少于门限值
   - 解决：提供足够数量的份额

3. **错误："shares must have the same length"**
   - 原因：加法同态时份额长度不匹配
   - 解决：确保两个份额集合长度相同

4. **错误："shares must have the same x coordinates"**
   - 原因：加法同态时份额的 x 坐标不匹配
   - 解决：确保两个份额集合的 x 坐标相同

## 13. 贡献指南

### 13.1 代码风格
- 遵循 Go 标准代码风格
- 使用 `go fmt` 格式化代码
- 编写清晰的文档和注释

### 13.2 测试要求
- 为新功能编写单元测试
- 确保测试覆盖率不低于 80%
- 测试边界条件和错误处理

### 13.3 提交规范
- 使用清晰的提交消息
- 提交前运行测试
- 遵循语义化版本规范

## 14. 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。
