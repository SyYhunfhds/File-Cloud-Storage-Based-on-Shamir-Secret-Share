package logictest

import (
	"backend/internal/config"
	"backend/internal/logic"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.MkdirAll("output", 0755)
	code := m.Run()
	os.Exit(code)
}

func writeResult(t *testing.T, testName, result string) {
	t.Helper()
	ts := time.Now().Format("20060102_150405")
	filename := filepath.Join("output", fmt.Sprintf("test_hash-kits_%s.log", ts))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s: %s\n", time.Now().Format("15:04:05"), testName, result)
}

func newHashUtils() *logic.HashUtils {
	hu := logic.NewHashUtils()
	hu.BuildWithConfig(&config.ArgonConfig{
		Secret:      "test-secret-key-argon2",
		Memory:      64 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	})
	return hu
}

// ============================================================================
// 正向用例
// ============================================================================

func TestArgon2_HashGen_Success(t *testing.T) {
	hu := newHashUtils()
	hash, err := hu.HashGen("test-password-123")
	if err != nil {
		t.Fatalf("HashGen failed: %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
	if !strings.HasPrefix(hash, "$argon2") {
		t.Errorf("hash should start with $argon2, got: %s", hash[:20])
	}
	writeResult(t, "TestArgon2_HashGen_Success", "PASS")
}

func TestArgon2_HashGen_DifferentInputs(t *testing.T) {
	hu := newHashUtils()
	hash1, err := hu.HashGen("password-one")
	if err != nil {
		t.Fatalf("HashGen 1 failed: %v", err)
	}
	hash2, err := hu.HashGen("password-two")
	if err != nil {
		t.Fatalf("HashGen 2 failed: %v", err)
	}
	if hash1 == hash2 {
		t.Error("different passwords should produce different hashes")
	}
	writeResult(t, "TestArgon2_HashGen_DifferentInputs", "PASS")
}

func TestArgon2_HashVerify_CorrectPassword(t *testing.T) {
	hu := newHashUtils()
	password := "correct-password-456"
	hash, err := hu.HashGen(password)
	if err != nil {
		t.Fatalf("HashGen failed: %v", err)
	}
	valid, err := hu.HashVerify(password, hash)
	if err != nil {
		t.Fatalf("HashVerify failed: %v", err)
	}
	if !valid {
		t.Error("correct password should verify successfully")
	}
	writeResult(t, "TestArgon2_HashVerify_CorrectPassword", "PASS")
}

func TestArgon2_Memclr_ZeroesBuffer(t *testing.T) {
	buf := make([]byte, 1024)
	for i := range buf {
		buf[i] = 0xFF
	}
	logic.Memclr(buf)
	for i, b := range buf {
		if b != 0 {
			t.Errorf("buf[%d] = 0x%02X, want 0x00", i, b)
		}
	}
	writeResult(t, "TestArgon2_Memclr_ZeroesBuffer", "PASS")
}

// ============================================================================
// 反向用例
// ============================================================================

func TestArgon2_HashVerify_WrongPassword(t *testing.T) {
	hu := newHashUtils()
	hash, err := hu.HashGen("original-password")
	if err != nil {
		t.Fatalf("HashGen failed: %v", err)
	}
	valid, err := hu.HashVerify("wrong-password", hash)
	if err != nil {
		t.Fatalf("HashVerify failed: %v", err)
	}
	if valid {
		t.Error("wrong password should not verify")
	}
	writeResult(t, "TestArgon2_HashVerify_WrongPassword", "PASS")
}

func TestArgon2_HashVerify_TamperedHash(t *testing.T) {
	hu := newHashUtils()
	password := "tamper-test-password"
	hash, err := hu.HashGen(password)
	if err != nil {
		t.Fatalf("HashGen failed: %v", err)
	}
	// Tamper one byte in the hash
	tampered := []byte(hash)
	mid := len(tampered) / 2
	tampered[mid] ^= 0x01 // flip one bit
	valid, err := hu.HashVerify(password, string(tampered))
	if err != nil {
		t.Fatalf("HashVerify failed: %v", err)
	}
	if valid {
		t.Error("tampered hash should not verify")
	}
	writeResult(t, "TestArgon2_HashVerify_TamperedHash", "PASS")
}

func TestArgon2_HashVerify_EmptyPassword(t *testing.T) {
	hu := newHashUtils()
	hash, err := hu.HashGen("real-password")
	if err != nil {
		t.Fatalf("HashGen failed: %v", err)
	}
	valid, err := hu.HashVerify("", hash)
	if err != nil {
		t.Fatalf("HashVerify failed: %v", err)
	}
	if valid {
		t.Error("empty password against non-empty hash should not verify")
	}
	writeResult(t, "TestArgon2_HashVerify_EmptyPassword", "PASS")
}

func TestArgon2_HashVerify_EmptyHash(t *testing.T) {
	hu := newHashUtils()
	valid, err := hu.HashVerify("any-password", "")
	// Empty hash should either return error or false
	if err != nil {
		t.Logf("empty hash correctly returned error: %v", err)
		writeResult(t, "TestArgon2_HashVerify_EmptyHash", "PASS")
		return
	}
	if valid {
		t.Error("empty hash should not verify")
	}
	writeResult(t, "TestArgon2_HashVerify_EmptyHash", "PASS")
}

func TestArgon2_HashVerify_CaseSensitive(t *testing.T) {
	hu := newHashUtils()
	hash, err := hu.HashGen("CaseSensitive")
	if err != nil {
		t.Fatalf("HashGen failed: %v", err)
	}
	valid, err := hu.HashVerify("casesensitive", hash)
	if err != nil {
		t.Fatalf("HashVerify failed: %v", err)
	}
	if valid {
		t.Error("case-sensitive password should not match lowercase")
	}
	writeResult(t, "TestArgon2_HashVerify_CaseSensitive", "PASS")
}

func TestArgon2_Memclr_NilBuffer(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Memclr on nil buffer should not panic, got: %v", r)
		}
	}()
	logic.Memclr(nil)
	writeResult(t, "TestArgon2_Memclr_NilBuffer", "PASS")
}

func TestArgon2_Memclr_AlreadyZero(t *testing.T) {
	buf := make([]byte, 256) // all zeros by default
	logic.Memclr(buf)
	for i, b := range buf {
		if b != 0 {
			t.Errorf("buf[%d] = 0x%02X, want 0x00", i, b)
		}
	}
	writeResult(t, "TestArgon2_Memclr_AlreadyZero", "PASS")
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkArgon2HashGen(b *testing.B) {
	hu := newHashUtils()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hu.HashGen("benchmark-password")
	}
}

func BenchmarkArgon2HashVerify(b *testing.B) {
	hu := newHashUtils()
	hash, _ := hu.HashGen("benchmark-password")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hu.HashVerify("benchmark-password", hash)
	}
}

func BenchmarkMemclr_1KB(b *testing.B) {
	buf := make([]byte, 1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range buf {
			buf[j] = 0xFF
		}
		logic.Memclr(buf)
	}
}

func BenchmarkMemclr_1MB(b *testing.B) {
	buf := make([]byte, 1024*1024)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := range buf {
			buf[j] = 0xFF
		}
		logic.Memclr(buf)
	}
}
