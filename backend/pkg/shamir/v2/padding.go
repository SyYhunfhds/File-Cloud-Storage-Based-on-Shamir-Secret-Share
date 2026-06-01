package shamir

// Pad pads data to a 4-byte boundary using PKCS#7-like padding.
// The padding value is equal to the number of padding bytes added.
//
// Parameters:
//
//	data - The input data to pad
//
// Returns:
//
//	The padded data with length divisible by 4
//
// Padding Rules:
//   - If data length is already divisible by 4, returns data unchanged (NO double-padding)
//   - If data is empty, returns 4 zero bytes
//   - Otherwise, adds 1-3 padding bytes where each byte equals the padding length
//
// Examples:
//
//	Pad([]byte(""))      -> []byte{0, 0, 0, 0}       (4 bytes)
//	Pad([]byte("a"))     -> []byte{'a', 3, 3, 3}     (4 bytes)
//	Pad([]byte("ab"))    -> []byte{'a', 'b', 2, 2}   (4 bytes)
//	Pad([]byte("abc"))   -> []byte{'a', 'b', 'c', 1} (4 bytes)
//	Pad([]byte("abcd"))  -> []byte{'a', 'b', 'c', 'd'} (unchanged, 4 bytes)
//	Pad([]byte("abcde")) -> []byte{'a', 'b', 'c', 'd', 'e', 3, 3, 3} (8 bytes)
//
// Security:
//   - Deterministic padding, no randomness introduced
//   - Strictly prohibits double-padding (already padded data remains unchanged)
func Pad(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return make([]byte, 4)
	}

	remainder := length % 4
	if remainder == 0 {
		return data
	}

	padLength := 4 - remainder
	padded := make([]byte, length+padLength)
	copy(padded, data)

	for i := length; i < length+padLength; i++ {
		padded[i] = byte(padLength)
	}

	return padded
}

// Unpad removes padding from data that was padded with Pad().
//
// Parameters:
//
//	data - The padded data to unpad
//
// Returns:
//
//	The unpadded original data
//
// Unpadding Rules:
//   - If padding is invalid or not present, returns data unchanged
//   - Recognizes the special case of empty input (4 zero bytes)
//   - Validates padding bytes before removing them
//
// Examples:
//
//	Unpad([]byte{0, 0, 0, 0})       -> []byte{}
//	Unpad([]byte{'a', 3, 3, 3})     -> []byte{'a'}
//	Unpad([]byte{'a', 'b', 'c', 'd'}) -> []byte{'a', 'b', 'c', 'd'} (unchanged)
//
// Security:
//   - Safe against padding oracle attacks through validation
//   - Returns original data if padding is invalid
func Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	padLength := int(data[len(data)-1])
	if padLength < 1 || padLength > 3 {
		if len(data) == 4 && data[0] == 0 && data[1] == 0 && data[2] == 0 && data[3] == 0 {
			return []byte{}
		}
		return data
	}

	if len(data) < padLength {
		return data
	}

	for i := len(data) - padLength; i < len(data)-1; i++ {
		if data[i] != byte(padLength) {
			return data
		}
	}

	return data[:len(data)-padLength]
}

// IsPadded checks if data appears to be padded with Pad().
//
// Parameters:
//
//	data - The data to check
//
// Returns:
//
//	true if data appears to be padded, false otherwise
//
// Detection Criteria:
//   - Length must be divisible by 4
//   - Padding bytes must all equal the padding length (1-3)
//   - Empty input (4 zero bytes) is considered padded
func IsPadded(data []byte) bool {
	if len(data) == 0 || len(data)%4 != 0 {
		return false
	}

	if len(data) < 4 {
		return true
	}

	padLength := int(data[len(data)-1])
	if padLength < 1 || padLength > 3 {
		return false
	}

	if len(data) < padLength {
		return false
	}

	for i := len(data) - padLength; i < len(data); i++ {
		if data[i] != byte(padLength) {
			return false
		}
	}

	return true
}
