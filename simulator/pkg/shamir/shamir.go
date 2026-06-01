package shamir

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// Share 表示一个秘密份额
type Share struct {
	X int64 // 份额的x坐标
	Y int64 // 份额的y坐标
}

// Config 表示Shamir算法的配置
type Config struct {
	threshold int
	numShares int
	prime     *big.Int
}

// Option 是配置选项函数
type Option func(*Config)

// WithThreshold 设置门限值
func WithThreshold(threshold int) Option {
	return func(c *Config) {
		c.threshold = threshold
	}
}

// WithNumShares 设置份额数量
func WithNumShares(numShares int) Option {
	return func(c *Config) {
		c.numShares = numShares
	}
}

// WithPrime 设置素数
func WithPrime(prime int64) Option {
	return func(c *Config) {
		c.prime = big.NewInt(prime)
	}
}

// Shamir 表示Shamir秘密共享算法的实例
type Shamir struct {
	config *Config
	coeffs []*big.Int
}

// New 创建一个新的Shamir实例
func New(opts ...Option) (*Shamir, error) {
	// 默认配置
	config := &Config{
		threshold: 3,
		numShares: 5,
		prime:     big.NewInt(257),
	}

	// 应用选项
	for _, opt := range opts {
		opt(config)
	}

	// 验证配置
	if config.threshold < 2 {
		return nil, errors.New("threshold must be at least 2")
	}

	if config.numShares < config.threshold {
		return nil, errors.New("numShares must be at least threshold")
	}

	if config.prime.Int64() < 255 {
		return nil, errors.New("prime must be at least 255")
	}

	return &Shamir{
		config: config,
		coeffs: make([]*big.Int, 0),
	}, nil
}

// Split 将秘密分割为多个份额
func (s *Shamir) Split(secret int64) ([]Share, error) {
	if secret >= s.config.prime.Int64() {
		return nil, errors.New("secret must be less than prime")
	}

	// 生成多项式系数
	s.coeffs = make([]*big.Int, s.config.threshold)
	s.coeffs[0] = big.NewInt(secret)

	for i := 1; i < s.config.threshold; i++ {
		coef, err := rand.Int(rand.Reader, s.config.prime)
		if err != nil {
			return nil, err
		}
		s.coeffs[i] = coef
	}

	// 生成份额
	shares := make([]Share, s.config.numShares)
	for x := int64(1); x <= int64(s.config.numShares); x++ {
		y := s.evaluatePolynomial(x)
		shares[x-1] = Share{X: x, Y: y}
	}

	return shares, nil
}

// evaluatePolynomial 计算多项式在x处的值
func (s *Shamir) evaluatePolynomial(x int64) int64 {
	result := big.NewInt(0)
	power := big.NewInt(1)

	for _, coeff := range s.coeffs {
		temp := new(big.Int).Mul(coeff, power)
		result.Add(result, temp)
		power.Mul(power, big.NewInt(x))
	}

	return new(big.Int).Mod(result, s.config.prime).Int64()
}

// Recover 从份额中恢复秘密
func (s *Shamir) Recover(shares []Share) (int64, error) {
	if len(shares) < s.config.threshold {
		return 0, errors.New("need at least threshold shares to recover")
	}

	secret := big.NewInt(0)

	for i := 0; i < len(shares); i++ {
		x_i := shares[i].X
		y_i := shares[i].Y

		numerator := big.NewInt(1)
		denominator := big.NewInt(1)

		for j := 0; j < len(shares); j++ {
			if i != j {
				x_j := shares[j].X

				negXj := big.NewInt(-x_j)
				numerator.Mul(numerator, negXj)

				diff := big.NewInt(x_i - x_j)
				denominator.Mul(denominator, diff)
			}
		}

		numerator.Mod(numerator, s.config.prime)
		denominator.Mod(denominator, s.config.prime)

		denomInverse := new(big.Int).ModInverse(denominator, s.config.prime)
		if denomInverse == nil {
			return 0, errors.New("failed to compute modular inverse")
		}

		lagrangeCoeff := new(big.Int).Mul(big.NewInt(y_i), numerator)
		lagrangeCoeff.Mul(lagrangeCoeff, denomInverse)
		lagrangeCoeff.Mod(lagrangeCoeff, s.config.prime)

		secret.Add(secret, lagrangeCoeff)
		secret.Mod(secret, s.config.prime)
	}

	return secret.Int64(), nil
}

// GetThreshold 获取门限值
func (s *Shamir) GetThreshold() int {
	return s.config.threshold
}

// GetNumShares 获取份额数量
func (s *Shamir) GetNumShares() int {
	return s.config.numShares
}

// GetPrime 获取素数
func (s *Shamir) GetPrime() int64 {
	return s.config.prime.Int64()
}

// Reshare 重新生成份额（动态重构）
func (s *Shamir) Reshare(shares []Share) ([]Share, error) {
	// 首先恢复原始秘密
	secret, err := s.Recover(shares)
	if err != nil {
		return nil, err
	}

	// 重新分割秘密
	return s.Split(secret)
}
