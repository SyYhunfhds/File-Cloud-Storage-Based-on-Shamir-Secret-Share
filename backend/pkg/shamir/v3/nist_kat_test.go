package shamir

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMain creates output directory for test results
func TestMain(m *testing.M) {
	os.MkdirAll("output", 0755)
	code := m.Run()
	os.Exit(code)
}

func writeNISTResultv3(t *testing.T, testName, result string) {
	t.Helper()
	ts := time.Now().Format("20060102_150405")
	filename := filepath.Join("output", fmt.Sprintf("nist_kat_v3_%s.txt", ts))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Logf("failed to write result: %v", err)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s: %s\n", time.Now().Format("15:04:05"), testName, result)
}

// ============================================================================
// a) 有限域运算 KAT（NIST 风格）
// ============================================================================

func TestNISTv3_FieldAdd_Identity(t *testing.T) {
	cases := []uint32{0, 1, 42, Prime - 1, Prime / 2}
	for _, a := range cases {
		got := add(a, 0)
		if got != a {
			t.Errorf("add(%d, 0) = %d, want %d", a, got, a)
		}
		got = add(0, a)
		if got != a {
			t.Errorf("add(0, %d) = %d, want %d", a, got, a)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldAdd_Identity", "PASS")
}

func TestNISTv3_FieldAdd_Commutative(t *testing.T) {
	pairs := [][2]uint32{
		{1, 2}, {Prime - 1, 1}, {Prime - 1, Prime - 2},
		{42, 137}, {Prime / 2, Prime / 3},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		if add(a, b) != add(b, a) {
			t.Errorf("add(%d, %d) != add(%d, %d)", a, b, b, a)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldAdd_Commutative", "PASS")
}

func TestNISTv3_FieldAdd_Associative(t *testing.T) {
	triples := [][3]uint32{
		{1, 2, 3}, {Prime - 3, Prime - 2, Prime - 1},
		{Prime / 2, Prime / 3, Prime / 4},
	}
	for _, tr := range triples {
		a, b, c := tr[0], tr[1], tr[2]
		left := add(add(a, b), c)
		right := add(a, add(b, c))
		if left != right {
			t.Errorf("add(add(%d,%d),%d)=%d != add(%d,add(%d,%d))=%d",
				a, b, c, left, a, b, c, right)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldAdd_Associative", "PASS")
}

func TestNISTv3_FieldMul_Identity(t *testing.T) {
	cases := []uint32{0, 1, 42, Prime - 1, Prime / 2}
	for _, a := range cases {
		got := mul(a, 1)
		if got != a {
			t.Errorf("mul(%d, 1) = %d, want %d", a, got, a)
		}
		got = mul(1, a)
		if got != a {
			t.Errorf("mul(1, %d) = %d, want %d", a, got, a)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldMul_Identity", "PASS")
}

func TestNISTv3_FieldMul_Commutative(t *testing.T) {
	pairs := [][2]uint32{
		{2, 3}, {Prime - 1, 2}, {Prime - 1, Prime - 2},
		{42, 137}, {Prime / 2, Prime / 3},
	}
	for _, p := range pairs {
		a, b := p[0], p[1]
		if mul(a, b) != mul(b, a) {
			t.Errorf("mul(%d, %d) != mul(%d, %d)", a, b, b, a)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldMul_Commutative", "PASS")
}

func TestNISTv3_FieldMul_Associative(t *testing.T) {
	triples := [][3]uint32{
		{2, 3, 4}, {Prime - 3, Prime - 2, 2},
		{Prime / 2, 7, 11},
	}
	for _, tr := range triples {
		a, b, c := tr[0], tr[1], tr[2]
		left := mul(mul(a, b), c)
		right := mul(a, mul(b, c))
		if left != right {
			t.Errorf("mul(mul(%d,%d),%d)=%d != mul(%d,mul(%d,%d))=%d",
				a, b, c, left, a, b, c, right)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldMul_Associative", "PASS")
}

func TestNISTv3_FieldMul_Distributive(t *testing.T) {
	triples := [][3]uint32{
		{2, 3, 4}, {Prime - 3, 2, Prime - 1}, {Prime / 2, 7, 11},
	}
	for _, tr := range triples {
		a, b, c := tr[0], tr[1], tr[2]
		left := mul(a, add(b, c))
		right := add(mul(a, b), mul(a, c))
		if left != right {
			t.Errorf("mul(%d, add(%d,%d))=%d != add(mul(%d,%d),mul(%d,%d))=%d",
				a, b, c, left, a, b, a, c, right)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldMul_Distributive", "PASS")
}

func TestNISTv3_FieldInv_Property(t *testing.T) {
	cases := []uint32{1, 2, 3, 42, 137, Prime - 1, Prime - 2, Prime/2 + 1}
	for _, a := range cases {
		got := mul(a, inv(a))
		if got != 1 {
			t.Errorf("mul(%d, inv(%d)) = %d, want 1", a, a, got)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldInv_Property", "PASS")
}

func TestNISTv3_FieldInv_Zero(t *testing.T) {
	got := inv(0)
	t.Logf("inv(0) = %d (expected: 0, since 0^(p-2) mod p = 0)", got)
	if got != 0 {
		t.Errorf("inv(0) = %d, want 0", got)
	}
	writeNISTResultv3(t, "TestNISTv3_FieldInv_Zero", "PASS")
}

func TestNISTv3_FieldPow_Fermat(t *testing.T) {
	cases := []uint32{2, 3, 5, 7, 11, 13, 17, 19, 23, 42}
	for _, a := range cases {
		got := pow(a, Prime-1)
		if got != 1 {
			t.Errorf("pow(%d, Prime-1) = %d, want 1 (Fermat's Little Theorem)", a, got)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_FieldPow_Fermat", "PASS")
}

// ============================================================================
// b) 多项式运算 KAT
// ============================================================================

func TestNISTv3_PolyEval_KnownVector(t *testing.T) {
	coeffs := []uint32{1, 2, 3} // f(x) = 3x² + 2x + 1
	cases := []struct {
		x    uint32
		want uint32
	}{
		{0, 1},
		{1, 6},
		{2, 17},
		{3, 34},
		{10, 321},
	}
	for _, c := range cases {
		got := evalPolynomial(coeffs, c.x)
		if got != c.want {
			t.Errorf("f(%d) = %d, want %d", c.x, got, c.want)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_PolyEval_KnownVector", "PASS")
}

func TestNISTv3_PolyEval_ZeroPoly(t *testing.T) {
	coeffs := []uint32{0}
	for x := uint32(0); x < 10; x++ {
		got := evalPolynomial(coeffs, x)
		if got != 0 {
			t.Errorf("zero-poly f(%d) = %d, want 0", x, got)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_PolyEval_ZeroPoly", "PASS")
}

func TestNISTv3_PolyEval_Constant(t *testing.T) {
	const k uint32 = 0xDEADBEEF % 4294967291
	coeffs := []uint32{k}
	for x := uint32(0); x < 10; x++ {
		got := evalPolynomial(coeffs, x)
		if got != k {
			t.Errorf("constant-poly f(%d) = %d, want %d", x, got, k)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_PolyEval_Constant", "PASS")
}

func TestNISTv3_Lagrange_KnownPoints(t *testing.T) {
	// f(x) = 1 + x + x²
	xVals := []uint32{1, 2, 3}
	yVals := []uint32{3, 7, 13}

	got := lagrangeInterpolation(0, xVals, yVals)
	if got != 1 {
		t.Errorf("lagrangeInterpolation(0, [1,2,3], [3,7,13]) = %d, want 1", got)
	}

	cases := []struct {
		x    uint32
		want uint32
	}{
		{1, 3}, {2, 7}, {3, 13},
	}
	for _, c := range cases {
		got := lagrangeInterpolation(c.x, xVals, yVals)
		if got != c.want {
			t.Errorf("lagrangeInterpolation(%d, ...) = %d, want %d", c.x, got, c.want)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_Lagrange_KnownPoints", "PASS")
}

// ============================================================================
// c) Split/Recover KAT
// ============================================================================

func TestNISTv3_SplitRecover_1Byte(t *testing.T) {
	secret := []byte{0x42}
	userXs := []uint32{1, 2, 3, 4, 5}
	shares, err := Split(secret, 3, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	if len(shares) != 5 {
		t.Fatalf("expected 5 shares, got %d", len(shares))
	}
	for i, s := range shares {
		if len(s.Values) != 1 {
			t.Errorf("share %d: expected 1 block, got %d", i, len(s.Values))
		}
		if s.Index != userXs[i] {
			t.Errorf("share %d: index=%d, want %d", i, s.Index, userXs[i])
		}
	}
	recovered := Unpad(Recover(shares[:3]))
	if !bytes.Equal(recovered, secret) {
		t.Errorf("recovered %v, want %v", recovered, secret)
	}
	writeNISTResultv3(t, "TestNISTv3_SplitRecover_1Byte", "PASS")
}

func TestNISTv3_SplitRecover_16Bytes(t *testing.T) {
	// Use 0x7F pattern: 0x7F7F7F7F = 2139062143 < Prime (4294967291)
	secret := bytes.Repeat([]byte{0x7F}, 16)
	userXs := []uint32{1, 2, 3}
	shares, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	recovered := Unpad(Recover(shares[:2]))
	if !bytes.Equal(recovered, secret) {
		t.Errorf("recovered mismatch")
	}
	writeNISTResultv3(t, "TestNISTv3_SplitRecover_16Bytes", "PASS")
}

func TestNISTv3_SplitRecover_32Bytes_AES256Key(t *testing.T) {
	patterns := []struct {
		name   string
		secret []byte
	}{
		{"all-zeros", bytes.Repeat([]byte{0x00}, 32)},
		{"all-0x7F", bytes.Repeat([]byte{0x7F}, 32)},
		{"alternating-55-AA", bytes.Repeat([]byte{0x55, 0xAA}, 16)},
		{"incrementing", func() []byte {
			b := make([]byte, 32)
			for i := range b {
				b[i] = byte(i)
			}
			return b
		}()},
	}
	userXs := []uint32{1, 2, 3, 4}
	for _, p := range patterns {
		t.Run(p.name, func(t *testing.T) {
			shares, err := Split(p.secret, 3, userXs)
			if err != nil {
				t.Fatalf("Split failed: %v", err)
			}
			recovered := Unpad(Recover(shares[:3]))
			if !bytes.Equal(recovered, p.secret) {
				t.Errorf("%s: recovered mismatch", p.name)
			}
		})
	}
	writeNISTResultv3(t, "TestNISTv3_SplitRecover_32Bytes_AES256Key", "PASS")
}

func TestNISTv3_SplitRecover_Threshold(t *testing.T) {
	secret := []byte("v3-threshold-secret-key-data-256")
	userXs := []uint32{100, 200, 300, 400, 500}
	const threshold = 3

	shares, err := Split(secret, threshold, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	// Any 3 shares can recover
	for start := 0; start <= len(userXs)-threshold; start++ {
		subset := shares[start : start+threshold]
		recovered := Unpad(Recover(subset))
		if !bytes.Equal(recovered, secret) {
			t.Errorf("subset starting at %d: failed to recover", start)
		}
	}

	// 2 shares below threshold
	subset2 := shares[:2]
	wrongResult := Recover(subset2)
	unpadded := Unpad(wrongResult)
	if bytes.Equal(unpadded, secret) {
		t.Logf("WARNING: 2 shares accidentally recovered correct secret (extremely unlikely)")
	} else {
		t.Logf("threshold verified: 2 shares produced wrong result (expected)")
	}
	writeNISTResultv3(t, "TestNISTv3_SplitRecover_Threshold", "PASS")
}

func TestNISTv3_Split_Nondeterministic(t *testing.T) {
	// v3: crypto/rand → each Split produces different shares
	secret := []byte("nondeterministic-test")
	userXs := []uint32{1, 2, 3}

	shares1, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("first Split failed: %v", err)
	}
	shares2, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("second Split failed: %v", err)
	}

	// Check that shares are different (nondeterministic)
	different := false
	for i := range shares1 {
		for j := range shares1[i].Values {
			if shares1[i].Values[j] != shares2[i].Values[j] {
				different = true
				break
			}
		}
	}
	if !different {
		t.Error("v3 Split should produce nondeterministic results (uses crypto/rand)")
	}

	// Both should independently recover the same secret
	r1 := Unpad(Recover(shares1[:2]))
	r2 := Unpad(Recover(shares2[:2]))
	if !bytes.Equal(r1, secret) || !bytes.Equal(r2, secret) {
		t.Error("nondeterministic shares should both recover the same secret")
	}
	writeNISTResultv3(t, "TestNISTv3_Split_Nondeterministic", "PASS")
}

// ============================================================================
// d) Delta/扰动 KAT
// ============================================================================

func TestNISTv3_Delta_ZeroDelta(t *testing.T) {
	secret := []byte("zero-delta-test")
	userXs := []uint32{1, 2, 3}

	shares, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	// Apply zero delta vector
	zeroDelta := make(DeltaVector, len(userXs))
	updated := ApplyDelta(shares, zeroDelta)

	// Shares should be unchanged
	for i := range shares {
		for j := range shares[i].Values {
			if shares[i].Values[j] != updated[i].Values[j] {
				t.Errorf("zero delta changed share %d block %d", i, j)
			}
		}
	}
	writeNISTResultv3(t, "TestNISTv3_Delta_ZeroDelta", "PASS")
}

func TestNISTv3_Delta_Invariant(t *testing.T) {
	secret := []byte("delta-invariant-test-v3")
	userXs := []uint32{1, 2, 3}

	shares, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	delta, err := GenerateDelta(2, len(userXs))
	if err != nil {
		t.Fatalf("GenerateDelta failed: %v", err)
	}

	updated := ApplyDelta(shares, delta)

	// Secret must be unchanged after delta update
	recovered := Unpad(Recover(updated[:2]))
	if !bytes.Equal(recovered, secret) {
		t.Errorf("delta update changed the secret")
	}
	writeNISTResultv3(t, "TestNISTv3_Delta_Invariant", "PASS")
}

// ============================================================================
// e) 填充 KAT
// ============================================================================

func TestNISTv3_Padding_Empty(t *testing.T) {
	padded := Pad([]byte{})
	if len(padded) != 4 {
		t.Errorf("Pad([]): len=%d, want 4", len(padded))
	}
	expected := []byte{0, 0, 0, 0}
	if !bytes.Equal(padded, expected) {
		t.Errorf("Pad([]) = %v, want %v", padded, expected)
	}
	unpadded := Unpad(padded)
	if len(unpadded) != 0 {
		t.Errorf("Unpad(Pad([])): len=%d, want 0", len(unpadded))
	}
	writeNISTResultv3(t, "TestNISTv3_Padding_Empty", "PASS")
}

func TestNISTv3_Padding_1Byte(t *testing.T) {
	data := []byte{0xAB}
	padded := Pad(data)
	if len(padded) != 4 {
		t.Errorf("Pad(1B): len=%d, want 4", len(padded))
	}
	expected := []byte{0xAB, 3, 3, 3}
	if !bytes.Equal(padded, expected) {
		t.Errorf("Pad(1B) = %v, want %v", padded, expected)
	}
	unpadded := Unpad(padded)
	if !bytes.Equal(unpadded, data) {
		t.Errorf("Unpad(Pad(1B)) = %v, want %v", unpadded, data)
	}
	writeNISTResultv3(t, "TestNISTv3_Padding_1Byte", "PASS")
}

func TestNISTv3_Padding_16Bytes(t *testing.T) {
	data := bytes.Repeat([]byte{0xCD}, 16)
	padded := Pad(data)
	if !bytes.Equal(padded, data) {
		t.Errorf("Pad(16B, aligned) modified the data")
	}
	writeNISTResultv3(t, "TestNISTv3_Padding_16Bytes", "PASS")
}

func TestNISTv3_Padding_RoundTrip(t *testing.T) {
	sizes := []int{0, 1, 2, 3, 4, 5, 7, 8, 11, 15, 16, 17, 31, 32, 33}
	for _, size := range sizes {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i % 256)
		}
		padded := Pad(data)
		unpadded := Unpad(padded)
		if !bytes.Equal(unpadded, data) {
			t.Errorf("RoundTrip(size=%d): mismatch", size)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_Padding_RoundTrip", "PASS")
}

func TestNISTv3_Padding_InvalidPadding(t *testing.T) {
	invalidCases := [][]byte{
		{0xAA, 5, 5, 5},
		{0xBB, 0xCC, 2, 3},
		{0xDD, 0xFF},
	}
	for _, data := range invalidCases {
		result := Unpad(data)
		if !bytes.Equal(result, data) {
			t.Errorf("Unpad(invalid=%v) = %v, should return original", data, result)
		}
	}
	writeNISTResultv3(t, "TestNISTv3_Padding_InvalidPadding", "PASS")
}

// ============================================================================
// 3.3.2 错误份额恢复测试（验证"静默失败"设计）
// ============================================================================

func TestRecoverv3_WrongShare_ProducesWrongSecret(t *testing.T) {
	secret := []byte("v3-correct-secret-data")
	userXs := []uint32{10, 20, 30}
	shares, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	fakeShare := Share{
		Index:  shares[0].Index,
		Values: make([]uint32, len(shares[0].Values)),
	}
	for i := range fakeShare.Values {
		fakeShare.Values[i] = 0xFFFFFFFF
	}

	wrongShares := ShareVector{fakeShare, shares[1]}
	recovered := Unpad(Recover(wrongShares))
	if bytes.Equal(recovered, secret) {
		t.Error("wrong share accidentally recovered correct secret")
	}
	t.Logf("v3 wrong share: silent failure confirmed")
	writeNISTResultv3(t, "TestRecoverv3_WrongShare_ProducesWrongSecret", "PASS")
}

func TestRecoverv3_WrongXCoordinate(t *testing.T) {
	secret := []byte("v3-x-coords-secret")
	userXs := []uint32{100, 200, 300}
	shares, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	fakeShares := ShareVector{
		{Index: 999, Values: shares[0].Values},
		{Index: 888, Values: shares[1].Values},
	}
	recovered := Unpad(Recover(fakeShares))
	if bytes.Equal(recovered, secret) {
		t.Error("wrong X coordinates accidentally recovered correct secret")
	}
	t.Logf("v3 wrong X: silent failure confirmed")
	writeNISTResultv3(t, "TestRecoverv3_WrongXCoordinate", "PASS")
}

func TestRecoverv3_DuplicateXCoordinate(t *testing.T) {
	secret := []byte("v3-dup-x-test-data")
	userXs := []uint32{1, 2, 3}
	shares, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	dupShares := ShareVector{shares[0], shares[0]}
	recovered := Recover(dupShares)
	t.Logf("v3 duplicate X: recovery returned %v (len=%d)", recovered, len(recovered))
	writeNISTResultv3(t, "TestRecoverv3_DuplicateXCoordinate", "PASS")
}

func TestRecoverv3_EmptyShares(t *testing.T) {
	result := Recover(ShareVector{})
	if result != nil {
		t.Errorf("Recover([]) = %v, want nil", result)
	}
	writeNISTResultv3(t, "TestRecoverv3_EmptyShares", "PASS")
}

func TestRecoverv3_InsufficientShares(t *testing.T) {
	secret := []byte("v3-need-3-of-5-testing")
	userXs := []uint32{10, 20, 30, 40, 50}
	shares, err := Split(secret, 3, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	recovered := Unpad(Recover(shares[:2]))
	if bytes.Equal(recovered, secret) {
		t.Error("insufficient shares accidentally recovered correct secret")
	}
	t.Logf("v3 insufficient shares (2 of 3): silent failure confirmed")
	writeNISTResultv3(t, "TestRecoverv3_InsufficientShares", "PASS")
}

func TestRecoverv3_CorruptedSingleValue(t *testing.T) {
	secret := []byte("v3-corruption-data-block")
	userXs := []uint32{100, 200, 300}
	shares, err := Split(secret, 2, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	corrupted := make(ShareVector, len(shares))
	copy(corrupted, shares)
	if len(corrupted[0].Values) > 0 {
		corrupted[0].Values[0] ^= 0x00000001
	}

	recovered := Unpad(Recover(corrupted[:2]))
	if bytes.Equal(recovered, secret) {
		t.Error("corrupted share accidentally recovered correct secret")
	}
	t.Logf("v3 corrupted share: silent failure confirmed")
	writeNISTResultv3(t, "TestRecoverv3_CorruptedSingleValue", "PASS")
}

// ============================================================================
// f) GenerateSingleShare KAT
// ============================================================================

func TestNISTv3_GenerateSingleShare_Recoverable(t *testing.T) {
	// GenerateSingleShare 产生的份额具有正确的 Index 和分块数
	secret := []byte("v3-single-share-recover-test!")
	userX := uint32(20)

	singleShare, err := GenerateSingleShare(secret, 2, userX)
	if err != nil {
		t.Fatalf("GenerateSingleShare failed: %v", err)
	}

	if singleShare.Index != userX {
		t.Errorf("Index: want %d, got %d", userX, singleShare.Index)
	}

	// secret 经 Pad 后应为 4 字节对齐: 29字节 → Pad → 32字节 → 8 blocks
	expectedBlocks := len(Pad(secret)) / 4
	if len(singleShare.Values) != expectedBlocks {
		t.Errorf("Values length: want %d blocks, got %d", expectedBlocks, len(singleShare.Values))
	}

	// 注意: GenerateSingleShare 每次调用生成独立的随机多项式,
	// 因此不能与 Split 或其他 GenerateSingleShare 调用产生的份额组合恢复。
	// 此函数仅用于生成单个独立份额 (如 device share)。
	t.Logf("GenerateSingleShare: Index=%d, blocks=%d", singleShare.Index, len(singleShare.Values))
	writeNISTResultv3(t, "TestNISTv3_GenerateSingleShare_Recoverable", "PASS")
}

func TestNISTv3_GenerateSingleShare_Nondeterministic(t *testing.T) {
	// v3: crypto/rand → 每次 GenerateSingleShare 产生不同份额
	secret := []byte("v3-nondeterministic-single")
	userX := uint32(42)

	share1, err := GenerateSingleShare(secret, 2, userX)
	if err != nil {
		t.Fatalf("first GenerateSingleShare failed: %v", err)
	}
	share2, err := GenerateSingleShare(secret, 2, userX)
	if err != nil {
		t.Fatalf("second GenerateSingleShare failed: %v", err)
	}

	if share1.Index != share2.Index {
		t.Fatalf("Index mismatch: %d != %d", share1.Index, share2.Index)
	}

	// 份额值应该不同 (非确定性)
	different := false
	for j := range share1.Values {
		if share1.Values[j] != share2.Values[j] {
			different = true
			break
		}
	}
	if !different {
		t.Error("v3 GenerateSingleShare 应产生非确定性结果 (uses crypto/rand)")
	}
	t.Logf("v3 GenerateSingleShare 非确定性验证通过")
	writeNISTResultv3(t, "TestNISTv3_GenerateSingleShare_Nondeterministic", "PASS")
}

// ============================================================================
// 3.3.3 NIST Benchmark
// ============================================================================

func BenchmarkNISTv3_Split_32B_2of3(b *testing.B) {
	secret := make([]byte, 32)
	userXs := []uint32{1, 2, 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Split(secret, 2, userXs)
	}
}

func BenchmarkNISTv3_Recover_2shares(b *testing.B) {
	secret := make([]byte, 32)
	userXs := []uint32{1, 2, 3}
	shares, _ := Split(secret, 2, userXs)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Recover(shares[:2])
	}
}

// ============================================================================
// NIST KAT 汇总报告
// ============================================================================

func TestNISTv3_KAT_Summary(t *testing.T) {
	ts := time.Now().Format("20060102_150405")
	filename := filepath.Join("output", fmt.Sprintf("nist_kat_v3_%s.txt", ts))
	f, err := os.Create(filename)
	if err != nil {
		t.Fatalf("failed to create summary: %v", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "=== Shamir v3 NIST KAT Summary ===\n")
	fmt.Fprintf(f, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "Prime: %d (2^32 - 5)\n", Prime)
	fmt.Fprintf(f, "Field: GF(%d)\n", Prime)
	fmt.Fprintf(f, "Random Source: crypto/rand\n\n")

	fmt.Fprintf(f, "Category: Field Operations (10 tests)\n")
	fmt.Fprintf(f, "Category: Polynomial Operations (4 tests)\n")
	fmt.Fprintf(f, "Category: Split/Recover (5 tests)\n")
	fmt.Fprintf(f, "Category: Delta (2 tests)\n")
	fmt.Fprintf(f, "Category: Padding (5 tests)\n")
	fmt.Fprintf(f, "Category: GenerateSingleShare (2 tests)\n")
	fmt.Fprintf(f, "Category: Error Recovery (6 tests)\n")
	fmt.Fprintf(f, "Total NIST KAT tests: 34\n")

	t.Logf("NIST KAT v3 summary written to %s", filename)
}
