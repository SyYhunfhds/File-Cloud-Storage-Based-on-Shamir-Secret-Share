# Shamir uint32数组输入方案规范

## Why
当前Shamir秘密分享方案仅支持GF(257)域的字节级处理，无法高效处理大容量数据。需要实现基于uint32数组的Shamir方案，支持任意长度字节输入，并具备PSS（Proactive Secret Sharing）特性以增强安全性。

## What Changes
- 实现基于模数 `2^32 - 5` 的Shamir秘密分享算法
- 支持uint32数组输入和输出
- 实现零值扰动多项式以支持PSS特性
- 实现PKCS风格Padding处理非4字节倍数输入
- 支持大端序/小端序字节转换
- 完整的测试套件和性能基准测试
- **BREAKING**: 新实现与现有GF(257)版本不兼容

## Impact
- Affected specs: 新增独立模块，不影响现有功能
- Affected code: `exploration/shamir-uint32-array-input/` 目录

## ADDED Requirements

### Requirement: 核心算法实现
系统应当提供基于uint32数组的Shamir秘密分享功能。

#### Scenario: 基础分发与恢复
- **WHEN** 用户提供任意长度的字节数组作为秘密
- **THEN** 系统能够将其分割为指定数量的份额，并能从足够数量的份额中恢复原始秘密

#### Scenario: 模数验证
- **WHEN** 输入数据中某个uint32值超过模数 `2^32 - 5`
- **THEN** 系统应当抛出自定义错误

### Requirement: 零值扰动多项式（PSS特性）
系统应当支持通过零值扰动多项式更新份额，而不改变原始秘密。

#### Scenario: 份额更新
- **WHEN** 用户请求更新份额
- **THEN** 系统生成新的份额，但恢复的秘密保持不变

#### Scenario: 多次更新
- **WHEN** 用户连续多次更新份额
- **THEN** 每次更新后秘密仍保持不变

### Requirement: Padding处理
系统应当能够处理不以4字节为倍数的字节数组输入。

#### Scenario: 非对齐输入
- **WHEN** 输入字节数组长度不是4的倍数
- **THEN** 系统自动应用PKCS风格Padding，恢复时正确移除Padding

### Requirement: 字节序支持
系统应当支持大端序和小端序的字节转换。

#### Scenario: 字节序选择
- **WHEN** 用户指定字节序（大端/小端）
- **THEN** 系统按照指定字节序进行字节数组与uint32数组的转换

### Requirement: 测试验证
系统应当通过以下测试验证：

1. **基础功能测试**: 分发份额并还原秘密
2. **PSS验证测试**: 验证零值扰动多项式特性
3. **任意长度输入测试**: 验证各种长度的字节数组输入
4. **Padding测试**: 验证非对齐输入的处理
5. **性能基准测试**: 测量分发和恢复的性能

### Requirement: 开发文档
系统应当提供完整的开发文档，说明算法原理、API使用方法和测试结果。
