/*
Package shamir implements Shamir's Secret Sharing scheme over GF(2^32-5).

Simplified v3 version - uses crypto/rand internally for random number generation.
No external seed input required.
*/
package shamir

import (
	"crypto/rand"
	"encoding/binary"
)

// Prime is the Prime modulus for GF(2^32-5): 2^32 - 5 = 4294967291
const Prime = 4294967291

// add performs addition in GF(2^32-5).
func add(a, b uint32) uint32 {
	sum := uint64(a) + uint64(b)
	if sum >= Prime {
		sum -= Prime
	}
	return uint32(sum)
}

// sub performs subtraction in GF(2^32-5) without underflow.
func sub(a, b uint32) uint32 {
	if a >= b {
		return a - b
	}
	return Prime - (b - a)
}

// mul performs multiplication in GF(2^32-5).
func mul(a, b uint32) uint32 {
	product := uint64(a) * uint64(b)
	return uint32(product % uint64(Prime))
}

// pow computes exponentiation in GF(2^32-5) using fast exponentiation.
func pow(a, b uint32) uint32 {
	result := uint32(1)
	base := a
	for b > 0 {
		if b&1 == 1 {
			result = mul(result, base)
		}
		base = mul(base, base)
		b >>= 1
	}
	return result
}

// inv computes the multiplicative inverse in GF(2^32-5).
// Uses Fermat's Little Theorem: a^(-1) = a^(p-2) mod p
func inv(a uint32) uint32 {
	return pow(a, Prime-2)
}

// div performs division in GF(2^32-5).
func div(a, b uint32) uint32 {
	return mul(a, inv(b))
}

// evalPolynomial evaluates a polynomial at x using Horner's method.
func evalPolynomial(coeffs []uint32, x uint32) uint32 {
	result := uint32(0)
	for i := len(coeffs) - 1; i >= 0; i-- {
		result = add(mul(result, x), coeffs[i])
	}
	return result
}

// lagrangeInterpolation performs Lagrange interpolation to find f(x).
func lagrangeInterpolation(x uint32, xValues []uint32, yValues []uint32) uint32 {
	result := uint32(0)
	n := len(xValues)

	for i := 0; i < n; i++ {
		term := yValues[i]
		for j := 0; j < n; j++ {
			if i != j {
				numerator := sub(x, xValues[j])
				denominator := sub(xValues[i], xValues[j])
				term = mul(term, div(numerator, denominator))
			}
		}
		result = add(result, term)
	}

	return result
}

// generateRandomCoefficients generates cryptographically secure random coefficients.
func generateRandomCoefficients(count int) ([]uint32, error) {
	coeffs := make([]uint32, count)
	for i := 0; i < count; i++ {
		var randVal uint64
		if err := binary.Read(rand.Reader, binary.BigEndian, &randVal); err != nil {
			return nil, err
		}
		coeffs[i] = uint32(randVal) % Prime
	}
	return coeffs, nil
}
