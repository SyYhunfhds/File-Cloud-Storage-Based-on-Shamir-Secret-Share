package shamir

import (
	"testing"
)

func BenchmarkSplit(b *testing.B) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(123)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Split(secret)
	}
}

func BenchmarkRecover(b *testing.B) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(123)
	shares, _ := shamir.Split(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Recover(shares[:3])
	}
}

func BenchmarkSplitAndRecover(b *testing.B) {
	shamir, _ := NewShamir(3, 5, 257)
	secret := int64(123)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shares, _ := shamir.Split(secret)
		_, _ = shamir.Recover(shares[:3])
	}
}

func BenchmarkSplitLargeThreshold(b *testing.B) {
	shamir, _ := NewShamir(10, 20, 257)
	secret := int64(123)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Split(secret)
	}
}

func BenchmarkRecoverLargeThreshold(b *testing.B) {
	shamir, _ := NewShamir(10, 20, 257)
	secret := int64(123)
	shares, _ := shamir.Split(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Recover(shares[:10])
	}
}

func BenchmarkSplitLargePrime(b *testing.B) {
	shamir, _ := NewShamir(5, 10, 10007)
	secret := int64(9999)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Split(secret)
	}
}

func BenchmarkRecoverLargePrime(b *testing.B) {
	shamir, _ := NewShamir(5, 10, 10007)
	secret := int64(9999)
	shares, _ := shamir.Split(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Recover(shares[:5])
	}
}

func BenchmarkSplitManyShares(b *testing.B) {
	shamir, _ := NewShamir(3, 100, 257)
	secret := int64(123)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Split(secret)
	}
}

func BenchmarkRecoverManyShares(b *testing.B) {
	shamir, _ := NewShamir(3, 100, 257)
	secret := int64(123)
	shares, _ := shamir.Split(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Recover(shares[:3])
	}
}

func BenchmarkSplitHugePrime(b *testing.B) {
	shamir, _ := NewShamir(3, 5, 16384)
	secret := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Split(secret)
	}
}

func BenchmarkRecoverHugePrime(b *testing.B) {
	shamir, _ := NewShamir(3, 5, 16384)
	secret := int64(12345)
	shares, _ := shamir.Split(secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = shamir.Recover(shares[:3])
	}
}

func BenchmarkSplitAndRecoverHugePrime(b *testing.B) {
	shamir, _ := NewShamir(3, 5, 16384)
	secret := int64(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shares, _ := shamir.Split(secret)
		_, _ = shamir.Recover(shares[:3])
	}
}
