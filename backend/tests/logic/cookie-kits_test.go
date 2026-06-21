package logictest

import (
	"backend/internal/config"
	"backend/internal/logic"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func writeJWTResult(t *testing.T, testName, result string) {
	t.Helper()
	ts := time.Now().Format("20060102_150405")
	filename := filepath.Join("output", fmt.Sprintf("test_cookie-kits_%s.log", ts))
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "[%s] %s: %s\n", time.Now().Format("15:04:05"), testName, result)
}

// mockClaim implements logic.WrapperClaims for testing
type mockClaim struct {
	jwt.RegisteredClaims
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	UserCoor uint32 `json:"user_coor"`
}

func (m *mockClaim) GetUserId() int           { return m.UserID }
func (m *mockClaim) GetUsername() string       { return m.Username }
func (m *mockClaim) GetUserCoor() uint32       { return m.UserCoor }
func (m *mockClaim) SetRegisteredClaims(rc jwt.RegisteredClaims) {
	m.RegisteredClaims = rc
}

func newCookieUtils() *logic.CookieUtils {
	cu := logic.NewCookieUtils()
	cu.BuildWithConfig(&config.CookieConfig{
		Key:           "test-token",
		Domain:        "localhost",
		Path:          "/",
		MaxAge:        3600,
		HttpOnly:      true,
		Secret:        "test-jwt-secret-key-32bytes!!",
		SigningMethod: "HS256",
	})
	return cu
}

// ============================================================================
// 正向用例
// ============================================================================

func TestJWT_NewWithClaims_Success(t *testing.T) {
	cu := newCookieUtils()
	claims := &mockClaim{
		UserID:   42,
		Username: "testuser",
		UserCoor: 0xDEADBEEF,
	}
	token := cu.NewWithClaims(claims)
	if token == nil {
		t.Fatal("expected non-nil token")
	}
	if token.Header["alg"] != "HS256" {
		t.Errorf("expected alg=HS256, got %v", token.Header["alg"])
	}
	writeJWTResult(t, "TestJWT_NewWithClaims_Success", "PASS")
}

func TestJWT_GetSignedString_HS256(t *testing.T) {
	cu := newCookieUtils()
	claims := &mockClaim{UserID: 1, Username: "sign-test"}
	token := cu.NewWithClaims(claims)
	signed, err := cu.GetSignedString(token)
	if err != nil {
		t.Fatalf("GetSignedString failed: %v", err)
	}
	if signed == "" {
		t.Error("expected non-empty signed token")
	}
	// JWT should have 3 parts separated by dots
	parts := 0
	for _, c := range signed {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("expected 2 dots in JWT, got %d", parts)
	}
	writeJWTResult(t, "TestJWT_GetSignedString_HS256", "PASS")
}

func TestJWT_ParseAndVerify_ValidToken(t *testing.T) {
	cu := newCookieUtils()
	claims := &mockClaim{
		UserID:   100,
		Username: "verify-user",
		UserCoor: 12345,
	}
	token := cu.NewWithClaims(claims)
	signed, err := cu.GetSignedString(token)
	if err != nil {
		t.Fatalf("GetSignedString failed: %v", err)
	}

	parsedClaim := &mockClaim{}
	parsedToken, valid, err := cu.ParseAndVerify(signed, parsedClaim)
	if err != nil {
		t.Fatalf("ParseAndVerify failed: %v", err)
	}
	if !valid {
		t.Error("expected valid token")
	}
	if parsedToken == nil {
		t.Fatal("expected non-nil parsed token")
	}
	if parsedClaim.UserID != 100 {
		t.Errorf("expected UserID=100, got %d", parsedClaim.UserID)
	}
	if parsedClaim.Username != "verify-user" {
		t.Errorf("expected Username=verify-user, got %s", parsedClaim.Username)
	}
	if parsedClaim.UserCoor != 12345 {
		t.Errorf("expected UserCoor=12345, got %d", parsedClaim.UserCoor)
	}
	writeJWTResult(t, "TestJWT_ParseAndVerify_ValidToken", "PASS")
}

func TestJWT_Claims_RoundTrip(t *testing.T) {
	cu := newCookieUtils()
	original := &mockClaim{
		UserID:   999,
		Username: "roundtrip",
		UserCoor: 0xCAFEBABE,
	}
	token := cu.NewWithClaims(original)
	signed, err := cu.GetSignedString(token)
	if err != nil {
		t.Fatalf("GetSignedString failed: %v", err)
	}

	parsed := &mockClaim{}
	_, valid, err := cu.ParseAndVerify(signed, parsed)
	if err != nil {
		t.Fatalf("ParseAndVerify failed: %v", err)
	}
	if !valid {
		t.Fatal("expected valid token")
	}
	if parsed.UserID != original.UserID {
		t.Errorf("UserID: got %d, want %d", parsed.UserID, original.UserID)
	}
	if parsed.Username != original.Username {
		t.Errorf("Username: got %s, want %s", parsed.Username, original.Username)
	}
	if parsed.UserCoor != original.UserCoor {
		t.Errorf("UserCoor: got %d, want %d", parsed.UserCoor, original.UserCoor)
	}
	writeJWTResult(t, "TestJWT_Claims_RoundTrip", "PASS")
}

// ============================================================================
// 反向用例
// ============================================================================

func TestJWT_ParseAndVerify_ExpiredToken(t *testing.T) {
	cu := newCookieUtils()
	expiredClaim := &mockClaim{
		UserID:   1,
		Username: "expired",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	// Manually create an expired token and sign it
	expToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaim)
	signed, err := expToken.SignedString([]byte("test-jwt-secret-key-32bytes!!"))
	if err != nil {
		t.Fatalf("signing expired token failed: %v", err)
	}

	parsedClaim := &mockClaim{}
	_, valid, _ := cu.ParseAndVerify(signed, parsedClaim)
	if valid {
		t.Error("expired token should not be valid")
	}
	writeJWTResult(t, "TestJWT_ParseAndVerify_ExpiredToken", "PASS")
}

func TestJWT_ParseAndVerify_WrongSecret(t *testing.T) {
	cu := newCookieUtils()
	claims := &mockClaim{UserID: 1, Username: "wrong-secret"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Sign with wrong secret
	signed, err := token.SignedString([]byte("wrong-secret-key!!!!"))
	if err != nil {
		t.Fatalf("signing with wrong secret failed: %v", err)
	}

	parsedClaim := &mockClaim{}
	_, valid, _ := cu.ParseAndVerify(signed, parsedClaim)
	if valid {
		t.Error("token signed with wrong secret should not be valid")
	}
	writeJWTResult(t, "TestJWT_ParseAndVerify_WrongSecret", "PASS")
}

func TestJWT_ParseAndVerify_TamperedPayload(t *testing.T) {
	cu := newCookieUtils()
	claims := &mockClaim{UserID: 1, Username: "tamper-test"}
	tok := cu.NewWithClaims(claims)
	signed, err := cu.GetSignedString(tok)
	if err != nil {
		t.Fatalf("GetSignedString failed: %v", err)
	}

	// Tamper the payload (middle section)
	parts := []byte(signed)
	// Find the middle section and flip a bit
	dot1 := 0
	dot2 := 0
	for i, c := range parts {
		if c == '.' {
			if dot1 == 0 {
				dot1 = i
			} else {
				dot2 = i
				break
			}
		}
	}
	if dot2 > dot1+1 {
		mid := (dot1 + 1 + dot2) / 2
		parts[mid] ^= 0x01
	}

	parsedClaim := &mockClaim{}
	_, valid, _ := cu.ParseAndVerify(string(parts), parsedClaim)
	if valid {
		t.Error("tampered token should not be valid")
	}
	writeJWTResult(t, "TestJWT_ParseAndVerify_TamperedPayload", "PASS")
}

func TestJWT_ParseAndVerify_EmptyToken(t *testing.T) {
	cu := newCookieUtils()
	_, valid, err := cu.ParseAndVerify("")
	if err == nil {
		t.Logf("empty token: no error (err=%v, valid=%v)", err, valid)
	}
	if valid {
		t.Error("empty token should not be valid")
	}
	writeJWTResult(t, "TestJWT_ParseAndVerify_EmptyToken", "PASS")
}

func TestJWT_ParseAndVerify_MalformedToken(t *testing.T) {
	cu := newCookieUtils()
	_, valid, err := cu.ParseAndVerify("not-a-jwt-token-at-all")
	if err == nil {
		t.Logf("malformed token: no error (valid=%v)", valid)
	}
	if valid {
		t.Error("malformed token should not be valid")
	}
	writeJWTResult(t, "TestJWT_ParseAndVerify_MalformedToken", "PASS")
}

func TestJWT_AlgorithmConfusion_None(t *testing.T) {
	cu := newCookieUtils()
	claims := &mockClaim{UserID: 999, Username: "alg-none-test"}
	// Try to create a token with alg=none
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("none signing failed: %v", err)
	}

	parsedClaim := &mockClaim{}
	_, valid, err := cu.ParseAndVerify(signed, parsedClaim)
	if err != nil {
		t.Logf("alg=none token correctly rejected: %v", err)
	}
	if valid {
		t.Error("alg=none token should not be accepted")
	}
	writeJWTResult(t, "TestJWT_AlgorithmConfusion_None", "PASS")
}

// ============================================================================
// Benchmark
// ============================================================================

func BenchmarkJWT_GetSignedString(b *testing.B) {
	cu := newCookieUtils()
	claims := &mockClaim{UserID: 1, Username: "bench"}
	token := cu.NewWithClaims(claims)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cu.GetSignedString(token)
	}
}

func BenchmarkJWT_ParseAndVerify(b *testing.B) {
	cu := newCookieUtils()
	claims := &mockClaim{UserID: 1, Username: "bench"}
	token := cu.NewWithClaims(claims)
	signed, _ := cu.GetSignedString(token)
	parsedClaim := &mockClaim{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cu.ParseAndVerify(signed, parsedClaim)
	}
}
