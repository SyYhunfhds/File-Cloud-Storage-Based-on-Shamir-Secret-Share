package shamir

import (
	"testing"
)

func TestNew(t *testing.T) {
	// 测试默认配置
	s, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if s.GetThreshold() != 3 {
		t.Errorf("Expected threshold 3, got %d", s.GetThreshold())
	}

	if s.GetNumShares() != 5 {
		t.Errorf("Expected numShares 5, got %d", s.GetNumShares())
	}

	if s.GetPrime() != 257 {
		t.Errorf("Expected prime 257, got %d", s.GetPrime())
	}

	// 测试自定义配置
	s, err = New(
		WithThreshold(4),
		WithNumShares(6),
		WithPrime(1031),
	)
	if err != nil {
		t.Fatalf("New() with custom options failed: %v", err)
	}

	if s.GetThreshold() != 4 {
		t.Errorf("Expected threshold 4, got %d", s.GetThreshold())
	}

	if s.GetNumShares() != 6 {
		t.Errorf("Expected numShares 6, got %d", s.GetNumShares())
	}

	if s.GetPrime() != 1031 {
		t.Errorf("Expected prime 1031, got %d", s.GetPrime())
	}

	// 测试无效配置
	_, err = New(WithThreshold(1))
	if err == nil {
		t.Error("Expected error for threshold < 2, got nil")
	}

	_, err = New(WithThreshold(5), WithNumShares(3))
	if err == nil {
		t.Error("Expected error for numShares < threshold, got nil")
	}

	_, err = New(WithPrime(100))
	if err == nil {
		t.Error("Expected error for prime < 255, got nil")
	}
}

func TestSplitAndRecover(t *testing.T) {
	s, err := New(
		WithThreshold(3),
		WithNumShares(5),
		WithPrime(257),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	secret := int64(42)

	// 分割秘密
	shares, err := s.Split(secret)
	if err != nil {
		t.Fatalf("Split() failed: %v", err)
	}

	if len(shares) != 5 {
		t.Errorf("Expected 5 shares, got %d", len(shares))
	}

	// 测试恢复秘密（使用足够的份额）
	recovered, err := s.Recover(shares[:3])
	if err != nil {
		t.Fatalf("Recover() failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected secret %d, got %d", secret, recovered)
	}

	// 测试恢复秘密（使用所有份额）
	recovered, err = s.Recover(shares)
	if err != nil {
		t.Fatalf("Recover() with all shares failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected secret %d, got %d", secret, recovered)
	}

	// 测试恢复秘密（份额不足）
	_, err = s.Recover(shares[:2])
	if err == nil {
		t.Error("Expected error for insufficient shares, got nil")
	}

	// 测试秘密过大
	_, err = s.Split(257)
	if err == nil {
		t.Error("Expected error for secret >= prime, got nil")
	}
}

func TestReshare(t *testing.T) {
	s, err := New(
		WithThreshold(3),
		WithNumShares(5),
		WithPrime(257),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	secret := int64(123)

	// 分割秘密
	shares, err := s.Split(secret)
	if err != nil {
		t.Fatalf("Split() failed: %v", err)
	}

	// 重新生成份额
	newShares, err := s.Reshare(shares[:3])
	if err != nil {
		t.Fatalf("Reshare() failed: %v", err)
	}

	if len(newShares) != 5 {
		t.Errorf("Expected 5 new shares, got %d", len(newShares))
	}

	// 从新份额中恢复秘密
	recovered, err := s.Recover(newShares[:3])
	if err != nil {
		t.Fatalf("Recover() from new shares failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected secret %d, got %d", secret, recovered)
	}
}
