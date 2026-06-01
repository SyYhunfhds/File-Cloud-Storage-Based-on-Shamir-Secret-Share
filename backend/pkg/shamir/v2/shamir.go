/*
Package shamir implements Shamir's Secret Sharing scheme over the finite field GF(2^32-5).

This is a business-specialized implementation with the following key features:
- Supports arbitrary number of shares (no 256 limit due to GF(2^32-5))
- Zero-value perturbation for secure share updates without changing the secret
- Cryptographically secure random number generation using HMAC-SHA256
- 4-byte padding for non-aligned input data (prohibits double-padding)
- External seed input for reproducibility and auditability

Algorithm Overview:
1. Split: Divides a secret into n shares, requiring t shares to recover
2. Recover: Uses Lagrange interpolation to reconstruct the secret
3. GenerateDelta: Creates a zero-value perturbation vector (g(0)=0)
4. ApplyDelta: Updates shares using the perturbation, secret remains unchanged

Security Properties:
- Information-theoretically secure: t-1 shares reveal no information about the secret
- Zero-value perturbation preserves the secret mathematically (g(0)=0)
- All randomness derived from external seed input (no internal state)
- Backend stores only perturbation history, not original polynomials or secrets

Field GF(2^32-5):
- Prime modulus: 2^32 - 5 = 4294967291
- Supports uint32 values natively
- Multiplication fits in uint64 without overflow
- Supports up to 2^32-6 shares theoretically
*/
package shamir

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
)

// Share represents a single share in the Shamir secret sharing scheme.
type Share struct {
	Index  uint32   // User index (1-based, must be unique)
	Values []uint32 // Share values for each 4-byte block
}

// ShareVector represents a collection of shares.
type ShareVector []Share

// XValues extracts the indices from all shares.
func (sv ShareVector) XValues() []uint32 {
	x := make([]uint32, len(sv))
	for i, s := range sv {
		x[i] = s.Index
	}
	return x
}

// YValuesForBlock extracts share values for a specific block index.
func (sv ShareVector) YValuesForBlock(blockIndex int) []uint32 {
	y := make([]uint32, len(sv))
	for i, s := range sv {
		if blockIndex < len(s.Values) {
			y[i] = s.Values[blockIndex]
		}
	}
	return y
}

// DeltaVector represents a perturbation vector for share updates.
type DeltaVector []uint32

// Hash computes SHA-256 hash of the delta vector.
func (dv DeltaVector) Hash() [32]byte {
	buf := make([]byte, len(dv)*4)
	for i, v := range dv {
		binary.BigEndian.PutUint32(buf[i*4:], v)
	}
	return sha256.Sum256(buf)
}

// Split divides a secret into shares using provided user indices.
// Each share is generated for a specific user X coordinate (discrete hash value).
func Split(secret []byte, threshold int, userXs []uint32, seed uint64) ShareVector {
	paddedSecret := Pad(secret)
	numBlocks := len(paddedSecret) / 4

	secretWords := make([]uint32, numBlocks)
	for i := 0; i < numBlocks; i++ {
		secretWords[i] = binary.BigEndian.Uint32(paddedSecret[i*4:])
	}

	shares := make(ShareVector, len(userXs))
	for i, x := range userXs {
		shares[i].Index = x
		shares[i].Values = make([]uint32, numBlocks)
	}

	blockSeeds := GenerateSeedSequence(seed, numBlocks)

	for blockIdx, word := range secretWords {
		blockSeed := blockSeeds[blockIdx]
		coeffs := generateRandomCoefficients(blockSeed, threshold-1)
		coeffs = append([]uint32{word}, coeffs...)

		for i, x := range userXs {
			y := evalPolynomial(coeffs, x)
			shares[i].Values[blockIdx] = y
		}
	}

	return shares
}

// Recover reconstructs the original secret from shares using Lagrange interpolation.
func Recover(shares ShareVector) []byte {
	if len(shares) == 0 {
		return nil
	}

	numBlocks := len(shares[0].Values)
	result := make([]byte, numBlocks*4)

	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		xValues := shares.XValues()
		yValues := shares.YValuesForBlock(blockIdx)
		secretWord := lagrangeInterpolation(uint32(0), xValues, yValues)
		binary.BigEndian.PutUint32(result[blockIdx*4:], secretWord)
	}

	return result
}

// GenerateSeedSequence generates a sequence of deterministic seeds.
func GenerateSeedSequence(seed uint64, count int) []uint64 {
	seedBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(seedBytes, seed)

	h := hmac.New(sha256.New, seedBytes)
	sequence := make([]uint64, count)

	for i := 0; i < count; i++ {
		counter := make([]byte, 4)
		binary.BigEndian.PutUint32(counter, uint32(i))
		h.Reset()
		h.Write(counter)
		digest := h.Sum(nil)
		sequence[i] = binary.BigEndian.Uint64(digest[:8])
	}
	return sequence
}

// GenerateShare creates a single share for a specific user X coordinate.
func GenerateShare(secret []byte, threshold int, userX uint32, seed uint64) Share {
	paddedSecret := Pad(secret)
	numBlocks := len(paddedSecret) / 4

	secretWords := make([]uint32, numBlocks)
	for i := 0; i < numBlocks; i++ {
		secretWords[i] = binary.BigEndian.Uint32(paddedSecret[i*4:])
	}

	blockSeeds := GenerateSeedSequence(seed, numBlocks)

	values := make([]uint32, numBlocks)
	for blockIdx, word := range secretWords {
		blockSeed := blockSeeds[blockIdx]
		coeffs := generateRandomCoefficients(blockSeed, threshold-1)
		coeffs = append([]uint32{word}, coeffs...)
		values[blockIdx] = evalPolynomial(coeffs, userX)
	}

	return Share{Index: userX, Values: values}
}

// CalculateSingleDelta computes a single delta value for a specific user X coordinate.
// This is used for stateless share updates where only the delta is needed for a specific user.
// The perturbation polynomial g(x) has g(0) = 0, ensuring the secret remains unchanged.
func CalculateSingleDelta(seed uint64, threshold int, userX uint32) uint32 {
	coeffs := generateRandomCoefficients(seed, threshold-1)
	coeffs = append([]uint32{0}, coeffs...)
	return evalPolynomial(coeffs, userX)
}
