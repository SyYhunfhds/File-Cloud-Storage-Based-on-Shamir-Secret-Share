## 内存问题
#### 字节比特翻转导致加密份额无法解密
- 文件上传时的加密份额字节(Base64编码)
```
sX7WMQl1VWLLjE6+vcWOZMQxsPwaUFVNLBhype92eyl3rDsAURV0AZg591lGWiQ9lIuAjmVr3lAA/u26e49Z55H6va1zlHP8pHWx71LA5DsDvj3BDpS/pAcI7MbKVPAs9vKlKD0ioRzUnxfCEZt+NVLCQlis00H39vgJgeOFwWCJk22QWr+f3oPeFO+J3wErQP9E13IQaIH6oxhOXsf5gh+ZgOG6lkY8QIpP0HlxUM+Y
```
校验和为3419171272
- 文件下载时从数据库取出的加密份额字节(Base64编码)
```
sX7WMQl1VWLLjE6+vcWOZMQxsPwaUFVNLBhype92Wyl3rDsAURV0AZg591lGWiQ9lIuAjmVr3lAA/u26e49Z55H6va1zlHP8pHWx71LA5DsDvj3BDpS/pAcI7MbKVPAs9vKlKD0ioRzUnxfCEZt+NVLCQlis00H39vgJgeOFwWCJk22QWr+f3oPeFO+J3wErQP9E13IQaIH6oxhOXsf5gh+ZgOG6lkY8QIpP0HlxUM+Y
```
校验和为2977688604

**唯一的不同**在`92`后面的字母, 原本是`e`, 存入数据库后变成了`W`
在 Base64 编码中，字符的改变直接对应着底层物理二进制位（Bits）的翻转。
- e 对应的 Base64 六位二进制是：011110
- W 对应的 Base64 六位二进制是：010110
> 这意味着：在数据从“上传完成”到“写入数据库（bytea）”，再到“从数据库取出”的某个微秒间隙，数据的某一个 bit，从 1 被静默篡改成了 0！

---
更多数据测试……

只在存盘时做一次内存拷贝
```
pzavKfENrYu8YF1qyuihU8ROeNibaWwon3JpNgRZ2QLu02T+WQVGoMKkVY5UztUSoSdGZ9PV+aZz1Yp2OXSXlrusT3/SsgV6LIW9QpAqiSycf0j8qJG+cvGwQFg8ZdqAT5Y4A8DZ4N5W9cAqr9We4iEkxPUTDrRoSpdj0RjLGV3xNe2kYTRsNx5bVrSM+zw/TbRn+eRvDt8U2OCLJuGkpWJvdF+u4nHf7rfhAGysAA==
```
校验和为 683044620

```
pzavKfENrYu8YH1qyuihU8ROeNibaWwon3JpNgRZ2QLu02T+WQVGoMKkVY5UztUSoSdGZ9PV+aZz1Yp2OXSXlrusT3/SsgV6LIW9QpAqiSycf0j8qJG+cvGwQFg8ZdqAT5Y4A8DZ4N5W9cAqr9We4iEkxPUTDrRoSpdj0RjLGX3xNe2kYTRsNx57VrSM+zw/TbRn+eRvDt8U2OCLJuGkpWJvdF+u4nHf7rfhAGysAA==
```
校验和为 232505424

---
然后在读取时也做一次内存拷贝
```
DcfjQTk9bANnpafsHLvE1P0D7F3zj+18Q0n9cq9VZxtUuLH0w2zm/QTFiW1LXasBlX0QP+i4Q3ry9Fy+JZVqKwzzD+RR+5XXR0sPbSNHNBZnSJm1fA3kNmJBCm6/ACxhXckcly697cz7NO0g55zSHgKblMggxjNeAFN3TT0+iZc8eYvVCCQKb52cNUwfLSFN8usO24JzE9EmQwaefu1A1eNJp4a2rA7sItFUDXtamg==
```
校验和为 2392434223

```
DcfjQTk9bANnpafsHLvE1P0D7H3zj+18Q0n9cq9VZxtUuLH0w2zm/QTFiW1LfasBlX0QP+i4Q3ry9Fy+JZVqKwzzD+RR+5XXR0sPbSNHNBZnSJm1fA3kNmJBCm6/ACxhfckcly697cz7NO0g55zSHgKblMggxjNeAFN3TT0+iZc8eYvVCCQKb52cNUwfLSFN8usO24JzE9EmQwaefu1A1eNJp4a2rA7sItFUDXtamg==
```
校验和为 15643013

第一组：存盘时深拷贝的数据（第 21 个字符）
- 上传加密： ...YF1qyuihU... （字符是 F，二进制：000101）
    - DB读出： ...YH1qyuihU... （字符是 H，二进制：000111）
- 右半部分突变（第 167 个字符）：
    - 上传加密： ...Wv3Ne2kYTRsNx5bV... （字符是 b，二进制：011011）
    - DB读出： ...Wv3Xne2kYTRsNx57V... （字符是 7，二进制：111011）

第二组：读取时深拷贝的数据（第 22 个字符）
- 上传加密： ...0D7F3zj+1... （字符是 F，二进制：000101）
    - DB读出： ...0D7H3zj+1... （字符是 H，二进制：000111）
- 右半部分突变（第 164 个字符）：
    - 上传加密： ...ACxhXckcl... （字符是 X，二进制：010111）
    - DB读出： ...ACxhfckcl... （字符是 f，二进制：100111）

1. F (000101) ➔ H (000111) —— 第 4 位 bit 从 0 变成了 1
2. b (011011) ➔ 7 (111011) —— 第 2 位 bit 发生了翻转
3. F (000101) ➔ H (000111) —— 第 4 位 bit 从 0 变成了 1
4. X (010111) ➔ f (100111) —— 第 2 位 bit 发生了翻转