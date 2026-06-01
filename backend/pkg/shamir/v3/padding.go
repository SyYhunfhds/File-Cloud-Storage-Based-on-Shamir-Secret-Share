package shamir

// Pad pads data to a 4-byte boundary using PKCS#7-like padding.
// If data is already 4-byte aligned, returns it unchanged (no double-padding).
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

// Unpad removes padding from data.
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

// IsPadded checks if data appears to be padded.
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
