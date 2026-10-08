package calc

import (
	"math"
	"math/rand/v2"
	"slices"
	"sort"
)

// AngleMode is the unit of the angles of the trigonometric functions
type AngleMode int

const (
	Degrees AngleMode = iota
	Radians
	Gradians
)

// toRadians converts an angle of the mode to radians
func (m AngleMode) toRadians(x float64) float64 {
	switch m {
	case Degrees:
		return x * math.Pi / 180
	case Gradians:
		return x * math.Pi / 200
	}
	return x
}

// fromRadians converts an angle in radians to the mode
func (m AngleMode) fromRadians(x float64) float64 {
	switch m {
	case Degrees:
		return x * 180 / math.Pi
	case Gradians:
		return x * 200 / math.Pi
	}
	return x
}

// fullTurn is the full circle in the units of the mode
func (m AngleMode) fullTurn() float64 {
	switch m {
	case Degrees:
		return 360
	case Gradians:
		return 400
	}
	return 2 * math.Pi
}

// function is a function of the expressions: its arguments, at least
// minArgs and at most maxArgs (-1: any number)
type function struct {
	minArgs, maxArgs int
	f                func(c *Calc, args []Value) (Value, error)
}

// unary makes a function of one float argument
func unary(f func(x float64) float64) function {
	return function{1, 1, func(c *Calc, a []Value) (Value, error) { return checked(f(a[0].Float())) }}
}

// trig makes a trigonometric function: the argument is an angle of the mode
func trig(f func(c *Calc, x float64) (float64, error)) function {
	return function{1, 1, func(c *Calc, a []Value) (Value, error) {
		r, err := f(c, a[0].Float())
		if err != nil {
			return Value{}, err
		}
		return checked(r)
	}}
}

// inverseTrig makes an inverse trigonometric function: the result is an angle of the mode
func inverseTrig(f func(x float64) float64) function {
	return function{1, 1, func(c *Calc, a []Value) (Value, error) {
		r := f(a[0].Float())
		if math.IsNaN(r) {
			return Value{}, &Error{Code: ErrDomain, Pos: -1}
		}
		return checked(cleanAngle(c.Angle.fromRadians(r)))
	}}
}

// cleanAngle rounds an angle that is a whole number of degrees but for the
// error of the conversion: asin(0.5) = 30, not 30.000000000000004
func cleanAngle(x float64) float64 {
	if r := math.Round(x); r != x && math.Abs(x-r) < 1e-12*math.Max(1, math.Abs(x)) {
		return r
	}
	return x
}

// sinCos returns the sine and the cosine of the angle of the mode. The
// angles that are multiples of a quarter turn give exact 0 and ±1: sin 180° = 0.
func sinCos(c *Calc, x float64) (sin, cos float64) {
	turn := c.Angle.fullTurn()
	if q := x / (turn / 4); q == math.Trunc(q) && math.Abs(q) < 1<<52 {
		switch ((int64(q) % 4) + 4) % 4 {
		case 0:
			return 0, 1
		case 1:
			return 1, 0
		case 2:
			return 0, -1
		default:
			return -1, 0
		}
	}
	// 30° and the like: sin 30° = 0.5 exactly
	if c.Angle == Degrees {
		if m := math.Mod(x, 360); m == math.Trunc(m) {
			if m < 0 {
				m += 360
			}
			d := int(m)
			s, okS := exactSinDeg[d]
			co, okC := exactSinDeg[(d+90)%360] // cos x = sin(x + 90°)
			if okS || okC {
				r := c.Angle.toRadians(x)
				if !okS {
					s = math.Sin(r)
				}
				if !okC {
					co = math.Cos(r)
				}
				return s, co
			}
		}
	}
	r := c.Angle.toRadians(x)
	sin, cos = math.Sin(r), math.Cos(r)
	// sin π is 1.2e-16 because π is not exact: an angle that is not tiny
	// itself gives 0 for that
	if math.Abs(r) > 1e-6 {
		if math.Abs(sin) < 1e-15 {
			sin = 0
		}
		if math.Abs(cos) < 1e-15 {
			cos = 0
		}
	}
	return sin, cos
}

// exactSinDeg are the sines of the angles that are multiples of 30°, exact
// where a float can be
var exactSinDeg = map[int]float64{
	0: 0, 30: 0.5, 90: 1, 150: 0.5, 180: 0, 210: -0.5, 270: -1, 330: -0.5,
}

var functions map[string]function

func init() {
	functions = map[string]function{
		"sin": trig(func(c *Calc, x float64) (float64, error) { s, _ := sinCos(c, x); return s, nil }),
		"cos": trig(func(c *Calc, x float64) (float64, error) { _, co := sinCos(c, x); return co, nil }),
		"tan": trig(func(c *Calc, x float64) (float64, error) {
			s, co := sinCos(c, x)
			if co == 0 {
				return 0, &Error{Code: ErrDomain, Pos: -1}
			}
			if s == 0 {
				return 0, nil
			}
			return s / co, nil
		}),
		"cot": trig(func(c *Calc, x float64) (float64, error) {
			s, co := sinCos(c, x)
			if s == 0 {
				return 0, &Error{Code: ErrDomain, Pos: -1}
			}
			if co == 0 {
				return 0, nil
			}
			return co / s, nil
		}),
		"sec": trig(func(c *Calc, x float64) (float64, error) {
			_, co := sinCos(c, x)
			if co == 0 {
				return 0, &Error{Code: ErrDomain, Pos: -1}
			}
			return 1 / co, nil
		}),
		"csc": trig(func(c *Calc, x float64) (float64, error) {
			s, _ := sinCos(c, x)
			if s == 0 {
				return 0, &Error{Code: ErrDomain, Pos: -1}
			}
			return 1 / s, nil
		}),
		"asin": inverseTrig(math.Asin),
		"acos": inverseTrig(math.Acos),
		"atan": inverseTrig(math.Atan),
		"acot": inverseTrig(func(x float64) float64 { return math.Pi/2 - math.Atan(x) }),
		"atan2": {2, 2, func(c *Calc, a []Value) (Value, error) {
			return checked(cleanAngle(c.Angle.fromRadians(math.Atan2(a[0].Float(), a[1].Float()))))
		}},
		"sinh":  unary(math.Sinh),
		"cosh":  unary(math.Cosh),
		"tanh":  unary(math.Tanh),
		"asinh": unary(math.Asinh),
		"acosh": unary(math.Acosh),
		"atanh": unary(math.Atanh),

		"sqrt": {1, 1, func(c *Calc, a []Value) (Value, error) { return sqrt(a[0]) }},
		"cbrt": unary(math.Cbrt),
		"root": {2, 2, func(c *Calc, a []Value) (Value, error) {
			n := a[1].Float()
			if n == 0 {
				return Value{}, &Error{Code: ErrDomain, Pos: -1}
			}
			if a[1].Int && a[1].I%2 != 0 && a[0].Float() < 0 {
				return checked(-math.Pow(-a[0].Float(), 1/n))
			}
			return pow(a[0], FloatValue(1/n))
		}},
		"exp": unary(math.Exp),
		"ln":  unary(logOf(math.Log)),
		"log": {1, 2, func(c *Calc, a []Value) (Value, error) {
			if len(a) == 2 {
				b := a[1].Float()
				if b <= 0 || b == 1 {
					return Value{}, &Error{Code: ErrDomain, Pos: -1}
				}
				return checked(cleanLog(logOf(math.Log)(a[0].Float()) / math.Log(b)))
			}
			return checked(logOf(math.Log10)(a[0].Float()))
		}},
		"lg":    unary(logOf(math.Log10)),
		"log10": unary(logOf(math.Log10)),
		"log2":  unary(logOf(math.Log2)),

		"abs": {1, 1, func(c *Calc, a []Value) (Value, error) {
			if a[0].Float() < 0 {
				return neg(a[0]), nil
			}
			return a[0], nil
		}},
		"sign": unary(func(x float64) float64 {
			if x == 0 {
				return 0
			}
			return math.Copysign(1, x)
		}),
		"round": {1, 2, func(c *Calc, a []Value) (Value, error) {
			digits := int64(0)
			if len(a) == 2 {
				var ok bool
				if digits, ok = a[1].asInt(); !ok {
					return Value{}, &Error{Code: ErrNotInteger, Pos: -1}
				}
			}
			return checked(roundDecimal(a[0].Float(), int(max(-300, min(300, digits)))))
		}},
		"floor": unary(math.Floor),
		"ceil":  unary(math.Ceil),
		"trunc": unary(math.Trunc),
		"int":   unary(math.Trunc),
		"frac": {1, 1, func(c *Calc, a []Value) (Value, error) {
			if a[0].Int {
				return IntValue(0), nil
			}
			// The fraction of the number as it is shown: frac(1.1) = 0.1
			x := roundDecimal(a[0].Float(), 15-decimalExponent(a[0].Float()))
			return checked(roundDecimal(x-math.Trunc(x), 15))
		}},
		"min": {1, -1, func(c *Calc, a []Value) (Value, error) {
			return slices.MinFunc(a, func(x, y Value) int { return cmpValues(x, y) }), nil
		}},
		"max": {1, -1, func(c *Calc, a []Value) (Value, error) {
			return slices.MaxFunc(a, func(x, y Value) int { return cmpValues(x, y) }), nil
		}},
		"sum": {1, -1, func(c *Calc, a []Value) (Value, error) { return sumOf(a) }},
		"avg": {1, -1, func(c *Calc, a []Value) (Value, error) {
			s, err := sumOf(a)
			if err != nil {
				return Value{}, err
			}
			return div(s, IntValue(int64(len(a))))
		}},
		"median": {1, -1, func(c *Calc, a []Value) (Value, error) {
			s := slices.Clone(a)
			sort.Slice(s, func(i, j int) bool { return cmpValues(s[i], s[j]) < 0 })
			if len(s)%2 == 1 {
				return s[len(s)/2], nil
			}
			m, err := add(s[len(s)/2-1], s[len(s)/2])
			if err != nil {
				return Value{}, err
			}
			return div(m, IntValue(2))
		}},
		"hypot": {2, 2, func(c *Calc, a []Value) (Value, error) { return checked(math.Hypot(a[0].Float(), a[1].Float())) }},
		"gcd":   {1, -1, func(c *Calc, a []Value) (Value, error) { return intFold(a, gcd) }},
		"lcm": {1, -1, func(c *Calc, a []Value) (Value, error) {
			return intFold(a, func(x, y int64) (int64, bool) {
				if x == 0 || y == 0 {
					return 0, true
				}
				g, _ := gcd(x, y)
				return mulInt(absI(x/g), absI(y))
			})
		}},
		"fact": {1, 1, func(c *Calc, a []Value) (Value, error) { return factorial(a[0]) }},
		"ncr":  {2, 2, func(c *Calc, a []Value) (Value, error) { return combinations(a[0], a[1], false) }},
		"npr":  {2, 2, func(c *Calc, a []Value) (Value, error) { return combinations(a[0], a[1], true) }},
		"rand": {0, 2, func(c *Calc, a []Value) (Value, error) {
			switch len(a) {
			case 0:
				return checked(c.random())
			case 1:
				return randomInt(c, IntValue(1), a[0])
			}
			return randomInt(c, a[0], a[1])
		}},
		// The angle converted to the mode from degrees, from radians
		"deg": {1, 1, func(c *Calc, a []Value) (Value, error) {
			return checked(cleanAngle(c.Angle.fromRadians(Degrees.toRadians(a[0].Float()))))
		}},
		"rad": {1, 1, func(c *Calc, a []Value) (Value, error) {
			return checked(cleanAngle(c.Angle.fromRadians(a[0].Float())))
		}},
	}
}

func isFunction(name string) bool {
	_, ok := functions[name]
	return ok
}

func sqrt(a Value) (Value, error) {
	x := a.Float()
	if x < 0 {
		return Value{}, &Error{Code: ErrDomain, Pos: -1}
	}
	return checked(math.Sqrt(x))
}

// logOf makes the logarithm NaN for 0 and the negative numbers, which are out of its domain
func logOf(f func(float64) float64) func(float64) float64 {
	return func(x float64) float64 {
		if x <= 0 {
			return math.NaN()
		}
		return f(x)
	}
}

// cleanLog rounds the logarithm that is a whole number but for the error of
// the division: log(1000, 10) = 3
func cleanLog(x float64) float64 {
	if r := math.Round(x); math.Abs(x-r) < 1e-14*math.Max(1, math.Abs(x)) {
		return r
	}
	return x
}

func cmpValues(x, y Value) int {
	if x.Int && y.Int {
		switch {
		case x.I < y.I:
			return -1
		case x.I > y.I:
			return 1
		}
		return 0
	}
	switch a, b := x.Float(), y.Float(); {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func sumOf(a []Value) (Value, error) {
	s := IntValue(0)
	for _, v := range a {
		var err error
		if s, err = add(s, v); err != nil {
			return Value{}, err
		}
	}
	return s, nil
}

func absI(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func gcd(x, y int64) (int64, bool) {
	x, y = absI(x), absI(y)
	for y != 0 {
		x, y = y, x%y
	}
	return x, x >= 0
}

// intFold combines the integers with f; the others are an error
func intFold(a []Value, f func(x, y int64) (int64, bool)) (Value, error) {
	acc, ok := a[0].asInt()
	if !ok {
		return Value{}, &Error{Code: ErrNotInteger, Pos: -1}
	}
	acc = absI(acc)
	for _, v := range a[1:] {
		x, ok := v.asInt()
		if !ok {
			return Value{}, &Error{Code: ErrNotInteger, Pos: -1}
		}
		if acc, ok = f(acc, x); !ok {
			return Value{}, &Error{Code: ErrOverflow, Pos: -1}
		}
	}
	return IntValue(acc), nil
}

// combinations is nCr, or nPr when ordered
func combinations(nv, kv Value, ordered bool) (Value, error) {
	n, ok1 := nv.asInt()
	k, ok2 := kv.asInt()
	if !ok1 || !ok2 {
		return Value{}, &Error{Code: ErrNotInteger, Pos: -1}
	}
	if n < 0 || k < 0 {
		return Value{}, &Error{Code: ErrDomain, Pos: -1}
	}
	if k > n {
		return IntValue(0), nil
	}
	if !ordered && k > n-k {
		k = n - k
	}
	// Exact while the integers allow, then the floats
	r, f := int64(1), 1.0
	exact := true
	for i := int64(0); i < k; i++ {
		f *= float64(n - i)
		if !ordered {
			f /= float64(i + 1)
		}
		if exact {
			p, ok := mulInt(r, n-i)
			if !ok {
				exact = false
				continue
			}
			r = p
			if !ordered {
				r /= i + 1 // the product of i+1 consecutive numbers divides by (i+1)!
			}
		}
	}
	if exact {
		return IntValue(r), nil
	}
	return checked(math.Round(f))
}

func randomInt(c *Calc, lo, hi Value) (Value, error) {
	a, ok1 := lo.asInt()
	b, ok2 := hi.asInt()
	if !ok1 || !ok2 {
		return Value{}, &Error{Code: ErrNotInteger, Pos: -1}
	}
	if a > b {
		a, b = b, a
	}
	span := uint64(b - a + 1)
	if span == 0 {
		return IntValue(int64(rand.Uint64())), nil
	}
	return IntValue(a + int64(rand.Uint64N(span))), nil
}

func (c *Calc) random() float64 {
	if c.Rand != nil {
		return c.Rand()
	}
	return rand.Float64()
}
