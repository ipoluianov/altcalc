// Package calc evaluates the expressions typed into AltCalc: the numbers,
// the operators, the functions, the constants and the variables. It has no
// UI and is covered by tests.
package calc

import (
	"math"
	"math/bits"
)

// Value is a number: an exact integer while the operations allow it (the
// sums, products and powers of integers up to ±9.2·10¹⁸), a float otherwise
type Value struct {
	F   float64
	I   int64
	Int bool // the value is the integer I
}

// maxExact is the largest integer a float64 holds exactly: a float result
// up to it that has no fraction is made an integer again
const maxExact = 1 << 53

// IntValue returns the exact integer
func IntValue(i int64) Value {
	return Value{F: float64(i), I: i, Int: true}
}

// FloatValue returns the float; one without a fraction becomes an integer.
// NaN and the infinities are not checked here, see checked.
func FloatValue(f float64) Value {
	if f == math.Trunc(f) && math.Abs(f) <= maxExact {
		return IntValue(int64(f))
	}
	return Value{F: f}
}

// Float returns the value as a float
func (v Value) Float() float64 {
	if v.Int {
		return float64(v.I)
	}
	return v.F
}

// IsZero tells whether the value is 0
func (v Value) IsZero() bool {
	return v.Float() == 0
}

// asInt returns the value as an integer when it is one
func (v Value) asInt() (int64, bool) {
	if v.Int {
		return v.I, true
	}
	f := v.F
	if f != math.Trunc(f) || f < math.MinInt64 || f >= math.MaxInt64 {
		return 0, false
	}
	return int64(f), true
}

// checked makes the float a Value, or the error for NaN (a value outside of
// the domain of a function) and the infinities (an overflow)
func checked(f float64) (Value, error) {
	if math.IsNaN(f) {
		return Value{}, &Error{Code: ErrDomain, Pos: -1}
	}
	if math.IsInf(f, 0) {
		return Value{}, &Error{Code: ErrOverflow, Pos: -1}
	}
	return FloatValue(f), nil
}

// Add, Sub and Neg are the operations of the expressions, for the memory keys
func Add(a, b Value) (Value, error) { return add(a, b) }
func Sub(a, b Value) (Value, error) { return sub(a, b) }
func Neg(a Value) Value             { return neg(a) }

func add(a, b Value) (Value, error) {
	if a.Int && b.Int {
		s := a.I + b.I
		// Overflow when both have the same sign and the sum the other one
		if (a.I >= 0) == (b.I >= 0) && (s >= 0) != (a.I >= 0) {
			return checked(a.Float() + b.Float())
		}
		return IntValue(s), nil
	}
	return checked(a.Float() + b.Float())
}

func neg(a Value) Value {
	if a.Int && a.I != math.MinInt64 {
		return IntValue(-a.I)
	}
	return FloatValue(-a.Float())
}

func sub(a, b Value) (Value, error) {
	if b.Int && b.I == math.MinInt64 {
		return checked(a.Float() - b.Float())
	}
	return add(a, neg(b))
}

func mul(a, b Value) (Value, error) {
	if a.Int && b.Int {
		if p, ok := mulInt(a.I, b.I); ok {
			return IntValue(p), nil
		}
	}
	return checked(a.Float() * b.Float())
}

// mulInt multiplies the integers; ok is false on an overflow
func mulInt(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	negative := (a < 0) != (b < 0)
	ua, ub := absU(a), absU(b)
	hi, lo := bits.Mul64(ua, ub)
	if hi != 0 || lo > math.MaxInt64+1 || (lo == math.MaxInt64+1 && !negative) {
		return 0, false
	}
	if negative {
		return int64(-lo), true
	}
	return int64(lo), true
}

func absU(a int64) uint64 {
	if a < 0 {
		return uint64(-a) // MinInt64 too: its two's complement is 2^63
	}
	return uint64(a)
}

func div(a, b Value) (Value, error) {
	if b.IsZero() {
		return Value{}, &Error{Code: ErrDivByZero, Pos: -1}
	}
	if a.Int && b.Int && b.I != -1 && a.I%b.I == 0 {
		return IntValue(a.I / b.I), nil
	}
	return checked(a.Float() / b.Float())
}

// mod is the remainder with the sign of the divisor, as on the calculators:
// -7 mod 3 = 2
func mod(a, b Value) (Value, error) {
	if b.IsZero() {
		return Value{}, &Error{Code: ErrDivByZero, Pos: -1}
	}
	if a.Int && b.Int {
		if b.I == -1 {
			return IntValue(0), nil
		}
		r := a.I % b.I
		if r != 0 && (r < 0) != (b.I < 0) {
			r += b.I
		}
		return IntValue(r), nil
	}
	x, y := a.Float(), b.Float()
	r := math.Mod(x, y)
	if r != 0 && (r < 0) != (y < 0) {
		r += y
	}
	return checked(r)
}

func pow(a, b Value) (Value, error) {
	if a.Int && b.Int && b.I >= 0 {
		if r, ok := powInt(a.I, b.I); ok {
			return IntValue(r), nil
		}
	}
	x, y := a.Float(), b.Float()
	if x == 0 && y < 0 {
		return Value{}, &Error{Code: ErrDivByZero, Pos: -1}
	}
	// An odd root of a negative number: (-8)^(1/3) = -2
	if x < 0 && y != math.Trunc(y) {
		if inv := 1 / y; math.Abs(inv-math.Round(inv)) < 1e-9 && int64(math.Round(inv))%2 != 0 {
			return checked(-math.Pow(-x, y))
		}
	}
	return checked(math.Pow(x, y))
}

// powInt raises the integer to the power by squaring; ok is false on an overflow
func powInt(base, exp int64) (int64, bool) {
	result := int64(1)
	for exp > 0 {
		if exp&1 == 1 {
			var ok bool
			if result, ok = mulInt(result, base); !ok {
				return 0, false
			}
		}
		exp >>= 1
		if exp > 0 {
			var ok bool
			if base, ok = mulInt(base, base); !ok {
				return 0, false
			}
		}
	}
	return result, true
}

// factorial is n! for the integers; Γ(x+1) for the others
func factorial(a Value) (Value, error) {
	if n, ok := a.asInt(); ok {
		if n < 0 {
			return Value{}, &Error{Code: ErrDomain, Pos: -1}
		}
		r := int64(1)
		for i := int64(2); i <= n; i++ {
			var ok bool
			if r, ok = mulInt(r, i); !ok {
				return checked(math.Gamma(float64(n) + 1))
			}
		}
		return IntValue(r), nil
	}
	return checked(math.Gamma(a.Float() + 1))
}
