# Tasks

- [x] Task 1: 项目初始化和基础结构搭建
  - [x] SubTask 1.1: 创建项目目录结构和uv.toml配置
  - [x] SubTask 1.2: 创建main.py主文件框架
  - [x] SubTask 1.3: 定义自定义异常类型（模数溢出错误等）
  - [x] SubTask 1.4: 定义字节序枚举类型

- [x] Task 2: 实现核心数学运算模块
  - [x] SubTask 2.1: 实现模数运算（mod = 2^32 - 5）
  - [x] SubTask 2.2: 实现扩展欧几里得算法和模逆元计算
  - [x] SubTask 2.3: 实现多项式评估函数（支持uint64中间运算）
  - [x] SubTask 2.4: 实现Lagrange插值恢复算法

- [x] Task 3: 实现字节序转换和Padding模块
  - [x] SubTask 3.1: 实现字节数组到uint32数组转换（大端序/小端序）
  - [x] SubTask 3.2: 实现uint32数组到字节数组转换
  - [x] SubTask 3.3: 实现PKCS风格Padding（填充到4字节倍数）
  - [x] SubTask 3.4: 实现Padding移除功能

- [x] Task 4: 实现Shamir秘密分享核心功能
  - [x] SubTask 4.1: 实现份额生成（split）功能
  - [x] SubTask 4.2: 实现秘密恢复（recover）功能
  - [x] SubTask 4.3: 实现uint32数组输入验证（检查是否超过模数）

- [x] Task 5: 实现零值扰动多项式（PSS特性）
  - [x] SubTask 5.1: 实现零值扰动多项式生成（常数项为0）
  - [x] SubTask 5.2: 实现份额更新功能
  - [x] SubTask 5.3: 验证更新后秘密不变

- [x] Task 6: 编写测试套件
  - [x] SubTask 6.1: 基础功能测试（分发与恢复）
  - [x] SubTask 6.2: PSS特性验证测试
  - [x] SubTask 6.3: 任意长度输入测试
  - [x] SubTask 6.4: Padding功能测试
  - [x] SubTask 6.5: 性能基准测试

- [x] Task 7: 编写开发文档
  - [x] SubTask 7.1: 编写算法原理说明
  - [x] SubTask 7.2: 编写API使用文档
  - [x] SubTask 7.3: 记录测试结果

- [x] Task 8: Git提交和最终验证
  - [x] SubTask 8.1: 完成所有功能的Git提交
  - [x] SubTask 8.2: 最终集成测试验证

# Task Dependencies
- [Task 2] depends on [Task 1]
- [Task 3] depends on [Task 1]
- [Task 4] depends on [Task 2, Task 3]
- [Task 5] depends on [Task 4]
- [Task 6] depends on [Task 5]
- [Task 7] depends on [Task 6]
- [Task 8] depends on [Task 7]
