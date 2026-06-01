package algorithm

import (
	"crypto/rand"
	"errors"
	"math/big"
)

type Share struct {
	X int64
	Y int64
}

type Config struct {
	Threshold int
	NumShares int
	Prime     int64
}

type Shamir struct {
	field     *Field
	threshold int
	numShares int
	coeffs    []int64
	version   int
}

func NewShamir(config *Config) (*Shamir, error) {
	if config.Threshold < 2 {
		return nil, errors.New("threshold must be at least 2")
	}
	if config.NumShares < config.Threshold {
		return nil, errors.New("numShares must be at least threshold")
	}
	if config.Prime < 255 {
		return nil, errors.New("prime must be at least 255")
	}

	field, err := NewField(config.Prime)
	if err != nil {
		return nil, err
	}

	return &Shamir{
		field:     field,
		threshold: config.Threshold,
		numShares: config.NumShares,
		coeffs:    make([]int64, 0),
		version:   1,
	}, nil
}

func (s *Shamir) Split(secret int64) ([]Share, error) {
	if secret >= s.field.prime {
		return nil, errors.New("secret must be less than prime")
	}

	s.coeffs = make([]int64, s.threshold)
	s.coeffs[0] = secret

	primeBig := big.NewInt(s.field.prime)
	for i := 1; i < s.threshold; i++ {
		coef, err := rand.Int(rand.Reader, primeBig)
		if err != nil {
			return nil, err
		}
		s.coeffs[i] = coef.Int64()
	}

	shares := make([]Share, s.numShares)
	for x := int64(1); x <= int64(s.numShares); x++ {
		y := s.evaluatePolynomial(x)
		shares[x-1] = Share{X: x, Y: y}
	}

	s.version++
	return shares, nil
}

func (s *Shamir) evaluatePolynomial(x int64) int64 {
	result := int64(0)
	power := int64(1)

	for _, coeff := range s.coeffs {
		result = s.field.Add(result, s.field.Mul(coeff, power))
		power = s.field.Mul(power, x)
	}

	return result
}

func (s *Shamir) Recover(shares []Share) (int64, error) {
	if len(shares) < s.threshold {
		return 0, errors.New("need at least threshold shares to recover")
	}

	secret := int64(0)

	for i := 0; i < len(shares); i++ {
		x_i := shares[i].X
		y_i := shares[i].Y

		numerator := int64(1)
		denominator := int64(1)

		for j := 0; j < len(shares); j++ {
			if i != j {
				x_j := shares[j].X
				numerator = s.field.Mul(numerator, s.field.Neg(x_j))
				diff := s.field.Sub(x_i, x_j)
				denominator = s.field.Mul(denominator, diff)
			}
		}

		denomInverse, err := s.field.Inv(denominator)
		if err != nil {
			return 0, errors.New("failed to compute modular inverse")
		}

		lagrangeCoeff := s.field.Mul(y_i, numerator)
		lagrangeCoeff = s.field.Mul(lagrangeCoeff, denomInverse)

		secret = s.field.Add(secret, lagrangeCoeff)
	}

	return secret, nil
}

func (s *Shamir) Reshare(shares []Share) ([]Share, error) {
	secret, err := s.Recover(shares)
	if err != nil {
		return nil, err
	}
	return s.Split(secret)
}

func (s *Shamir) GetThreshold() int {
	return s.threshold
}

func (s *Shamir) GetNumShares() int {
	return s.numShares
}

func (s *Shamir) GetPrime() int64 {
	return s.field.prime
}

func (s *Shamir) GetVersion() int {
	return s.version
}

func (s *Shamir) GetCoeffs() []int64 {
	return s.coeffs
}

func (s *Shamir) EvaluateAt(x int64) int64 {
	return s.evaluatePolynomial(x)
}

// AddShares 实现加法同态，将两个秘密的份额相加
// 结果对应于两个原始秘密的和
func (s *Shamir) AddShares(shares1, shares2 []Share) ([]Share, error) {
	if len(shares1) != len(shares2) {
		return nil, errors.New("shares must have the same length")
	}

	result := make([]Share, len(shares1))
	for i := range shares1 {
		if shares1[i].X != shares2[i].X {
			return nil, errors.New("shares must have the same x coordinates")
		}
		result[i] = Share{
			X: shares1[i].X,
			Y: s.field.Add(shares1[i].Y, shares2[i].Y),
		}
	}

	return result, nil
}

// MultiplySharesByConstant 实现常数乘法同态，将份额与常数相乘
// 结果对应于原始秘密与该常数的乘积
func (s *Shamir) MultiplySharesByConstant(shares []Share, constant int64) ([]Share, error) {
	result := make([]Share, len(shares))
	for i, share := range shares {
		result[i] = Share{
			X: share.X,
			Y: s.field.Mul(share.Y, constant),
		}
	}

	return result, nil
}
