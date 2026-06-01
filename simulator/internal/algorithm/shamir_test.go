package algorithm

import (
	"testing"
)

func TestNewShamir(t *testing.T) {
	// 测试默认配置
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	if shamir.GetThreshold() != 3 {
		t.Errorf("Expected threshold 3, got %d", shamir.GetThreshold())
	}

	if shamir.GetNumShares() != 5 {
		t.Errorf("Expected numShares 5, got %d", shamir.GetNumShares())
	}

	if shamir.GetPrime() != 257 {
		t.Errorf("Expected prime 257, got %d", shamir.GetPrime())
	}

	// 测试无效配置
	invalidConfig := &Config{
		Threshold: 10,
		NumShares: 5,
		Prime:     257,
	}

	_, err = NewShamir(invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid threshold > numShares")
	}

	// 测试无效素数
	smallPrimeConfig := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     10,
	}

	_, err = NewShamir(smallPrimeConfig)
	if err == nil {
		t.Error("Expected error for small prime")
	}
}

func TestSplitAndRecover(t *testing.T) {
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	// 测试秘密分割和恢复
	secret := int64(42)
	shares, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split() failed: %v", err)
	}

	if len(shares) != 5 {
		t.Errorf("Expected 5 shares, got %d", len(shares))
	}

	// 从3个份额恢复秘密
	recovered, err := shamir.Recover(shares[:3])
	if err != nil {
		t.Fatalf("Recover() failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected secret %d, got %d", secret, recovered)
	}

	// 从4个份额恢复秘密
	recovered, err = shamir.Recover(shares[:4])
	if err != nil {
		t.Fatalf("Recover() with 4 shares failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected secret %d, got %d", secret, recovered)
	}

	// 测试份额不足的情况
	_, err = shamir.Recover(shares[:2])
	if err == nil {
		t.Error("Expected error for insufficient shares")
	}
}

func TestReshare(t *testing.T) {
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	secret := int64(42)
	shares, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split() failed: %v", err)
	}

	// 重新生成份额
	newShares, err := shamir.Reshare(shares[:3])
	if err != nil {
		t.Fatalf("Reshare() failed: %v", err)
	}

	if len(newShares) != 5 {
		t.Errorf("Expected 5 new shares, got %d", len(newShares))
	}

	// 从新份额中恢复秘密
	recovered, err := shamir.Recover(newShares[:3])
	if err != nil {
		t.Fatalf("Recover() from new shares failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected secret %d, got %d", secret, recovered)
	}

	// 测试份额不足的情况
	_, err = shamir.Reshare(shares[:2])
	if err == nil {
		t.Error("Expected error for insufficient shares in reshare")
	}
}

func TestMultipleSecrets(t *testing.T) {
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	// 测试多个秘密
	secrets := []int64{10, 20, 30, 40, 50}

	for _, secret := range secrets {
		shares, err := shamir.Split(secret)
		if err != nil {
			t.Fatalf("Split() failed for secret %d: %v", secret, err)
		}

		recovered, err := shamir.Recover(shares[:3])
		if err != nil {
			t.Fatalf("Recover() failed for secret %d: %v", secret, err)
		}

		if recovered != secret {
			t.Errorf("Expected secret %d, got %d", secret, recovered)
		}
	}
}

func TestLargePrime(t *testing.T) {
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     65537, // 较大的素数
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	secret := int64(12345)
	shares, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split() failed: %v", err)
	}

	recovered, err := shamir.Recover(shares[:3])
	if err != nil {
		t.Fatalf("Recover() failed: %v", err)
	}

	if recovered != secret {
		t.Errorf("Expected secret %d, got %d", secret, recovered)
	}
}

func TestAdditiveHomomorphism(t *testing.T) {
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	// 测试两个秘密的加法
	secret1 := int64(42)
	secret2 := int64(123)
	expectedSum := secret1 + secret2

	shares1, err := shamir.Split(secret1)
	if err != nil {
		t.Fatalf("Split() for secret1 failed: %v", err)
	}

	shares2, err := shamir.Split(secret2)
	if err != nil {
		t.Fatalf("Split() for secret2 failed: %v", err)
	}

	// 相加份额
	sumShares, err := shamir.AddShares(shares1, shares2)
	if err != nil {
		t.Fatalf("AddShares() failed: %v", err)
	}

	// 从和的份额中恢复秘密
	recoveredSum, err := shamir.Recover(sumShares[:3])
	if err != nil {
		t.Fatalf("Recover() for sum failed: %v", err)
	}

	if recoveredSum != expectedSum {
		t.Errorf("Expected sum %d, got %d", expectedSum, recoveredSum)
	}
}

func TestConstantMultiplicationHomomorphism(t *testing.T) {
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	// 测试常数乘法
	secret := int64(50)
	constant := int64(3)
	expectedProduct := secret * constant

	shares, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split() failed: %v", err)
	}

	// 与常数相乘
	productShares, err := shamir.MultiplySharesByConstant(shares, constant)
	if err != nil {
		t.Fatalf("MultiplySharesByConstant() failed: %v", err)
	}

	// 从乘积的份额中恢复秘密
	recoveredProduct, err := shamir.Recover(productShares[:3])
	if err != nil {
		t.Fatalf("Recover() for product failed: %v", err)
	}

	if recoveredProduct != expectedProduct {
		t.Errorf("Expected product %d, got %d", expectedProduct, recoveredProduct)
	}
}

func TestAddSharesErrorCases(t *testing.T) {
	config := &Config{
		Threshold: 3,
		NumShares: 5,
		Prime:     257,
	}

	shamir, err := NewShamir(config)
	if err != nil {
		t.Fatalf("NewShamir() failed: %v", err)
	}

	secret := int64(42)
	shares1, err := shamir.Split(secret)
	if err != nil {
		t.Fatalf("Split() failed: %v", err)
	}

	// 测试长度不同的情况
	shares2 := shares1[:3] // 截取部分份额
	_, err = shamir.AddShares(shares1, shares2)
	if err == nil {
		t.Error("Expected error for different lengths")
	}

	// 测试x坐标不同的情况
	shares3 := make([]Share, len(shares1))
	copy(shares3, shares1)
	shares3[0].X = 999 // 修改x坐标
	_, err = shamir.AddShares(shares1, shares3)
	if err == nil {
		t.Error("Expected error for different x coordinates")
	}
}
