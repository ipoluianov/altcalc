package calc

import (
	"math"
	"strconv"
	"strings"
)

// Digits is how many significant digits of a float are shown: the ones
// after are the noise of the binary floats, 0.1+0.2 shows 0.3
const Digits = 15

// FormatOptions say how a number is written
type FormatOptions struct {
	// Group separates the thousands with the GroupSep
	Group    bool
	GroupSep string
	// Decimal is the decimal point
	Decimal string
	// Pretty writes the exponent as "E15", "E−9" instead of "e15", "e-9"
	Pretty bool
}

// Plain is the format of the numbers put into an expression: it reads them back
var Plain = FormatOptions{Decimal: "."}

// decimalDigits splits the float into its sign, the significant digits
// rounded to Digits (no trailing zeros) and the exponent of the first one:
// 1234.5 is "12345", 3
func decimalDigits(f float64) (negative bool, digits string, exp int) {
	if f == 0 {
		return false, "0", 0
	}
	s := strconv.FormatFloat(f, 'e', Digits-1, 64)
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}
	mant, e, _ := strings.Cut(s, "e")
	exp, _ = strconv.Atoi(e)
	digits = strings.TrimRight(strings.Replace(mant, ".", "", 1), "0")
	if digits == "" {
		digits = "0"
	}
	return negative, digits, exp
}

// decimalExponent is the exponent of the first significant digit: 2 for 123
func decimalExponent(f float64) int {
	_, _, exp := decimalDigits(f)
	return exp
}

// roundDecimal rounds to n digits after the point (before it when n is
// negative), half away from zero, as the number is shown: 2.675 is 2.68
// though the float is 2.67499999...
func roundDecimal(x float64, n int) float64 {
	if x == 0 || math.IsInf(x, 0) || math.IsNaN(x) {
		return x
	}
	negative, digits, exp := decimalDigits(x)
	keep := exp + 1 + n
	if keep >= len(digits) {
		return x
	}
	sign := 1.0
	if negative {
		sign = -1
	}
	if keep < 0 {
		return 0
	}
	kept := []byte(digits[:keep])
	if digits[keep] >= '5' {
		// Add one to the kept digits, carrying
		i := len(kept) - 1
		for ; i >= 0 && kept[i] == '9'; i-- {
			kept[i] = '0'
		}
		if i >= 0 {
			kept[i]++
		} else {
			kept = append([]byte{'1'}, kept...)
			exp++
		}
	}
	if len(kept) == 0 {
		return 0
	}
	f, _ := strconv.ParseFloat("0."+string(kept)+"e"+strconv.Itoa(exp+1), 64)
	return sign * f
}

// Format writes the number
func Format(v Value, o FormatOptions) string {
	if v.Int {
		s := strconv.FormatInt(v.I, 10)
		neg := strings.HasPrefix(s, "-")
		s = strings.TrimPrefix(s, "-")
		if o.Group {
			s = group(s, o.GroupSep)
		}
		if neg {
			s = "-" + s
		}
		return s
	}
	negative, digits, exp := decimalDigits(v.F)
	var sb strings.Builder
	if negative {
		sb.WriteString("-")
	}
	// Plain for 0.0000001 up to the 15 digits of the integers, else with the exponent
	if exp >= -7 && exp < Digits {
		var intPart, frac string
		if exp >= 0 {
			if len(digits) > exp+1 {
				intPart, frac = digits[:exp+1], digits[exp+1:]
			} else {
				intPart = digits + strings.Repeat("0", exp+1-len(digits))
			}
		} else {
			intPart, frac = "0", strings.Repeat("0", -exp-1)+digits
		}
		if o.Group {
			intPart = group(intPart, o.GroupSep)
		}
		sb.WriteString(intPart)
		if frac != "" {
			sb.WriteString(o.Decimal)
			sb.WriteString(frac)
		}
		return sb.String()
	}
	sb.WriteString(digits[:1])
	if len(digits) > 1 {
		sb.WriteString(o.Decimal)
		sb.WriteString(digits[1:])
	}
	if o.Pretty {
		sb.WriteString("E")
		sb.WriteString(strings.Replace(strconv.Itoa(exp), "-", "−", 1))
	} else {
		sb.WriteString("e")
		sb.WriteString(strconv.Itoa(exp))
	}
	return sb.String()
}

// FormatExact writes the number with all the digits of the float, to be
// saved and read back as it is: 1/3 is 0.3333333333333333
func FormatExact(v Value) string {
	if v.Int {
		return strconv.FormatInt(v.I, 10)
	}
	return strconv.FormatFloat(v.F, 'g', -1, 64)
}

// group separates the digits by three from the right
func group(digits, sep string) string {
	if len(digits) <= 3 {
		return digits
	}
	var sb strings.Builder
	first := len(digits) % 3
	if first > 0 {
		sb.WriteString(digits[:first])
	}
	for i := first; i < len(digits); i += 3 {
		if sb.Len() > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(digits[i : i+3])
	}
	return sb.String()
}

// FormatBase writes the integer in the base 16, 8 or 2; the negative ones
// as the 64-bit two's complement. The digits are grouped by four (by three
// in octal) with sep when it is not empty.
func FormatBase(i int64, base int, sep string) string {
	s := strings.ToUpper(strconv.FormatUint(uint64(i), base))
	if sep == "" {
		return s
	}
	size := 4
	if base == 8 {
		size = 3
	}
	var parts []string
	for len(s) > size {
		parts = append([]string{s[len(s)-size:]}, parts...)
		s = s[:len(s)-size]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, sep)
}

// Fraction finds the simple fraction num/den the float is, with the
// denominator up to maxDen: 0.75 is 3/4. ok is false for the integers and
// the numbers that are no such fraction.
func Fraction(v Value, maxDen int64) (num, den int64, ok bool) {
	if v.Int {
		return 0, 0, false
	}
	x := v.F
	if math.Abs(x) > 1e12 {
		return 0, 0, false
	}
	// The continued fraction of x, its convergents h/k
	h0, h1 := int64(0), int64(1)
	k0, k1 := int64(1), int64(0)
	y := x
	for range 64 {
		a := math.Floor(y)
		ai := int64(a)
		h0, h1 = h1, ai*h1+h0
		k0, k1 = k1, ai*k1+k0
		if k1 > maxDen {
			return 0, 0, false
		}
		if math.Abs(x-float64(h1)/float64(k1)) <= 1e-13*math.Max(1, math.Abs(x)) {
			return h1, k1, k1 > 1
		}
		frac := y - a
		if frac == 0 {
			break
		}
		y = 1 / frac
	}
	return 0, 0, false
}
