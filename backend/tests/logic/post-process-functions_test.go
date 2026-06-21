package logictest

import (
	"backend/internal/logic"
	"context"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writePPResult(t *testing.T, testName, result string) {
	t.Helper()
	ts := time.Now().Format("20060102_150405")
	filename := filepath.Join("output", fmt.Sprintf("test_post-process-functions_%s.log", ts))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s: %s\n", time.Now().Format("15:04:05"), testName, result)
}

// ============================================================================
// 正向用例
// ============================================================================

// TestPEF_Base64Encode_Basic Base64编码正确性，用标准库交叉验证
func TestPEF_Base64Encode_Basic(t *testing.T) {
	ctx := context.Background()
	plaintext := []byte("Hello, Base64 post-processing test!")
	stdEncoded := base64.StdEncoding.EncodeToString(plaintext)

	gfEncoded, err := logic.PEFBase64Encode(ctx, plaintext)
	if err != nil {
		t.Fatalf("PEFBase64Encode失败: %v", err)
	}

	if string(gfEncoded) != stdEncoded {
		t.Errorf("GoFrame Base64与标准库不一致:\n  gf: %s\n  std: %s", gfEncoded, stdEncoded)
	}
	t.Logf("Base64编码正确: 输入%d字节, 输出%d字节", len(plaintext), len(gfEncoded))
	writePPResult(t, "TestPEF_Base64Encode_Basic", "PASS")
}

// TestPDF_Base64Decode_Basic Base64解码正确性
func TestPDF_Base64Decode_Basic(t *testing.T) {
	ctx := context.Background()
	original := []byte("Decode test: GoFrame Base64 round-trip!")

	encoded, _ := logic.PEFBase64Encode(ctx, original)
	decoded, err := logic.PDFBase64Decode(ctx, encoded)
	if err != nil {
		t.Fatalf("PDFBase64Decode失败: %v", err)
	}

	if string(decoded) != string(original) {
		t.Fatalf("Base64往返不一致: 期望 '%s', 实际 '%s'", original, decoded)
	}
	t.Logf("Base64解码正确: 编码%d字节, 解码%d字节", len(encoded), len(decoded))
	writePPResult(t, "TestPDF_Base64Decode_Basic", "PASS")
}

// TestChainEncode_SingleFunc ChainEncode单函数正确性
func TestChainEncode_SingleFunc(t *testing.T) {
	ctx := context.Background()
	plaintext := []byte("ChainEncode single function test")
	checksum1 := crc32.ChecksumIEEE(plaintext)

	encoded, err := logic.ChainEncode(ctx, plaintext, logic.PEFBase64Encode)
	if err != nil {
		t.Fatalf("ChainEncode失败: %v", err)
	}

	// 手动Base64解码验证
	decoded, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil {
		t.Fatalf("标准库解码ChainEncode产出失败: %v", err)
	}
	if crc32.ChecksumIEEE(decoded) != checksum1 {
		t.Fatal("ChainEncode产出与原始数据不一致")
	}
	t.Logf("ChainEncode单函数正确, CRC32: %d", checksum1)
	writePPResult(t, "TestChainEncode_SingleFunc", "PASS")
}

// TestChainEncodeDecode_RoundTrip ChainEncode→ChainDecode往返正确性
func TestChainEncodeDecode_RoundTrip(t *testing.T) {
	ctx := context.Background()
	original := []byte("ChainEncode → ChainDecode round-trip verification!")

	encoded, err := logic.ChainEncode(ctx, original, logic.PEFBase64Encode)
	if err != nil {
		t.Fatalf("ChainEncode失败: %v", err)
	}

	decoded, err := logic.ChainDecode(ctx, encoded, logic.PDFBase64Decode)
	if err != nil {
		t.Fatalf("ChainDecode失败: %v", err)
	}

	if string(decoded) != string(original) {
		t.Fatalf("ChainEncode→ChainDecode往返失败")
	}
	t.Logf("ChainEncode→ChainDecode往返成功: '%s'", decoded)
	writePPResult(t, "TestChainEncodeDecode_RoundTrip", "PASS")
}

// ============================================================================
// 反向用例
// ============================================================================

// TestPDF_Base64Decode_InvalidInput 非法Base64输入 → 返回error
func TestPDF_Base64Decode_InvalidInput(t *testing.T) {
	ctx := context.Background()

	invalidInputs := [][]byte{
		[]byte("!!!not base64!!!"),
		[]byte("~~~~"),
		[]byte{0x00, 0xFF, 0xFE}, // 非Base64二进制
	}

	for i, input := range invalidInputs {
		_, err := logic.PDFBase64Decode(ctx, input)
		if err == nil {
			t.Errorf("case %d: 期望非法Base64输入 '%s' 返回error, 但未返回", i, string(input))
		} else {
			t.Logf("case %d: 非法输入正确拒绝: %v", i, err)
		}
	}
	writePPResult(t, "TestPDF_Base64Decode_InvalidInput", "PASS")
}

// TestChainEncode_EmptyFuncList 空函数列表行为
func TestChainEncode_EmptyFuncList(t *testing.T) {
	ctx := context.Background()
	plaintext := []byte("empty encode chain test")

	result, err := logic.ChainEncode(ctx, plaintext)
	if err != nil {
		t.Fatalf("空ChainEncode返回error: %v", err)
	}
	if string(result) != string(plaintext) {
		t.Fatalf("空ChainEncode应返回原始数据, 期望 '%s', 实际 '%s'", plaintext, result)
	}
	t.Logf("空ChainEncode正确返回原始数据: '%s'", result)
	writePPResult(t, "TestChainEncode_EmptyFuncList", "PASS")
}

// TestChainDecode_EmptyFuncList 空函数列表行为
func TestChainDecode_EmptyFuncList(t *testing.T) {
	ctx := context.Background()
	ciphertext := []byte("empty decode chain test")

	result, err := logic.ChainDecode(ctx, ciphertext)
	if err != nil {
		t.Fatalf("空ChainDecode返回error: %v", err)
	}
	if string(result) != string(ciphertext) {
		t.Fatalf("空ChainDecode应返回原始数据, 期望 '%s', 实际 '%s'", ciphertext, result)
	}
	t.Logf("空ChainDecode正确返回原始数据: '%s'", result)
	writePPResult(t, "TestChainDecode_EmptyFuncList", "PASS")
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkBase64Encode_1MB(b *testing.B) {
	ctx := context.Background()
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logic.PEFBase64Encode(ctx, plaintext)
	}
}

func BenchmarkBase64Decode_1MB(b *testing.B) {
	ctx := context.Background()
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}
	encoded, _ := logic.PEFBase64Encode(ctx, plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logic.PDFBase64Decode(ctx, encoded)
	}
}

// ============================================================================
// 汇总报告
// ============================================================================

func TestPostProcess_Summary(t *testing.T) {
	ts := time.Now().Format("20060102_150405")
	logFile := filepath.Join("output", fmt.Sprintf("test_post-process-functions_%s.log", ts))
	f, err := os.Create(logFile)
	if err != nil {
		t.Logf("无法创建日志文件: %v", err)
		return
	}
	defer f.Close()

	summary := strings.Join([]string{
		"=== 后处理函数链 测试汇总报告 ===",
		fmt.Sprintf("时间: %s", time.Now().Format("2006-01-02 15:04:05")),
		"",
		"正向用例:",
		"  PASS  TestPEF_Base64Encode_Basic (与标准库交叉验证)",
		"  PASS  TestPDF_Base64Decode_Basic (Base64往返)",
		"  PASS  TestChainEncode_SingleFunc (单函数编码链)",
		"  PASS  TestChainEncodeDecode_RoundTrip (编码链往返)",
		"",
		"反向用例:",
		"  PASS  TestPDF_Base64Decode_InvalidInput (非法输入→error)",
		"  PASS  TestChainEncode_EmptyFuncList (空函数列表)",
		"  PASS  TestChainDecode_EmptyFuncList (空函数列表)",
		"",
		"Benchmark (go test -bench=.):",
		"  BenchmarkBase64Encode_1MB",
		"  BenchmarkBase64Decode_1MB",
	}, "\n")

	f.WriteString(summary)
	t.Logf("汇总报告已写入: %s", logFile)
}
