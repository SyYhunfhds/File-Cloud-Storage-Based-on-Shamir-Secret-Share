package shamir

import (
	"fmt"
	"math"
	"testing"
)

func TestNewShamir(t *testing.T) {
	shamir, err := NewShamir(3, 5, 257)
	if err != nil {
		t.Fatalf("Failed to create Shamir: %v", err)
	}
	if shamir.threshold != 3 {
		t.Errorf("Expected threshold 3, got %d", shamir.threshold)
	}
	if shamir.numShares != 5 {
		t.Errorf("Expected numShares 5, got %d", shamir.numShares)
	}
	if shamir.prime.Int64() != 257 {
		t.Errorf("Expected prime 257, got %d", shamir.prime.Int64())
	}
}

func TestNewShamirInvalidPrime(t *testing.T) {
	_, err := NewShamir(3, 5, 100)
	if err == nil {
		t.Error("Expected error for prime < 255")
	}
}

func TestSplitAndRecover(t *testing.T) {
	testCases := []struct {
		secret    int64
		threshold int
		numShares int
		prime     int64
	}{
		{42, 3, 5, 257},
		{123, 3, 5, 257},
		{0, 3, 5, 257},
		{255, 3, 5, 257},
		{200, 4, 6, 257},
		{128, 2, 4, 257},
		{100, 3, 5, 257},
		{1, 2, 3, 257},
		{256, 3, 5, 521},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("secret=%d_t=%d_n=%d_p=%d", tc.secret, tc.threshold, tc.numShares, tc.prime), func(t *testing.T) {
			shamir, _ := NewShamir(tc.threshold, tc.numShares, tc.prime)
			shares, err := shamir.Split(tc.secret)
			if err != nil {
				t.Fatalf("Split failed: %v", err)
			}

			if len(shares) != tc.numShares {
				t.Errorf("Expected %d shares, got %d", tc.numShares, len(shares))
			}

			recovered, err := shamir.Recover(shares[:tc.threshold])
			if err != nil {
				t.Fatalf("Recover failed: %v", err)
			}

			if recovered != tc.secret {
				t.Errorf("Expected secret %d, got %d", tc.secret, recovered)
			}
		})
	}
}

func TestRecoverWithDifferentSubset(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(123)

	shares, _ := shamir.Split(secret)

	combinations := [][]int{
		{0, 1, 2},
		{0, 1, 3},
		{0, 1, 4},
		{0, 2, 3},
		{0, 2, 4},
		{1, 2, 3},
		{1, 2, 4},
		{2, 3, 4},
	}

	for _, combo := range combinations {
		recoverShares := make([][2]int64, 3)
		for i, idx := range combo {
			recoverShares[i] = shares[idx]
		}
		recovered, err := shamir.Recover(recoverShares)
		if err != nil {
			t.Fatalf("Recover failed for combination %v: %v", combo, err)
		}
		if recovered != secret {
			t.Errorf("Recovered %d, expected %d for combination %v", recovered, secret, combo)
		}
	}
}

func TestRecoverInsufficientShares(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 257)
	shares, _ := shamir.Split(42)

	_, err := shamir.Recover(shares[:2])
	if err == nil {
		t.Error("Expected error for insufficient shares")
	}
}

func TestSplitSecretTooLarge(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 257)
	_, err := shamir.Split(300)
	if err == nil {
		t.Error("Expected error for secret >= prime")
	}
}

func TestRecoverWithMoreThanThreshold(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(200)

	shares, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	recovered, err := shamir.Recover(shares[:4])
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected %d, got %d", secret, recovered)
	}
}

func TestShareUniqueness(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(123)

	shares1, _ := shamir.Split(secret)
	shares2, _ := shamir.Split(secret)

	for i := 0; i < len(shares1); i++ {
		if shares1[i][0] != shares2[i][0] {
			t.Error("Share x-values should be the same")
		}
	}

	different := false
	for i := 0; i < len(shares1); i++ {
		if shares1[i][1] != shares2[i][1] {
			different = true
			break
		}
	}
	if !different {
		t.Error("Two splits of the same secret should produce different shares (random coefficients)")
	}
}

func TestGetters(t *testing.T) {
	shamir, _ := NewShamir(4, 7, 521)

	if shamir.GetPrime() != 521 {
		t.Errorf("Expected prime 521, got %d", shamir.GetPrime())
	}
	if shamir.GetThreshold() != 4 {
		t.Errorf("Expected threshold 4, got %d", shamir.GetThreshold())
	}
	if shamir.GetNumShares() != 7 {
		t.Errorf("Expected numShares 7, got %d", shamir.GetNumShares())
	}
}

func TestLargePrime(t *testing.T) {
	shamir, _ := NewShamir(5, 10, 10007)
	secret := int64(9999)

	shares, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	recovered, err := shamir.Recover(shares[:5])
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected %d, got %d", secret, recovered)
	}
}

func TestRecoverWithExcessShares(t *testing.T) {
	shamir, _ := NewShamir(2, 5, 257)
	secret := int64(77)

	shares, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	recovered, err := shamir.Recover(shares[:4])
	if err != nil {
		t.Fatalf("Recover failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected %d, got %d", secret, recovered)
	}
}

func TestRecoverWithNoisyShares(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(123)

	shares, _ := shamir.Split(secret)

	errorLevels := []int64{1, 2, 5, 10, 20}

	for _, errorLevel := range errorLevels {
		t.Run(fmt.Sprintf("errorLevel=%d", errorLevel), func(t *testing.T) {
			nosyShares := make([][2]int64, len(shares))
			for i, share := range shares {
				nosyShares[i] = [2]int64{share[0], (share[1] + errorLevel) % 257}
			}

			recovered, err := shamir.Recover(nosyShares[:3])
			if err != nil {
				t.Fatalf("Recover failed: %v", err)
			}

			error := int64(math.Abs(float64(recovered - secret)))
			fmt.Printf("Input error: %d, Output error: %d, Recovered: %d, Expected: %d\n", errorLevel, error, recovered, secret)

			if recovered == secret {
				t.Logf("No error detected despite input noise - this is expected in some cases")
			} else {
				t.Logf("Error detected as expected: input error %d, output error %d", errorLevel, error)
			}
		})
	}
}

func TestRecoverWithVaryingNoiseLevels(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 10007)
	secret := int64(5000)

	shares, _ := shamir.Split(secret)

	noiseLevels := []int64{1, 10, 100, 1000, 2000}
	results := make(map[int64]int64)

	for _, noise := range noiseLevels {
		nosyShares := make([][2]int64, len(shares))
		for i, share := range shares {
			nosyShares[i] = [2]int64{share[0], (share[1] + noise) % 10007}
		}

		recovered, _ := shamir.Recover(nosyShares[:3])
		error := int64(math.Abs(float64(recovered - secret)))
		results[noise] = error

		fmt.Printf("Noise level: %d, Recovered: %d, Expected: %d, Error: %d\n", noise, recovered, secret, error)
	}

	for noise, error := range results {
		fmt.Printf("Input noise: %d, Output error: %d\n", noise, error)
	}
}

func TestRecoverWithSingleNoisyShare(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(100)

	shares, _ := shamir.Split(secret)

	for noisyIndex := 0; noisyIndex < 3; noisyIndex++ {
		t.Run(fmt.Sprintf("noisyShareIndex=%d", noisyIndex), func(t *testing.T) {
			nosyShares := make([][2]int64, 3)
			for i, share := range shares[:3] {
				if i == noisyIndex {
					nosyShares[i] = [2]int64{share[0], (share[1] + 50) % 257}
				} else {
					nosyShares[i] = share
				}
			}

			recovered, err := shamir.Recover(nosyShares)
			if err != nil {
				t.Fatalf("Recover failed: %v", err)
			}

			error := int64(math.Abs(float64(recovered - secret)))
			fmt.Printf("Noisy share index: %d, Recovered: %d, Expected: %d, Error: %d\n", noisyIndex, recovered, secret, error)

			if recovered != secret {
				t.Logf("Error detected as expected: recovered %d != %d", recovered, secret)
			} else {
				t.Logf("No error detected despite noisy share - this is possible due to modular arithmetic")
			}
		})
	}
}

func TestErrorPropagationAnalysis(t *testing.T) {
	shamir, _ := NewShamir(3, 5, 1000003)
	secret := int64(500000)

	shares, _ := shamir.Split(secret)

	baseShares := shares[:3]

	fmt.Println("=== Error Propagation Analysis ===")
	fmt.Printf("Original secret: %d\n", secret)
	fmt.Println("Base shares:")
	for i, share := range baseShares {
		fmt.Printf("  Share %d: (%d, %d)\n", i+1, share[0], share[1])
	}

	for errorMagnitude := int64(100); errorMagnitude <= 10000; errorMagnitude *= 10 {
		nosyShares := make([][2]int64, 3)
		for i, share := range baseShares {
			nosyShares[i] = [2]int64{share[0], (share[1] + errorMagnitude) % 1000003}
		}

		recovered, _ := shamir.Recover(nosyShares)
		outputError := int64(math.Abs(float64(recovered - secret)))
		inputError := errorMagnitude * 3

		fmt.Printf("\nInput error: %d (total), Output error: %d\n", inputError, outputError)
		fmt.Printf("Recovered: %d, Expected: %d\n", recovered, secret)
	}
}
