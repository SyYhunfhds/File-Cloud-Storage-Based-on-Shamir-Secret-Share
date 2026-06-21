package xxtea

import (
	"testing"
)

var testKey = [4]uint32{0x01234567, 0x89ABCDEF, 0xFEDCBA98, 0x76543210}

func TestXXTEA_EncryptDecrypt_RoundTrip(t *testing.T) {
	original := []uint32{0xDEADBEEF, 0xCAFEBABE, 0x12345678, 0x9ABCDEF0}
	v := make([]uint32, len(original))
	copy(v, original)

	Encrypt(v, testKey)
	// 加密后不应与原文相同
	same := true
	for i := range v {
		if v[i] != original[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("加密后数据应与原文不同")
	}

	Decrypt(v, testKey)
	for i := range v {
		if v[i] != original[i] {
			t.Errorf("解密后索引 %d 不匹配: got 0x%08X, want 0x%08X", i, v[i], original[i])
		}
	}
}

func TestXXTEA_EncryptDecrypt_Uint32(t *testing.T) {
	// EncryptUint32 使用 2 元素块 [v, 0] 加密，返回 buf[0]
	// DecryptUint32 需要完整的加密对 [c0, c1] 才能解密
	// 此测试验证：如果保留完整加密对，往返可以恢复
	original := uint32(0xABCD1234)
	buf := []uint32{original, 0}
	Encrypt(buf, testKey) // 原地加密得到 [c0, c1]
	c0, c1 := buf[0], buf[1]

	if c0 == original {
		t.Error("加密后 buf[0] 应与原文不同")
	}

	// 解密需要完整对
	buf2 := []uint32{c0, c1}
	Decrypt(buf2, testKey)
	if buf2[0] != original || buf2[1] != 0 {
		t.Errorf("uint32对往返失败: got [0x%08X, 0x%08X], want [0x%08X, 0x00000000]",
			buf2[0], buf2[1], original)
	}
}

func TestXXTEA_EncryptDecrypt_LargeBlock(t *testing.T) {
	// 32个 uint32 (模拟大型密钥的份额数据)
	original := make([]uint32, 32)
	for i := range original {
		original[i] = uint32(i*0x11111111 + 0xABCD)
	}
	v := make([]uint32, len(original))
	copy(v, original)

	Encrypt(v, testKey)
	Decrypt(v, testKey)
	for i := range v {
		if v[i] != original[i] {
			t.Errorf("大块往返索引 %d: 不匹配", i)
			return
		}
	}
}

func TestXXTEA_Encrypt_ProducesDifferentOutput(t *testing.T) {
	input := []uint32{0x11111111, 0x22222222}
	v1 := []uint32{input[0], input[1]}
	v2 := []uint32{input[0], input[1] + 1}

	Encrypt(v1, testKey)
	Encrypt(v2, testKey)

	// 不同输入应产生不同输出
	if v1[0] == v2[0] && v1[1] == v2[1] {
		t.Error("不同输入应产生不同密文")
	}
}

func TestXXTEA_Encrypt_SameInputSameOutput(t *testing.T) {
	input := []uint32{0xAAAAAAAA, 0xBBBBBBBB, 0xCCCCCCCC}
	v1 := []uint32{input[0], input[1], input[2]}
	v2 := []uint32{input[0], input[1], input[2]}

	Encrypt(v1, testKey)
	Encrypt(v2, testKey)

	for i := range v1 {
		if v1[i] != v2[i] {
			t.Errorf("相同输入应产生相同输出，索引 %d 不同", i)
		}
	}
}

func TestXXTEA_Decrypt_SingleBitFlip(t *testing.T) {
	// XXTEA 对单比特翻转的"容错"特性：
	// 与 AES-GCM 不同，Decrypt 不会因认证失败而拒绝解密（容错）。
	// 但 XXTEA 作为块加密，比特翻转会导致雪崩效应——所有元素都会不同。
	// 此测试验证：解密不会 panic/报错，且结果与原文完全不同。
	original := []uint32{0xA1B2C3D4, 0xE5F60718, 0x192A3B4C, 0x5D6E7F80}
	v := make([]uint32, len(original))
	copy(v, original)

	Encrypt(v, testKey)

	// 在密文中翻转一个比特
	tampered := make([]uint32, len(v))
	copy(tampered, v)
	tampered[1] ^= 0x00000001 // 翻转第2个值的LSB

	Decrypt(v, testKey)        // 原始密文解密 → 应等于 original
	Decrypt(tampered, testKey) // 篡改密文解密 → 不报错但结果完全不同

	// 原始密文应正确还原
	for i := range v {
		if v[i] != original[i] {
			t.Errorf("原始密文解密索引 %d: got 0x%08X, want 0x%08X", i, v[i], original[i])
		}
	}

	// 篡改密文解密结果应与原文完全不同 (雪崩效应)
	allSame := true
	for i := range tampered {
		if tampered[i] != original[i] {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("篡改密文解密结果应完全不同 (雪崩效应)")
	}
}

func TestXXTEA_Decrypt_WrongKey(t *testing.T) {
	original := []uint32{0x12345678, 0x9ABCDEF0}
	wrongKey := [4]uint32{0xFFFFFFFF, 0xEEEEEEEE, 0xDDDDDDDD, 0xCCCCCCCC}
	v := make([]uint32, len(original))
	copy(v, original)

	Encrypt(v, testKey)
	Decrypt(v, wrongKey)

	// 错误密钥解密应产生错误结果
	if v[0] == original[0] && v[1] == original[1] {
		t.Error("错误密钥不应还原原文")
	}
}

func TestXXTEA_Decrypt_EmptyInput(t *testing.T) {
	// 空输入应直接返回
	v := []uint32{}
	Encrypt(v, testKey)
	Decrypt(v, testKey)
	if len(v) != 0 {
		t.Error("空输入应保持空")
	}
}

func TestXXTEA_Decrypt_SingleElement(t *testing.T) {
	// 单元素输入应直接返回（不满足 n>=2 条件）
	v := []uint32{0xDEADBEEF}
	Encrypt(v, testKey)
	if v[0] != 0xDEADBEEF {
		t.Error("单元素不应被加密")
	}
}

func TestKeyFromBytes(t *testing.T) {
	b := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF, 0xFE, 0xDC, 0xBA, 0x98, 0x76, 0x54, 0x32, 0x10}
	key := KeyFromBytes(b)
	expected := [4]uint32{0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476}
	if key != expected {
		t.Errorf("KeyFromBytes: got %08X, want %08X", key, expected)
	}
}

func TestKeyFromBytes_Short(t *testing.T) {
	b := []byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xAB, 0xCD, 0xEF}
	key := KeyFromBytes(b)
	// 后8字节补零
	expected := [4]uint32{0x67452301, 0xEFCDAB89, 0, 0}
	if key != expected {
		t.Errorf("KeyFromBytes short: got %08X, want %08X", key, expected)
	}
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkXXTEA_Encrypt_2Words(b *testing.B) {
	key := testKey
	v := []uint32{0xDEADBEEF, 0xCAFEBABE}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vv := []uint32{v[0], v[1]}
		Encrypt(vv, key)
	}
}

func BenchmarkXXTEA_Encrypt_8Words(b *testing.B) {
	key := testKey
	v := []uint32{0, 1, 2, 3, 4, 5, 6, 7}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vv := make([]uint32, 8)
		copy(vv, v)
		Encrypt(vv, key)
	}
}

func BenchmarkXXTEA_Decrypt_8Words(b *testing.B) {
	key := testKey
	v := []uint32{0, 1, 2, 3, 4, 5, 6, 7}
	Encrypt(v, key)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vv := make([]uint32, 8)
		copy(vv, v)
		Decrypt(vv, key)
	}
}
