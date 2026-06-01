package shamir

import "testing"

func TestFieldOperations(t *testing.T) {
	a := uint32(100)
	b := uint32(200)

	sum := add(a, b)
	if sum != 300 {
		t.Errorf("add(100, 200) = %d, expected 300", sum)
	}

	mulResult := mul(a, b)
	expectedMul := uint32((100 * 200) % prime)
	if mulResult != expectedMul {
		t.Errorf("mul(100, 200) = %d, expected %d", mulResult, expectedMul)
	}

	invResult := inv(a)
	check := mul(a, invResult)
	if check != 1 {
		t.Errorf("mul(100, inv(100)) = %d, expected 1", check)
	}
}

func TestPolynomialEval(t *testing.T) {
	coeffs := []uint32{10, 20, 30}
	x := uint32(2)

	result := evalPolynomial(coeffs, x)
	expected := 10 + 20*2 + 30*2*2
	t.Logf("evalPolynomial([10,20,30], 2) = %d, expected %d", result, expected)
}

func TestLagrangeSimple(t *testing.T) {
	xValues := []uint32{1, 2, 3}
	yValues := []uint32{6, 14, 26}

	result := lagrangeInterpolation(0, xValues, yValues)
	t.Logf("lagrangeInterpolation(0) = %d, expected 2", result)
}

func TestSingleBlockSplitRecover(t *testing.T) {
	secret := []byte("test")
	threshold := 2
	userXs := []uint32{1001, 2002, 3003}
	seed := uint64(123456789)

	shares := Split(secret, threshold, userXs, seed)

	recovered := Recover(shares)
	t.Logf("Secret: %v", secret)
	t.Logf("Recovered: %v", recovered)
	t.Logf("Shares: %+v", shares)
}
