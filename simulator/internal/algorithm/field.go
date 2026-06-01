package algorithm

import "errors"

type Field struct {
	prime int64
}

func NewField(prime int64) (*Field, error) {
	if prime < 2 {
		return nil, errors.New("prime must be at least 2")
	}
	return &Field{prime: prime}, nil
}

func (f *Field) Add(a, b int64) int64 {
	result := a + b
	if result >= f.prime {
		result -= f.prime
	}
	if result < 0 {
		result += f.prime
	}
	return result
}

func (f *Field) Sub(a, b int64) int64 {
	result := a - b
	if result < 0 {
		result += f.prime
	}
	return result
}

func (f *Field) Mul(a, b int64) int64 {
	return (a * b) % f.prime
}

func (f *Field) Div(a, b int64) (int64, error) {
	inv, err := f.Inv(b)
	if err != nil {
		return 0, err
	}
	return f.Mul(a, inv), nil
}

func (f *Field) Inv(a int64) (int64, error) {
	if a == 0 {
		return 0, errors.New("cannot invert zero")
	}

	var t, newT int64 = 0, 1
	var r, newR int64 = f.prime, a % f.prime

	for newR != 0 {
		quot := r / newR
		t, newT = newT, t-quot*newT
		r, newR = newR, r-quot*newR
	}

	if r != 1 {
		return 0, errors.New("no modular inverse exists")
	}

	if t < 0 {
		t += f.prime
	}
	return t, nil
}

func (f *Field) Neg(a int64) int64 {
	if a == 0 {
		return 0
	}
	return f.prime - a
}

func (f *Field) Pow(a, exp int64) int64 {
	if exp == 0 {
		return 1
	}
	if exp == 1 {
		return a % f.prime
	}

	result := int64(1)
	base := a % f.prime

	for exp > 0 {
		if exp%2 == 1 {
			result = f.Mul(result, base)
		}
		base = f.Mul(base, base)
		exp /= 2
	}

	return result
}

func (f *Field) Prime() int64 {
	return f.prime
}
