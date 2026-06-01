package shamir

import (
	"encoding/hex"
	"fmt"

	"testing"
)

// Example demonstrates basic Shamir secret sharing usage.
func Example() {
	secret := []byte("my-secret-aes-key-16b")
	threshold := 3
	userXs := []uint32{1001, 2002, 3003, 4004, 5005}

	shares, err := Split(secret, threshold, userXs)
	if err != nil {
		fmt.Printf("Split failed: %v\n", err)
		return
	}

	fmt.Printf("Generated %d shares\n", len(shares))
	for i, share := range shares {
		fmt.Printf("Share %d (X=%d): %x\n", i+1, share.Index, share.Values)
	}

	recovered := Recover(shares)
	unpaddedRecovered := Unpad(recovered)

	fmt.Printf("\nRecovered secret: %s\n", string(unpaddedRecovered))
	fmt.Printf("Original matches recovered: %t\n", string(unpaddedRecovered) == string(secret))
}

// Example_split_recover demonstrates splitting and recovering with threshold.
func Example_split_recover() {
	secret := []byte("test-secret")
	threshold := 3
	userXs := []uint32{1, 2, 3, 4, 5}

	shares, _ := Split(secret, threshold, userXs)

	fmt.Printf("Total shares: %d\n", len(shares))
	fmt.Printf("Threshold: %d\n", threshold)

	for i := threshold; i <= len(shares); i++ {
		subset := shares[:i]
		recovered := Recover(subset)
		fmt.Printf("Using %d shares - recovery successful: %t\n", i, string(recovered) == string(secret))
	}
}

// Example_delta_update demonstrates zero-value perturbation for share updates.
func Example_delta_update() {
	secret := []byte("audit-key-1234")
	threshold := 3
	userXs := []uint32{100, 200, 300}

	shares, _ := Split(secret, threshold, userXs)
	originalRecovered := Recover(shares)
	fmt.Printf("Original secret: %s\n", string(originalRecovered))

	for updateRound := 1; updateRound <= 3; updateRound++ {
		delta, _ := GenerateDelta(threshold, len(userXs))
		shares = ApplyDelta(shares, delta)

		updatedRecovered := Recover(shares)
		fmt.Printf("After update %d - secret unchanged: %t\n", updateRound, string(updatedRecovered) == string(originalRecovered))
	}
}

// Example_user_x demonstrates generating deterministic user X coordinates.
func Example_user_x() {
	userIDs := []string{"alice@example.com", "bob@example.com", "charlie@example.com"}

	fmt.Println("User X coordinates:")
	for _, id := range userIDs {
		x := GenerateUserXFromID(id)
		fmt.Printf("  %s -> X=%d\n", id, x)
	}
}

// Example_padding demonstrates the 4-byte padding functionality.
func Example_padding() {
	testCases := [][]byte{
		[]byte(""),
		[]byte("a"),
		[]byte("ab"),
		[]byte("abc"),
		[]byte("abcd"),
		[]byte("hello"),
	}

	fmt.Println("Padding examples:")
	for _, input := range testCases {
		padded := Pad(input)
		unpadded := Unpad(padded)
		fmt.Printf("  Input: %q (len=%d) -> Padded: %x (len=%d) -> Unpad: %q\n",
			string(input), len(input), padded, len(padded), string(unpadded))
	}
}

// Example_real_world demonstrates a realistic usage scenario.
func Example_real_world() {
	aesKey := make([]byte, 32)
	for i := range aesKey {
		aesKey[i] = byte(i)
	}
	fmt.Printf("Generated AES-256 key: %s\n", hex.EncodeToString(aesKey))

	users := map[string]uint32{
		"admin1": GenerateUserXFromID("admin1@company.com"),
		"admin2": GenerateUserXFromID("admin2@company.com"),
		"admin3": GenerateUserXFromID("admin3@company.com"),
		"admin4": GenerateUserXFromID("admin4@company.com"),
		"admin5": GenerateUserXFromID("admin5@company.com"),
	}

	userXs := make([]uint32, 0, len(users))
	for _, x := range users {
		userXs = append(userXs, x)
	}

	threshold := 3
	shares, err := Split(aesKey, threshold, userXs)
	if err != nil {
		fmt.Printf("Error splitting key: %v\n", err)
		return
	}

	fmt.Printf("\nCreated %d shares with threshold %d\n", len(shares), threshold)

	availableShares := shares[:threshold]
	fmt.Printf("\nRecovering key from %d shares...\n", len(availableShares))

	recoveredKey := Recover(availableShares)
	fmt.Printf("Recovered key: %s\n", hex.EncodeToString(recoveredKey))
	fmt.Printf("Key recovery successful: %t\n", hex.EncodeToString(aesKey) == hex.EncodeToString(recoveredKey))

	fmt.Println("\nApplying zero-value perturbation...")
	delta, _ := GenerateDelta(threshold, len(shares))
	updatedShares := ApplyDelta(shares, delta)

	updatedKey := Recover(updatedShares)
	fmt.Printf("Key after perturbation: %s\n", hex.EncodeToString(updatedKey))
	fmt.Printf("Key unchanged after perturbation: %t\n", hex.EncodeToString(aesKey) == hex.EncodeToString(updatedKey))
}

func TestSomeExample(t *testing.T) {
	Example_delta_update()
}
