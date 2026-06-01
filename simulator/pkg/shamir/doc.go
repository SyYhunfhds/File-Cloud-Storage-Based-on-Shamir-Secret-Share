// Package shamir 提供了Shamir秘密共享算法的工程化实现
//
// 该包实现了基于有限域的Shamir秘密共享协议，支持将秘密分割为多个份额，并从阈值数量的份额中恢复原始秘密。
//
// 主要功能包括：
// 1. 密钥分割：将单个秘密分割为n个份额
// 2. 密钥恢复：从t个或更多份额中恢复原始秘密
// 3. 动态成员管理：支持成员加入和退出
// 4. 份额更新：通过零值扰动多项式实现份额更新
//
// 示例用法：
//
//     // 创建一个Shamir实例，门限值为3，份额数量为5，使用素数257
//     s, err := shamir.New(
//         shamir.WithThreshold(3),
//         shamir.WithNumShares(5),
//         shamir.WithPrime(257),
//     )
//
//     // 分割秘密
//     shares, err := s.Split(42)
//
//     // 从份额中恢复秘密
//     recovered, err := s.Recover(shares[:3])
//
package shamir
