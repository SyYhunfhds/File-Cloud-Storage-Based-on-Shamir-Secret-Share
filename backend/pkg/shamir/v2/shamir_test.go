package shamir

import (
	"bytes"
	"testing"
	"time"
)

type BasicTestCase struct {
	Name      string
	Secret    []byte
	Threshold int
	UserXs    []uint32
	Seed      uint64
}

func TestBasicSplitRecover(t *testing.T) {
	testCases := []BasicTestCase{
		{
			Name:      "4-byte-secret",
			Secret:    []byte("test"),
			Threshold: 2,
			UserXs:    []uint32{1001, 2002, 3003},
			Seed:      123456789,
		},
		{
			Name:      "8-byte-secret",
			Secret:    []byte("testing8"),
			Threshold: 3,
			UserXs:    []uint32{5, 17, 23, 42, 99},
			Seed:      987654321,
		},
		{
			Name:      "12-byte-secret",
			Secret:    []byte("123456789012"),
			Threshold: 2,
			UserXs:    []uint32{101, 202, 303, 404},
			Seed:      111222333,
		},
		{
			Name:      "16-byte-secret",
			Secret:    []byte("aes-key-16bytes!"),
			Threshold: 3,
			UserXs:    []uint32{7, 13, 19, 29, 37},
			Seed:      444555666,
		},
		{
			Name:      "5-byte-secret",
			Secret:    []byte("hello"),
			Threshold: 2,
			UserXs:    []uint32{100, 200, 300},
			Seed:      777888999,
		},
		{
			Name:      "7-byte-secret",
			Secret:    []byte("abcdefg"),
			Threshold: 2,
			UserXs:    []uint32{501, 502, 503},
			Seed:      1011121314,
		},
		{
			Name:      "1-byte-secret",
			Secret:    []byte("x"),
			Threshold: 2,
			UserXs:    []uint32{1, 2, 3},
			Seed:      1516171819,
		},
		{
			Name:      "empty-secret",
			Secret:    []byte(""),
			Threshold: 2,
			UserXs:    []uint32{10, 20, 30},
			Seed:      2021222324,
		},
		{
			Name:      "32-byte-secret",
			Secret:    []byte("this-is-a-32-byte-secret-key!!"),
			Threshold: 5,
			UserXs:    []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			Seed:      111222333444,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			testBasicSplitRecover(t, tc)
		})
	}
}

func testBasicSplitRecover(t *testing.T, tc BasicTestCase) {
	shares := Split(tc.Secret, tc.Threshold, tc.UserXs, tc.Seed)

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

type SequentialTestCase struct {
	Name      string
	Secret    []byte
	Threshold int
	UserXs    []uint32
	Seed      uint64
}

func TestSequentialShareGeneration(t *testing.T) {
	testCases := []SequentialTestCase{
		{
			Name:      "4-byte-sequential",
			Secret:    []byte("test"),
			Threshold: 2,
			UserXs:    []uint32{1001, 2002, 3003},
			Seed:      123456789,
		},
		{
			Name:      "8-byte-sequential",
			Secret:    []byte("testing8"),
			Threshold: 3,
			UserXs:    []uint32{5, 17, 23, 42, 99},
			Seed:      987654321,
		},
		{
			Name:      "5-byte-sequential",
			Secret:    []byte("hello"),
			Threshold: 2,
			UserXs:    []uint32{100, 200, 300},
			Seed:      777888999,
		},
		{
			Name:      "empty-secret-sequential",
			Secret:    []byte(""),
			Threshold: 2,
			UserXs:    []uint32{10, 20, 30},
			Seed:      2021222324,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			testSequentialShareGeneration(t, tc)
		})
	}
}

func testSequentialShareGeneration(t *testing.T, tc SequentialTestCase) {
	var shares ShareVector

	for _, userX := range tc.UserXs {
		newShare := GenerateShare(tc.Secret, tc.Threshold, userX, tc.Seed)
		shares = append(shares, newShare)
	}

	expectedBlocks := (len(tc.Secret) + 3) / 4
	if len(tc.Secret) == 0 {
		expectedBlocks = 1
	}

	if len(shares) != len(tc.UserXs) {
		t.Errorf("expected %d shares, got %d", len(tc.UserXs), len(shares))
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
	seed := uint64(123456789)

	shares := Split(secret, threshold, userXs, seed)
	originalRecovered := Recover(shares)

	deltaSeed := uint64(987654321)
	for i, userX := range userXs {
		delta := CalculateSingleDelta(deltaSeed, threshold, userX)
		for j := range shares[i].Values {
			shares[i].Values[j] = add(shares[i].Values[j], delta)
		}
	}

	updatedRecovered := Recover(shares)
	if !bytes.Equal(updatedRecovered, originalRecovered) {
		t.Errorf("secret changed after delta update: expected %v, got %v", originalRecovered, updatedRecovered)
	}
}

func TestCalculateSingleDelta(t *testing.T) {
	testCases := []struct {
		Name      string
		Seed      uint64
		Threshold int
		UserX     uint32
	}{
		{"delta-test-1", 123456789, 3, 1001},
		{"delta-test-2", 987654321, 5, 2002},
		{"delta-test-3", 111222333, 2, 3003},
		{"delta-test-4", 444555666, 4, 4004},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			delta1 := CalculateSingleDelta(tc.Seed, tc.Threshold, tc.UserX)
			delta2 := CalculateSingleDelta(tc.Seed, tc.Threshold, tc.UserX)

			if delta1 != delta2 {
				t.Errorf("delta not deterministic: got %d and %d", delta1, delta2)
			}

			if tc.UserX == 0 {
				if delta1 != 0 {
					t.Errorf("delta at x=0 should be 0, got %d", delta1)
				}
			}
		})
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
		{"7-byte", []byte("abcdefg"), 8},
		{"8-byte", []byte("testing8"), 8},
		{"1-byte", []byte("x"), 4},
		{"2-byte", []byte("xy"), 4},
		{"3-byte", []byte("xyz"), 4},
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

func BenchmarkSplit(b *testing.B) {
	benchmarkCases := []struct {
		name      string
		secretLen int
		threshold int
		userXs    []uint32
	}{
		{"small-secret-small-share", 4, 2, []uint32{1, 2, 3}},
		{"small-secret-medium-share", 4, 2, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"small-secret-large-share", 4, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
		{"medium-secret-small-share", 16, 3, []uint32{1, 2, 3, 4, 5}},
		{"medium-secret-medium-share", 16, 3, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"medium-secret-large-share", 16, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
		{"large-secret-small-share", 64, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"large-secret-medium-share", 64, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
		{"large-secret-large-share", 64, 10, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30}},
	}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			secret := make([]byte, tc.secretLen)
			for i := range secret {
				secret[i] = byte(i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Split(secret, tc.threshold, tc.userXs, uint64(i))
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
		{"small-secret-medium-share", 4, 2, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"medium-secret-small-share", 16, 3, []uint32{1, 2, 3, 4, 5}},
		{"medium-secret-medium-share", 16, 3, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"large-secret-small-share", 64, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"large-secret-large-share", 64, 10, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30}},
	}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			secret := make([]byte, tc.secretLen)
			for i := range secret {
				secret[i] = byte(i)
			}
			shares := Split(secret, tc.threshold, tc.userXs, 12345)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Recover(shares)
			}
		})
	}
}

func BenchmarkCalculateSingleDelta(b *testing.B) {
	benchmarkCases := []struct {
		name      string
		threshold int
	}{
		{"small", 2},
		{"medium", 5},
		{"large", 10},
		{"extra-large", 20},
	}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				CalculateSingleDelta(uint64(i), tc.threshold, uint32(i+1))
			}
		})
	}
}

func BenchmarkSequentialShareGeneration(b *testing.B) {
	benchmarkCases := []struct {
		name      string
		secretLen int
		threshold int
		userXs    []uint32
	}{
		{"small", 4, 2, []uint32{1, 2, 3, 4, 5}},
		{"medium", 16, 3, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		{"large", 64, 5, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
	}

	for _, tc := range benchmarkCases {
		b.Run(tc.name, func(b *testing.B) {
			secret := make([]byte, tc.secretLen)
			for i := range secret {
				secret[i] = byte(i)
			}

			b.ResetTimer()
			for iter := 0; iter < b.N; iter++ {
				var shares ShareVector
				for _, userX := range tc.userXs {
					newShare := GenerateShare(secret, tc.threshold, userX, uint64(iter))
					shares = append(shares, newShare)
				}
			}
		})
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
				Split(secret, tc.threshold, tc.userXs, uint64(i))
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
