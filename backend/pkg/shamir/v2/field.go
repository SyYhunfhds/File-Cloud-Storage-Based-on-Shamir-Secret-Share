package shamir

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
)

// prime is the prime modulus for the finite field GF(2^32-5).
// Value: 2^32 - 5 = 4294967291
//
// Choice rationale:
// - Large prime suitable for cryptographic applications
// - Fits in uint32 (max uint32 = 4294967295)
// - Multiplication of two uint32 values fits in uint64 without overflow
// - Supports more than 256 shares (unlike GF(256))
const (
	prime = 4294967291
)

// add performs addition in GF(2^32-5).
// Adds two uint32 values and reduces modulo prime if necessary.
//
// Parameters:
//
//	a, b - The values to add
//
// Returns:
//
//	(a + b) mod prime
func add(a, b uint32) uint32 {
	sum := uint64(a) + uint64(b)
	if sum >= prime {
		sum -= prime
	}
	return uint32(sum)
}

// sub performs subtraction in GF(2^32-5).
// Subtracts b from a and ensures the result is non-negative modulo prime.
//
// Parameters:
//
//	a - The minuend
//	b - The subtrahend
//
// Returns:
//
//	(a - b) mod prime
//
// Implementation:
//
//	Uses branch-based logic to avoid underflow issues. When a >= b, direct subtraction
//	is used. When a < b, the result is calculated as prime - (b - a), which avoids
//	any bit-level wrap-around that could occur with unsigned integer subtraction.
func sub(a, b uint32) uint32 {
	if a >= b {
		return a - b
	}
	return prime - (b - a)
}

// mul performs multiplication in GF(2^32-5).
// Multiplies two uint32 values and reduces modulo prime.
//
// Parameters:
//
//	a, b - The values to multiply
//
// Returns:
//
//	(a * b) mod prime
//
// Implementation Note:
//
//	Uses uint64 intermediate to avoid overflow during multiplication.
func mul(a, b uint32) uint32 {
	product := uint64(a) * uint64(b)
	return reduce(product)
}

// reduce reduces a uint64 value modulo prime to fit in uint32.
//
// Parameters:
//
//	x - The uint64 value to reduce
//
// Returns:
//
//	x mod prime as uint32
func reduce(x uint64) uint32 {
	return uint32(x % uint64(prime))
}

// pow computes exponentiation in GF(2^32-5) using fast exponentiation.
//
// Parameters:
//
//	a - The base
//	b - The exponent
//
// Returns:
//
//	a^b mod prime
//
// Implementation:
//
//	Uses binary exponentiation (exponentiation by squaring) for O(log b) complexity.
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
//
// Parameters:
//
//	a - The value to find the inverse of
//
// Returns:
//
//	a^(-1) mod prime
//
// Mathematical Basis:
//
//	Fermat's Little Theorem states that for prime p and a not divisible by p:
//	a^(p-1) ≡ 1 (mod p)
//	Therefore: a^(p-2) ≡ a^(-1) (mod p)
func inv(a uint32) uint32 {
	return pow(a, prime-2)
}

// div performs division in GF(2^32-5).
// Division is implemented as multiplication by the multiplicative inverse.
//
// Parameters:
//
//	a - The dividend
//	b - The divisor
//
// Returns:
//
//	(a / b) mod prime = a * b^(-1) mod prime
func div(a, b uint32) uint32 {
	return mul(a, inv(b))
}

// evalPolynomial evaluates a polynomial at a given point using Horner's method.
//
// Parameters:
//
//	coeffs - The polynomial coefficients, where coeffs[0] is the constant term
//	x - The point at which to evaluate the polynomial
//
// Returns:
//
//	The value of the polynomial at x
//
// Polynomial Form:
//
//	f(x) = coeffs[0] + coeffs[1]*x + coeffs[2]*x^2 + ... + coeffs[n-1]*x^(n-1)
//
// Implementation:
//
//	Uses Horner's method for O(n) complexity:
//	f(x) = coeffs[0] + x*(coeffs[1] + x*(coeffs[2] + ...))
func evalPolynomial(coeffs []uint32, x uint32) uint32 {
	result := uint32(0)
	for i := len(coeffs) - 1; i >= 0; i-- {
		result = add(mul(result, x), coeffs[i])
	}
	return result
}

// lagrangeInterpolation performs Lagrange interpolation to reconstruct a value.
// Given a set of points (xValues[i], yValues[i]), finds the value at x.
//
// Parameters:
//
//	x - The point at which to interpolate
//	xValues - The x-coordinates of the known points
//	yValues - The y-coordinates of the known points
//
// Returns:
//
//	The interpolated value at x
//
// Mathematical Formula:
//
//	L(x) = Σ yValues[i] * Π (x - xValues[j])/(xValues[i] - xValues[j]) for j != i
//
// Use Case:
//
//	Reconstructing the secret from shares (x=0 gives the secret)
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
// Uses HMAC-SHA256 with the provided seed for deterministic, secure randomness.
//
// Parameters:
//
//	seed - A uint64 seed for deterministic generation
//	count - The number of coefficients to generate
//
// Returns:
//
//	A slice of uint32 coefficients in the range [0, prime)
//
// Security:
//   - Uses HMAC-SHA256 for cryptographic security
//   - Deterministic: same seed produces same sequence
//   - Each coefficient is independent and uniformly distributed
func generateRandomCoefficients(seed uint64, count int) []uint32 {
	seedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(seedBytes, seed)

	h := hmac.New(sha256.New, seedBytes)
	coeffs := make([]uint32, count)

	for i := 0; i < count; i++ {
		counter := make([]byte, 4)
		binary.BigEndian.PutUint32(counter, uint32(i))
		h.Reset()
		h.Write(counter)
		digest := h.Sum(nil)

		randomVal := binary.BigEndian.Uint64(digest[:8])
		coeffs[i] = uint32(randomVal) % prime
	}
	return coeffs
}

// modPrime reduces a value modulo prime if it exceeds the prime.
//
// Parameters:
//
//	x - The value to reduce
//
// Returns:
//
//	x if x < prime, otherwise x mod prime
func modPrime(x uint32) uint32 {
	if x >= prime {
		return x % prime
	}
	return x
}
