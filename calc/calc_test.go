package calc

import (
	"errors"
	"testing"
)

func eval(t *testing.T, c *Calc, expr string) string {
	t.Helper()
	r, err := c.Eval(expr)
	if err != nil {
		return "error: " + err.Error()
	}
	return Format(r.Value, Plain)
}

func TestEval(t *testing.T) {
	tests := []struct{ expr, want string }{
		// Arithmetic and precedence
		{"2+3*4", "14"},
		{"(2+3)*4", "20"},
		{"10/4", "2.5"},
		{"10/5", "2"},
		{"2^10", "1024"},
		{"2^3^2", "512"},
		{"-2^2", "-4"},
		{"2^-1", "0.5"},
		{"(-2)^2", "4"},
		{"7 mod 3", "1"},
		{"-7 mod 3", "2"},
		{"7 % 3", "1"},
		{"5.5 mod 2", "1.5"},
		{"0.1+0.2", "0.3"},
		{"1/3*3", "1"},
		{"2 × 3 ÷ 4 − 1", "0.5"},
		{"2**8", "256"},
		{"--3", "3"},
		{"3!", "6"},
		{"5!/3!", "20"},
		{"20!", "2432902008176640000"},
		{"3²", "9"},
		{"2³", "8"},
		{"√16", "4"},
		{"√16+9", "13"},
		{"2√9", "6"},
		// Percent
		{"50%", "0.5"},
		{"200+15%", "230"},
		{"200-10%", "180"},
		{"200*15%", "30"},
		{"200/50%", "400"},
		// Implicit multiplication and the missing parentheses
		{"2(3+4)", "14"},
		{"(1+2)(3+4)", "21"},
		{"2pi/pi", "2"},
		{"(2+3", "5"},
		{"sqrt(sqrt(16", "2"},
		// Numbers
		{"1e3", "1000"},
		{"1.5e-3", "0.0015"},
		{".5+.5", "1"},
		{"1 000 000 + 1", "1000001"},
		{"1_000*2", "2000"},
		{"1,5*2", "3"},
		{"0xFF", "255"},
		{"0b1010", "10"},
		{"0o17", "15"},
		{"0xFFFFFFFFFFFFFFFF", "-1"},
		{"2^62", "4611686018427387904"},
		{"2^63", "9.22337203685478e18"},
		{"1e300*1e300", "error: overflow"},
		{"1/1e-320", "error: overflow"},
		// Functions
		{"sin(30)", "0.5"},
		{"sin 30", "0.5"},
		{"sin(180)", "0"},
		{"cos(90)", "0"},
		{"cos(60)", "0.5"},
		{"tan(45)", "1"},
		{"tan(90)", "error: out of domain"},
		{"asin(1)", "90"},
		{"atan2(1, 1)", "45"},
		{"sin(90°)", "1"},
		{"ln(e)", "1"},
		{"log(1000)", "3"},
		{"log(8, 2)", "3"},
		{"log2(1024)", "10"},
		{"ln(0)", "error: out of domain"},
		{"sqrt(-1)", "error: out of domain"},
		{"cbrt(-27)", "-3"},
		{"root(-32, 5)", "-2"},
		{"(-8)^(1/3)", "-2"},
		{"abs(-5)", "5"},
		{"round(2.675, 2)", "2.68"},
		{"round(-2.5)", "-3"},
		{"round(1234, -2)", "1200"},
		{"floor(-2.5)", "-3"},
		{"ceil(2.1)", "3"},
		{"frac(1.1)", "0.1"},
		{"min(3, 1, 2)", "1"},
		{"max(3; 1; 2)", "3"},
		{"max(1,5, 2)", "5"},
		{"sum(1, 2, 3)", "6"},
		{"avg(1, 2)", "1.5"},
		{"median(5, 1, 3, 2)", "2.5"},
		{"gcd(12, 18)", "6"},
		{"lcm(4, 6)", "12"},
		{"ncr(5, 2)", "10"},
		{"npr(5, 2)", "20"},
		{"ncr(60, 30)", "118264581564861424"},
		{"hypot(3, 4)", "5"},
		{"fact(5)", "120"},
		// Bitwise
		{"6 & 3", "2"},
		{"6 | 3", "7"},
		{"6 xor 3", "5"},
		{"1 << 10", "1024"},
		{"1024 shr 2", "256"},
		{"~0", "-1"},
		{"1.5 & 1", "error: integer required"},
		// Errors
		{"1/0", "error: division by zero"},
		{"2+", "error: incomplete expression"},
		{"sin", "error: incomplete expression"},
		{"max(", "error: incomplete expression"},
		{"foo+1", "error: unknown name: foo"},
		{"2)", "error: unmatched parenthesis"},
		{"2 $ 3", "error: syntax error"},
		{"atan2(1)", "error: wrong number of arguments: atan2"},
		{"pi = 3", "error: read-only name: pi"},
	}
	c := New()
	for _, tt := range tests {
		if got := eval(t, c, tt.expr); got != tt.want {
			t.Errorf("%q = %s, want %s", tt.expr, got, tt.want)
		}
	}
}

func TestRadians(t *testing.T) {
	c := New()
	c.Angle = Radians
	for expr, want := range map[string]string{
		"sin(pi)":    "0",
		"cos(pi)":    "-1",
		"sin(pi/2)":  "1",
		"180°":       "3.14159265358979",
		"asin(1)":    "1.5707963267949",
		"sin(1e-20)": "1e-20",
	} {
		if got := eval(t, c, expr); got != want {
			t.Errorf("%q = %s, want %s", expr, got, want)
		}
	}
}

func TestVariables(t *testing.T) {
	c := New()
	if got := eval(t, c, "ans"); got != "0" {
		t.Errorf("ans at the start = %s", got)
	}
	r, err := c.Eval("x = 2 + 3")
	if err != nil || r.Assign != "x" {
		t.Fatalf("assignment: %v %v", r, err)
	}
	// Eval changes nothing until the commit
	if _, err := c.Eval("x"); err == nil {
		t.Error("x is set before the commit")
	}
	c.Commit(r)
	for expr, want := range map[string]string{"x*2": "10", "ans+1": "6", "X": "5", "2x": "10"} {
		if got := eval(t, c, expr); got != want {
			t.Errorf("%q = %s, want %s", expr, got, want)
		}
	}
	if !c.ValidVarName("rate") || c.ValidVarName("sin") || c.ValidVarName("2x") || c.ValidVarName("ans") {
		t.Error("ValidVarName")
	}
	if got := c.VarNames(); len(got) != 1 || got[0] != "x" {
		t.Errorf("VarNames = %v", got)
	}
}

func TestErrorPosition(t *testing.T) {
	c := New()
	_, err := c.Eval("1 + foo")
	var e *Error
	if !errors.As(err, &e) || e.Pos != 4 {
		t.Errorf("position of foo: %v", err)
	}
	_, err = c.Eval("2 + 3/0")
	if !errors.As(err, &e) || e.Code != ErrDivByZero || e.Pos != 5 {
		t.Errorf("position of /0: %+v", e)
	}
}

func TestFormat(t *testing.T) {
	pretty := FormatOptions{Group: true, GroupSep: " ", Decimal: ",", Pretty: true}
	tests := []struct {
		v     Value
		plain string
		nice  string
	}{
		{IntValue(1234567), "1234567", "1 234 567"},
		{IntValue(-1234), "-1234", "-1 234"},
		{FloatValue(1234.5), "1234.5", "1 234,5"},
		{FloatValue(0.000123), "0.000123", "0,000123"},
		{FloatValue(1e-9), "1e-9", "1E−9"},
		{FloatValue(1.5e20), "1.5e20", "1,5E20"},
		{FloatValue(-2.5e-12), "-2.5e-12", "-2,5E−12"},
		{FloatValue(1e15), "1000000000000000", "1 000 000 000 000 000"},
	}
	for _, tt := range tests {
		if got := Format(tt.v, Plain); got != tt.plain {
			t.Errorf("Format(%v) = %s, want %s", tt.v, got, tt.plain)
		}
		if got := Format(tt.v, pretty); got != tt.nice {
			t.Errorf("Format(%v, pretty) = %s, want %s", tt.v, got, tt.nice)
		}
	}
	// What is formatted reads back
	c := New()
	for _, expr := range []string{"1/3", "2^70", "1e-9/7", "-pi"} {
		r, _ := c.Eval(expr)
		s := Format(r.Value, Plain)
		r2, err := c.Eval(s)
		if err != nil || Format(r2.Value, Plain) != s {
			t.Errorf("%s: %s does not read back: %v", expr, s, err)
		}
	}
}

func TestFormatBase(t *testing.T) {
	if got := FormatBase(255, 16, ""); got != "FF" {
		t.Error(got)
	}
	if got := FormatBase(0x12345, 16, " "); got != "1 2345" {
		t.Error(got)
	}
	if got := FormatBase(10, 2, " "); got != "1010" {
		t.Error(got)
	}
	if got := FormatBase(-1, 16, ""); got != "FFFFFFFFFFFFFFFF" {
		t.Error(got)
	}
}

func TestFraction(t *testing.T) {
	for _, tt := range []struct {
		x        float64
		num, den int64
		ok       bool
	}{
		{0.75, 3, 4, true},
		{-0.5, -1, 2, true},
		{1.0 / 3, 1, 3, true},
		{2.0 / 7 * 3, 6, 7, true},
		{3.14159265358979, 0, 0, false},
	} {
		n, d, ok := Fraction(FloatValue(tt.x), 10000)
		if ok != tt.ok || ok && (n != tt.num || d != tt.den) {
			t.Errorf("Fraction(%v) = %d/%d %v", tt.x, n, d, ok)
		}
	}
}

func TestTokenize(t *testing.T) {
	toks := Tokenize("sin(1,5) + 2,5×x")
	var kinds []TokenKind
	for _, tk := range toks {
		kinds = append(kinds, tk.Kind)
	}
	want := []TokenKind{TokIdent, TokLParen, TokNumber, TokComma, TokNumber, TokRParen, TokOp, TokNumber, TokOp, TokIdent}
	if len(kinds) != len(want) {
		t.Fatalf("kinds %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("kinds %v, want %v", kinds, want)
		}
	}
	if toks[7].Text != "2.5" || toks[8].Text != "*" {
		t.Errorf("normalized: %q %q", toks[7].Text, toks[8].Text)
	}
}

func TestRemember(t *testing.T) {
	c := New()
	r, _ := c.Eval("1/3")
	text := Format(r.Value, Plain)
	if got := eval(t, c, text+"*3"); got != "0.999999999999999" {
		t.Errorf("not remembered: %s", got)
	}
	c.Remember(r.Value)
	if got := eval(t, c, text+"*3"); got != "1" {
		t.Errorf("remembered: %s", got)
	}
	r, _ = c.Eval("-2/3")
	c.Remember(r.Value)
	if got := eval(t, c, Format(r.Value, Plain)+"*3"); got != "-2" {
		t.Errorf("remembered negative: %s", got)
	}
}

func TestFormatExact(t *testing.T) {
	c := New()
	for _, expr := range []string{"1/3", "2^70", "1e-300/3", "-pi", "123"} {
		r, _ := c.Eval(expr)
		r2, err := c.Eval(FormatExact(r.Value))
		if err != nil || r2.Value != r.Value {
			t.Errorf("%s: %s reads back as %v: %v", expr, FormatExact(r.Value), r2.Value, err)
		}
	}
}
