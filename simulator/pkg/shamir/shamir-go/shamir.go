package shamir

import (
	"crypto/rand"
	"errors"
	"math/big"
)

type Shamir struct {
	threshold int
	numShares int
	prime     *big.Int
	coeffs    []*big.Int
}

func NewShamir(threshold, numShares int, prime int64) (*Shamir, error) {
	if prime < 255 {
		return nil, errors.New("prime must be at least 255")
	}
	return &Shamir{
		threshold: threshold,
		numShares: numShares,
		prime:     big.NewInt(prime),
		coeffs:    make([]*big.Int, 0),
	}, nil
}

func (s *Shamir) Split(secret int64) (shares [][2]int64, err error) {
	if secret >= s.prime.Int64() {
		return nil, errors.New("secret must be less than prime")
	}

	s.coeffs = make([]*big.Int, s.threshold)
	s.coeffs[0] = big.NewInt(secret)

	for i := 1; i < s.threshold; i++ {
		coef, err := rand.Int(rand.Reader, s.prime)
		if err != nil {
			return nil, err
		}
		s.coeffs[i] = coef
	}

	shares = make([][2]int64, s.numShares)
	for x := int64(1); x <= int64(s.numShares); x++ {
		y := s.evaluatePolynomial(x)
		shares[x-1] = [2]int64{x, y}
	}

	return shares, nil
}

func (s *Shamir) evaluatePolynomial(x int64) int64 {
	result := big.NewInt(0)
	power := big.NewInt(1)

	for _, coeff := range s.coeffs {
		temp := new(big.Int).Mul(coeff, power)
		result.Add(result, temp)
		power.Mul(power, big.NewInt(x))
	}

	return new(big.Int).Mod(result, s.prime).Int64()
}

func (s *Shamir) Recover(shares [][2]int64) (int64, error) {
	if len(shares) < s.threshold {
		return 0, errors.New("need at least threshold shares to recover")
	}

	secret := big.NewInt(0)

	for i := 0; i < len(shares); i++ {
		x_i := shares[i][0]
		y_i := shares[i][1]

		numerator := big.NewInt(1)
		denominator := big.NewInt(1)

		for j := 0; j < len(shares); j++ {
			if i != j {
				x_j := shares[j][0]

				negXj := big.NewInt(-x_j)
				numerator.Mul(numerator, negXj)

				diff := big.NewInt(x_i - x_j)
				denominator.Mul(denominator, diff)
			}
		}

		numerator.Mod(numerator, s.prime)
		denominator.Mod(denominator, s.prime)

		denomInverse := new(big.Int).ModInverse(denominator, s.prime)
		if denomInverse == nil {
			return 0, errors.New("failed to compute modular inverse")
		}

		lagrangeCoeff := new(big.Int).Mul(big.NewInt(y_i), numerator)
		lagrangeCoeff.Mul(lagrangeCoeff, denomInverse)
		lagrangeCoeff.Mod(lagrangeCoeff, s.prime)

		secret.Add(secret, lagrangeCoeff)
		secret.Mod(secret, s.prime)
	}

	return secret.Int64(), nil
}

func (s *Shamir) GetPrime() int64 {
	return s.prime.Int64()
}

func (s *Shamir) GetThreshold() int {
	return s.threshold
}

func (s *Shamir) GetNumShares() int {
	return s.numShares
}
