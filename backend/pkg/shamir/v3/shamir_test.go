package shamir

import (
	"bytes"
	"testing"
	"time"
)

type basicTestCase struct {
	Name      string
	Secret    []byte
	Threshold int
	UserXs    []uint32
}

func TestBasicSplitRecover(t *testing.T) {
	testCases := []basicTestCase{
		{Name: "4-byte-secret", Secret: []byte("test"), Threshold: 2, UserXs: []uint32{1001, 2002, 3003}},
		{Name: "8-byte-secret", Secret: []byte("testing8"), Threshold: 3, UserXs: []uint32{5, 17, 23, 42, 99}},
		{Name: "16-byte-secret", Secret: []byte("aes-key-16bytes!"), Threshold: 3, UserXs: []uint32{7, 13, 19, 29, 37}},
		{Name: "5-byte-secret", Secret: []byte("hello"), Threshold: 2, UserXs: []uint32{100, 200, 300}},
		{Name: "1-byte-secret", Secret: []byte("x"), Threshold: 2, UserXs: []uint32{1, 2, 3}},
		{Name: "empty-secret", Secret: []byte(""), Threshold: 2, UserXs: []uint32{10, 20, 30}},
		{Name: "32-byte-secret", Secret: []byte("this-is-a-32-byte-secret-key!!"), Threshold: 5, UserXs: []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			testBasicSplitRecover(t, tc)
		})
	}
}

func testBasicSplitRecover(t *testing.T, tc basicTestCase) {
	shares, err := Split(tc.Secret, tc.Threshold, tc.UserXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if len(shares) != len(tc.UserXs) {
		t.Errorf("expected %d shares, got %d", len(tc.UserXs), len(shares))
	}

	expectedBlocks := (len(tc.Secret) + 3) / 4
	if len(tc.Secret) == 0 {
		expectedBlocks = 1
	}

	for i, share := range shares {
		found := false
		for _, expectedX := range tc.UserXs {
			if share.Index == expectedX {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("share %d: index %d not in expected userXs", i, share.Index)
		}
		if len(share.Values) != expectedBlocks {
			t.Errorf("share %d: expected %d blocks, got %d", i, expectedBlocks, len(share.Values))
		}
	}

	recovered := Recover(shares)
	expected := Pad(tc.Secret)

	if !bytes.Equal(recovered, expected) {
		t.Errorf("recovered secret mismatch: expected %v, got %v", expected, recovered)
	}

	if len(shares) >= tc.Threshold {
		subset := shares[:tc.Threshold]
		recoveredSubset := Recover(subset)
		if !bytes.Equal(recoveredSubset, expected) {
			t.Errorf("recovered subset secret mismatch: expected %v, got %v", expected, recoveredSubset)
		}
	}
}

func TestDeltaUpdate(t *testing.T) {
	secret := []byte("test-secret-for-delta")
	threshold := 3
	userXs := []uint32{1001, 2002, 3003, 4004, 5005}

	shares, err := Split(secret, threshold, userXs)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}
	originalRecovered := Recover(shares)

	delta, err := GenerateDelta(threshold, len(userXs))
	if err != nil {
		t.Fatalf("GenerateDelta failed: %v", err)
	}

	updatedShares := ApplyDelta(shares, delta)
	updatedRecovered := Recover(updatedShares)

	if !bytes.Equal(updatedRecovered, originalRecovered) {
		t.Errorf("secret changed after delta update: expected %v, got %v", originalRecovered, updatedRecovered)
	}
}

func TestGenerateUserXFromID(t *testing.T) {
	id1 := "user123"
	id2 := "user456"

	x1 := GenerateUserXFromID(id1)
	x2 := GenerateUserXFromID(id2)

	if x1 == 0 || x2 == 0 {
		t.Error("GenerateUserXFromID should never return 0")
	}

	if x1 == x2 {
		t.Error("Different user IDs should produce different X values")
	}

	if GenerateUserXFromID(id1) != GenerateUserXFromID(id1) {
		t.Error("GenerateUserXFromID should be deterministic")
	}
}

func TestPadding(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    []byte
		Expected int
	}{
		{"4-byte", []byte("test"), 4},
		{"5-byte", []byte("hello"), 8},
		{"8-byte", []byte("testing8"), 8},
		{"1-byte", []byte("x"), 4},
		{"empty", []byte(""), 4},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			padded := Pad(tc.Input)
			if len(padded) != tc.Expected {
				t.Errorf("expected padded length %d, got %d", tc.Expected, len(padded))
			}

			if len(padded)%4 != 0 {
				t.Errorf("padded length %d is not divisible by 4", len(padded))
			}

			unpadded := Unpad(padded)
			if !bytes.Equal(unpadded, tc.Input) {
				t.Errorf("unpad mismatch: expected %v, got %v", tc.Input, unpadded)
			}
		})
	}
}

func TestNoDoublePadding(t *testing.T) {
	secret := []byte("test")
	paddedOnce := Pad(secret)
	paddedTwice := Pad(paddedOnce)

	if len(paddedTwice) != len(paddedOnce) {
		t.Errorf("double padding changed length: %d vs %d", len(paddedOnce), len(paddedTwice))
	}

	if !bytes.Equal(paddedOnce, paddedTwice) {
		t.Errorf("double padding changed content")
	}
}

func TestFieldOperations(t *testing.T) {
	testCases := []struct {
		a, b        uint32
		expectedAdd uint32
		expectedSub uint32
	}{
		{10, 5, 15, 5},
		{5, 10, 15, 4294967291 - 5}, // add: 5+10=15; sub: 5-10 < 0, returns p - 5
		{0, 0, 0, 0},
		{4294967290, 1, 0, 4294967289}, // add: p-1+1=p=0; sub: p-1-1=p-2
	}

	for _, tc := range testCases {
		result := add(tc.a, tc.b)
		if result != tc.expectedAdd {
			t.Errorf("add(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expectedAdd)
		}

		result = sub(tc.a, tc.b)
		if result != tc.expectedSub {
			t.Errorf("sub(%d, %d) = %d, expected %d", tc.a, tc.b, result, tc.expectedSub)
		}
	}
}

func TestPerformanceThreshold(t *testing.T) {
	testCases := []struct {
		name      string
		secretLen int
		threshold int
		userXs    []uint32
		maxTime   time.Duration
	}{
		{"quick-test", 16, 3, makeUserXs(10), 500 * time.Millisecond},
		{"medium-test", 64, 5, makeUserXs(20), 500 * time.Millisecond},
		{"large-test", 256, 10, makeUserXs(50), 500 * time.Millisecond},
		{"huge-test", 1024, 20, makeUserXs(100), 500 * time.Millisecond},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			secret := make([]byte, tc.secretLen)
			for i := range secret {
				secret[i] = byte(i)
			}

			start := time.Now()
			for i := 0; i < 100; i++ {
				_, err := Split(secret, tc.threshold, tc.userXs)
				if err != nil {
					t.Fatalf("Split failed: %v", err)
				}
			}
			elapsed := time.Since(start)

			if elapsed > tc.maxTime {
				t.Logf("test %s took %v, which exceeds %v", tc.name, elapsed, tc.maxTime)
			}
		})
	}
}

func makeUserXs(n int) []uint32 {
	xs := make([]uint32, n)
	for i := 0; i < n; i++ {
		xs[i] = uint32(i + 1)
	}
	return xs
}

func BenchmarkSplit(b *testing.B) {
	benchmarkCases := []struct {
		name      string
		secretLen int
		threshold int
		userXs    []uint32
	}{
		{"small-secret-small-share", 4, 2, []uint32{1, 2, 3}},
		{"small-secret-medium-share", 4, 2, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"medium-secret-small-share", 16, 3, []uint32{1, 2, 3, 4, 5}},
		{"medium-secret-medium-share", 16, 3, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"large-secret-small-share", 64, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"large-secret-large-share", 64, 10, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
	}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			secret := make([]byte, tc.secretLen)
			for i := range secret {
				secret[i] = byte(i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Split(secret, tc.threshold, tc.userXs)
			}
		})
	}
}

func BenchmarkRecover(b *testing.B) {
	benchmarkCases := []struct {
		name      string
		secretLen int
		threshold int
		userXs    []uint32
	}{
		{"small-secret-small-share", 4, 2, []uint32{1, 2, 3}},
		{"medium-secret-small-share", 16, 3, []uint32{1, 2, 3, 4, 5}},
		{"large-secret-small-share", 64, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			secret := make([]byte, tc.secretLen)
			for i := range secret {
				secret[i] = byte(i)
			}
			shares, _ := Split(secret, tc.threshold, tc.userXs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Recover(shares)
			}
		})
	}
}

func BenchmarkGenerateDelta(b *testing.B) {
	benchmarkCases := []struct {
		name      string
		threshold int
		numDeltas int
	}{
		{"small", 2, 10},
		{"medium", 5, 50},
		{"large", 10, 100},
	}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				GenerateDelta(tc.threshold, tc.numDeltas)
			}
		})
	}
}
