/*
Package shamir implements Shamir's Secret Sharing scheme over GF(2^32-5).

Simplified v3 version features:
- No external seed input required (uses crypto/rand internally)
- GF(2^32-5) finite field for arbitrary number of shares
- Zero-value perturbation for secure share updates
- 4-byte padding for non-aligned input
- Cryptographically secure random number generation
*/
package shamir

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
)

func memclr(b []byte) {
	for i := 0; i < len(b); i++ {
		b[i] = 0
	}
}

// Share represents a single share with discrete user X coordinate.
type Share struct {
	Index  uint32   // User X coordinate (e.g., FNV-1a hash of user ID)
	Values []uint32 // Share values for each 4-byte block
}

func (s Share) ToBase64Bytes() (dest []byte, err error) {
	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	dest = make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(dest, data)
	return dest, nil
}

// ToBase64 业务特化需要, 反正Flutter没有本地解析份额的需要, 直接编码一下塞过去
func (s Share) ToBase64() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func (s Share) ToJSON() (string, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func ShareFromBase64(data string) (Share, error) {
	// Base64解码
	unencodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return Share{}, err
	}
	// Json序列化
	var share Share
	err = json.Unmarshal(unencodedBytes, &share)
	memclr(unencodedBytes)
	if err != nil {
		return Share{}, err
	}
	return share, nil
}
func ShareFromBase64Bytes(data []byte) (Share, error) {
	// Base64解码
	unencodedBytes := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	_, err := base64.StdEncoding.Decode(unencodedBytes, data)
	if err != nil {
		return Share{}, err
	}
	// Json序列化
	var share Share
	err = json.Unmarshal(unencodedBytes, &share)
	memclr(unencodedBytes)
	if err != nil {
		return Share{}, err
	}
	return share, nil
}

// ShareVector is a collection of shares.
type ShareVector []Share

// XValues extracts indices from all shares.
func (sv ShareVector) XValues() []uint32 {
	x := make([]uint32, len(sv))
	for i, s := range sv {
		x[i] = s.Index
	}
	return x
}

// YValuesForBlock extracts share values for a specific block.
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

// Split divides a secret into shares using provided user X coordinates.
// Uses crypto/rand internally for cryptographically secure randomness.
func Split(secret []byte, threshold int, userXs []uint32) (ShareVector, error) {
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

	for blockIdx := range secretWords {
		coeffs, err := generateRandomCoefficients(threshold - 1)
		if err != nil {
			return nil, err
		}
		coeffs = append([]uint32{secretWords[blockIdx]}, coeffs...)

		for i, x := range userXs {
			shares[i].Values[blockIdx] = evalPolynomial(coeffs, x)
		}
	}

	return shares, nil
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

// GenerateDelta creates a zero-value perturbation vector for secure share updates.
// The perturbation polynomial has g(0) = 0, ensuring the secret remains unchanged.
func GenerateDelta(threshold, numDeltas int) (DeltaVector, error) {
	delta := make(DeltaVector, numDeltas)

	coeffs, err := generateRandomCoefficients(threshold - 1)
	if err != nil {
		return nil, err
	}
	coeffs = append([]uint32{0}, coeffs...)

	for i := range delta {
		delta[i] = evalPolynomial(coeffs, uint32(i+1))
	}

	return delta, nil
}

// ApplyDelta applies a perturbation vector to shares.
func ApplyDelta(shares ShareVector, delta DeltaVector) ShareVector {
	result := make(ShareVector, len(shares))
	for i, s := range shares {
		newValues := make([]uint32, len(s.Values))
		for j, v := range s.Values {
			if i < len(delta) {
				newValues[j] = add(v, delta[i])
			} else {
				newValues[j] = v
			}
		}
		result[i] = Share{Index: s.Index, Values: newValues}
	}
	return result
}

// GenerateSingleShare creates a single share for a specific user X coordinate.
// Uses crypto/rand internally for cryptographically secure randomness.
func GenerateSingleShare(secret []byte, threshold int, userX uint32) (Share, error) {
	paddedSecret := Pad(secret)
	numBlocks := len(paddedSecret) / 4

	secretWords := make([]uint32, numBlocks)
	for i := 0; i < numBlocks; i++ {
		secretWords[i] = binary.BigEndian.Uint32(paddedSecret[i*4:])
	}

	values := make([]uint32, numBlocks)
	for blockIdx, word := range secretWords {
		coeffs, err := generateRandomCoefficients(threshold - 1)
		if err != nil {
			return Share{}, err
		}
		coeffs = append([]uint32{word}, coeffs...)
		values[blockIdx] = evalPolynomial(coeffs, userX)
	}

	return Share{Index: userX, Values: values}, nil
}

// GenerateUserXFromID generates a deterministic user X coordinate from a user ID.
// Uses FNV-1a hash and ensures non-zero result.
func GenerateUserXFromID(userID string) uint32 {
	hash := fnv1aHash(userID)
	if hash == 0 {
		hash = 1
	}
	return hash
}

// fnv1aHash computes FNV-1a hash of a string.
func fnv1aHash(s string) uint32 {
	h := uint32(2166136261)
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}
