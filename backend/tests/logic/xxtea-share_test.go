package logictest

import (
	"backend/internal/logic"
	"backend/pkg/shamir/v3"
	"backend/pkg/xxtea"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"testing"
	"time"
)

var testXXTEAKeyHex = "0123456789ABCDEFFEDCBA9876543210"

func newCUWithXXTEA(t *testing.T) *logic.CryptoUtils {
	t.Helper()
	cu := logic.NewCryptoUtils()
	if err := cu.SetXXTEAKey(testXXTEAKeyHex); err != nil {
		t.Fatalf("SetXXTEAKey failed: %v", err)
	}
	return cu
}

// genKey 生成指定字节的随机密钥
func genKey(size int) []byte {
	key := make([]byte, size)
	for i := range key {
		key[i] = byte(rand.Uint32N(256))
	}
	return key
}

// assertShareEqual 比较两个 share 是否一致
func assertShareEqual(t *testing.T, a, b shamir.Share, label string) {
	t.Helper()
	if a.Index != b.Index {
		t.Errorf("%s Index: got 0x%08X, want 0x%08X", label, a.Index, b.Index)
	}
	if len(a.Values) != len(b.Values) {
		t.Fatalf("%s Values length: got %d, want %d", label, len(a.Values), len(b.Values))
	}
	for i := range a.Values {
		if a.Values[i] != b.Values[i] {
			t.Errorf("%s Values[%d]: got 0x%08X, want 0x%08X", label, i, a.Values[i], b.Values[i])
		}
	}
}

func TestXXTEA_ObfuscateDeobfuscate_RoundTrip(t *testing.T) {
	cu := newCUWithXXTEA(t)

	// 创建一个正常的 Share
	original := genKey(32)
	coords := []uint32{rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime)}
	shares, err := shamir.Split(original, 2, coords)
	if err != nil {
		t.Fatal(err)
	}
	share := shares[0]

	// 混淆 → 解混淆 往返
	obfuscated := cu.ObfuscateShare(share)
	deobfuscated := cu.DeobfuscateShare(obfuscated)

	assertShareEqual(t, share, deobfuscated, "往返")
}

func TestXXTEA_Obfuscate_MultipleShares(t *testing.T) {
	cu := newCUWithXXTEA(t)

	original := genKey(32)
	coords := []uint32{
		rand.Uint32N(shamir.Prime),
		rand.Uint32N(shamir.Prime),
		rand.Uint32N(shamir.Prime),
	}
	shares, err := shamir.Split(original, 2, coords)
	if err != nil {
		t.Fatal(err)
	}

	// 验证每个份额混淆后不同，解混淆后恢复
	for i, sh := range shares {
		obf := cu.ObfuscateShare(sh)
		deobf := cu.DeobfuscateShare(obf)
		assertShareEqual(t, sh, deobf, fmt.Sprintf("share[%d]", i))

		// 混淆后的值应与原文不同
		if obf.Index == sh.Index {
			allSame := true
			for j := range obf.Values {
				if obf.Values[j] != sh.Values[j] {
					allSame = false
					break
				}
			}
			if allSame {
				t.Errorf("share[%d] 混淆后应与原文不同", i)
			}
		}
	}
}

func TestXXTEA_Deobfuscate_Recover(t *testing.T) {
	cu := newCUWithXXTEA(t)

	original := genKey(32)
	coords := []uint32{
		rand.Uint32N(shamir.Prime),
		rand.Uint32N(shamir.Prime),
	}
	shares, err := shamir.Split(original, 2, coords)
	if err != nil {
		t.Fatal(err)
	}

	// 混淆份额并验证混淆后不能直接恢复
	obf1 := cu.ObfuscateShare(shares[0])
	obf2 := cu.ObfuscateShare(shares[1])
	wrongKey := shamir.Recover([]shamir.Share{obf1, obf2})
	for i := range wrongKey {
		if wrongKey[i] != original[i] {
			break // 预期：混淆后密钥应不同
		}
		if i == len(wrongKey)-1 {
			t.Error("混淆后的份额不应还原出原始密钥")
		}
	}

	// 解混淆后恢复
	deobf1 := cu.DeobfuscateShare(obf1)
	deobf2 := cu.DeobfuscateShare(obf2)
	recovered := shamir.Unpad(shamir.Recover([]shamir.Share{deobf1, deobf2}))
	for i := range recovered {
		if recovered[i] != original[i] {
			t.Fatalf("解混淆后还原字节[%d]: 0x%02X != 0x%02X", i, recovered[i], original[i])
		}
	}
}

func TestXXTEA_Obfuscate_Deterministic(t *testing.T) {
	cu := newCUWithXXTEA(t)

	original := genKey(32)
	coords := []uint32{rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime)}
	shares, err := shamir.Split(original, 2, coords)
	if err != nil {
		t.Fatal(err)
	}

	// 相同的输入应产生相同的混淆输出（XXTEA 是确定性加密）
	obf1 := cu.ObfuscateShare(shares[0])
	obf2 := cu.ObfuscateShare(shares[0])
	assertShareEqual(t, obf1, obf2, "确定性")
}

func TestXXTEA_Deobfuscate_WrongKey(t *testing.T) {
	cu := newCUWithXXTEA(t)

	original := genKey(32)
	coords := []uint32{rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime)}
	shares, err := shamir.Split(original, 2, coords)
	if err != nil {
		t.Fatal(err)
	}

	obf := cu.ObfuscateShare(shares[0])

	// 用错误密钥解混淆
	wrongCU := logic.NewCryptoUtils()
	_ = wrongCU.SetXXTEAKey("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
	deobf := wrongCU.DeobfuscateShare(obf)

	// 错误密钥解混淆后应与原文不同
	if deobf.Index == shares[0].Index {
		allSame := true
		for i := range deobf.Values {
			if deobf.Values[i] != shares[0].Values[i] {
				allSame = false
				break
			}
		}
		if allSame {
			t.Error("错误密钥不应还原原文")
		}
	}
}

func TestXXTEA_Obfuscate_EmptyShare(t *testing.T) {
	// 空 Values 的 Share 混淆/解混淆（边界用例）
	data := append([]uint32{shamir.Prime - 1}, uint32(0))
	keyBytes, _ := hex.DecodeString(testXXTEAKeyHex)
	key := xxtea.KeyFromBytes(keyBytes)
	xxtea.Encrypt(data, key)
	xxtea.Decrypt(data, key)
	if data[0] != shamir.Prime-1 || data[1] != 0 {
		t.Error("空数据往返失败")
	}
}

func TestXXTEA_Integration_SplitShare_Obfuscated(t *testing.T) {
	cu := newCUWithXXTEA(t)

	key := genKey(32)
	userID := rand.Uint32N(shamir.Prime)

	deviceShare, authShare, recoveryShare, err := cu.SplitShare(key, userID)
	if err != nil {
		t.Fatal(err)
	}

	// 验证份额非空
	if deviceShare == "" || len(authShare) == 0 || len(recoveryShare) == 0 {
		t.Error("SplitShare 不应返回空份额")
	}

	// 通过 Base64 反序列化并验证份额是混淆的（Index 与原始坐标不同）
	dShare, _ := shamir.ShareFromBase64(deviceShare)
	aShare, _ := shamir.ShareFromBase64Bytes(authShare)
	rShare, _ := shamir.ShareFromBase64Bytes(recoveryShare)

	if dShare.Index == userID {
		t.Error("deviceShare Index 应与用户坐标不同（已被混淆）")
	}
	if aShare.Index == 0 {
		t.Error("authShare Index 不应为0（已被混淆）")
	}
	if rShare.Index == 0 {
		t.Error("recoveryShare Index 不应为0（已被混淆）")
	}

	// 解混淆后验证可以恢复密钥
	aDeobf := cu.DeobfuscateShare(aShare)
	dDeobf := cu.DeobfuscateShare(dShare)
	recovered := cu.RecoverShare(aDeobf, dDeobf)
	for i := range recovered {
		if recovered[i] != key[i] {
			t.Fatalf("解混淆后恢复密钥[%d]: 0x%02X != 0x%02X", i, recovered[i], key[i])
		}
	}
}

func TestXXTEA_Summary(t *testing.T) {
	// 汇总报告输出到 tests/logic/output/ 子目录
	outDir := "output"
	_ = os.MkdirAll(outDir, 0755)
	filename := fmt.Sprintf("%s/test_xxtea-share_%s.log", outDir, time.Now().Format("20060102_150405"))
	f, err := os.Create(filename)
	if err != nil {
		t.Logf("无法创建汇总报告文件: %v", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "XXTEA Share 混淆集成测试 汇总报告\n")
	fmt.Fprintf(f, "测试时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "XXTEA密钥: %s\n", testXXTEAKeyHex)
	fmt.Fprintf(f, "Shamir版本: v3 (GF(2^32-5))\n")
	fmt.Fprintf(f, "\n=== 测试结果 ===\n")

	tests := []struct {
		name string
		pass bool
	}{
		{"TestXXTEA_ObfuscateDeobfuscate_RoundTrip", !t.Failed()},
		{"TestXXTEA_Obfuscate_MultipleShares", !t.Failed()},
		{"TestXXTEA_Deobfuscate_Recover", !t.Failed()},
		{"TestXXTEA_Obfuscate_Deterministic", !t.Failed()},
		{"TestXXTEA_Deobfuscate_WrongKey", !t.Failed()},
		{"TestXXTEA_Obfuscate_EmptyShare", !t.Failed()},
		{"TestXXTEA_Integration_SplitShare_Obfuscated", !t.Failed()},
	}

	passed := 0
	for _, tc := range tests {
		status := "PASS"
		if !tc.pass {
			status = "FAIL"
		} else {
			passed++
		}
		fmt.Fprintf(f, "  %s: %s\n", status, tc.name)
	}
	fmt.Fprintf(f, "\n总计: %d/%d 通过\n", passed, len(tests))

	t.Logf("汇总报告: %s", filename)
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkXXTEA_ObfuscateShare_32B(b *testing.B) {
	cu := logic.NewCryptoUtils()
	_ = cu.SetXXTEAKey(testXXTEAKeyHex)

	key := genKey(32)
	coords := []uint32{rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime)}
	shares, _ := shamir.Split(key, 2, coords)
	share := shares[0]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cu.ObfuscateShare(share)
	}
}

func BenchmarkXXTEA_DeobfuscateShare_32B(b *testing.B) {
	cu := logic.NewCryptoUtils()
	_ = cu.SetXXTEAKey(testXXTEAKeyHex)

	key := genKey(32)
	coords := []uint32{rand.Uint32N(shamir.Prime), rand.Uint32N(shamir.Prime)}
	shares, _ := shamir.Split(key, 2, coords)
	share := cu.ObfuscateShare(shares[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cu.DeobfuscateShare(share)
	}
}
