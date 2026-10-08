package calc

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

// Calc evaluates the expressions; it keeps the angle mode, the last answer
// and the variables
type Calc struct {
	Angle AngleMode
	// Rand gives the random numbers of rand(), in [0, 1); nil - the default source
	Rand func() float64

	vars map[string]Value
	// exact are the results that were shown rounded, by their text: put back
	// into an expression, "0.333333333333333" is the 1/3 it came from
	exact map[string]Value
}

// exactLimit is how many results Remember keeps
const exactLimit = 256

// AnsName is the variable with the last answer
const AnsName = "ans"

// constants are the names that cannot be assigned
var constants = map[string]float64{
	"pi": math.Pi, "π": math.Pi,
	"e":   math.E,
	"tau": 2 * math.Pi, "τ": 2 * math.Pi,
	"phi": math.Phi, "φ": math.Phi,
}

func New() *Calc {
	return &Calc{vars: map[string]Value{}, exact: map[string]Value{}}
}

// Remember keeps the exact value of a result shown rounded, so that its
// text typed or pasted back into an expression means that value
func (c *Calc) Remember(v Value) {
	if v.Int {
		return
	}
	if len(c.exact) >= exactLimit {
		clear(c.exact)
	}
	if v.F < 0 {
		v = neg(v) // the minus is an operator in an expression
	}
	c.exact[Format(v, Plain)] = v
}

// Result is what an expression gives: the value, and the variable it is
// assigned to for "x = ..."
type Result struct {
	Value  Value
	Assign string
}

// Commit makes the result the last answer and assigns its variable
func (c *Calc) Commit(r Result) {
	c.vars[AnsName] = r.Value
	if r.Assign != "" {
		c.vars[r.Assign] = r.Value
	}
}

// Ans returns the last answer, 0 when there is none yet
func (c *Calc) Ans() Value {
	return c.vars[AnsName]
}

// SetVar sets the variable; the name is lower-cased
func (c *Calc) SetVar(name string, v Value) {
	c.vars[strings.ToLower(name)] = v
}

// DeleteVar removes the variable
func (c *Calc) DeleteVar(name string) {
	delete(c.vars, strings.ToLower(name))
}

// Var returns the variable
func (c *Calc) Var(name string) (Value, bool) {
	v, ok := c.vars[strings.ToLower(name)]
	return v, ok
}

// VarNames returns the names of the variables the user assigned, sorted; ans is not one
func (c *Calc) VarNames() []string {
	var names []string
	for n := range c.vars {
		if n != AnsName {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

// NameKind is what a name in an expression is
type NameKind int

const (
	NameUnknown NameKind = iota
	NameFunction
	NameConstant
	NameVariable // ans included
	NameOperator // mod, and...
)

// KindOf tells what the name is
func (c *Calc) KindOf(name string) NameKind {
	name = strings.ToLower(name)
	switch {
	case wordOps[name]:
		return NameOperator
	case isFunction(name):
		return NameFunction
	case constants[name] != 0:
		return NameConstant
	case name == AnsName:
		return NameVariable
	}
	if _, ok := c.vars[name]; ok {
		return NameVariable
	}
	return NameUnknown
}

// ValidVarName tells whether a variable can have the name
func (c *Calc) ValidVarName(name string) bool {
	toks := Tokenize(name)
	return len(toks) == 1 && toks[0].Kind == TokIdent && toks[0].Start == 0 &&
		toks[0].End == len([]rune(name)) && c.KindOf(name) != NameFunction &&
		c.KindOf(name) != NameConstant && c.KindOf(name) != NameOperator && toks[0].Text != AnsName
}

// Eval evaluates the expression. It changes nothing: Commit makes the
// result the answer. Missing closing parentheses at the end are added.
func (c *Calc) Eval(expr string) (Result, error) {
	p := parser{c: c, toks: Tokenize(expr), end: len([]rune(expr))}
	if len(p.toks) == 0 {
		return Result{}, &Error{Code: ErrIncomplete, Pos: 0}
	}
	var res Result
	// "name = expression" assigns the variable
	if len(p.toks) >= 2 && p.toks[0].Kind == TokIdent && p.toks[1].Kind == TokOp && p.toks[1].Text == "=" {
		name := p.toks[0].Text
		if k := c.KindOf(name); k == NameFunction || k == NameConstant || k == NameOperator || name == AnsName {
			return Result{}, &Error{Code: ErrReadOnly, Pos: p.toks[0].Start, Name: name}
		}
		res.Assign = name
		p.pos = 2
	}
	v, err := p.parseOr()
	if err != nil {
		return Result{}, err
	}
	if t := p.peek(); t != nil {
		if t.Kind == TokRParen {
			return Result{}, &Error{Code: ErrUnmatched, Pos: t.Start}
		}
		return Result{}, &Error{Code: ErrSyntax, Pos: t.Start}
	}
	res.Value = v
	return res, nil
}

type parser struct {
	c    *Calc
	toks []Token
	pos  int
	end  int // the length of the expression, where ErrIncomplete is
	// pct: the last operand parsed was "x%", for "a + b%" = a + a·b/100
	pct bool
}

func (p *parser) peek() *Token {
	if p.pos < len(p.toks) {
		return &p.toks[p.pos]
	}
	return nil
}

// isOp tells whether the next token is one of the operators, the word ones included
func (p *parser) isOp(ops ...string) bool {
	t := p.peek()
	if t == nil || t.Kind != TokOp && !(t.Kind == TokIdent && wordOps[t.Text]) {
		return false
	}
	for _, op := range ops {
		if t.Text == op {
			return true
		}
	}
	return false
}

// startsOperand tells whether the token at i begins an operand: then a
// number or a name after an operand multiplies it ("2π", "3(4+5)")
func (p *parser) startsOperand(i int) bool {
	if i >= len(p.toks) {
		return false
	}
	t := p.toks[i]
	switch t.Kind {
	case TokNumber, TokLParen:
		return true
	case TokIdent:
		return !wordOps[t.Text]
	case TokOp:
		return t.Text == "√"
	}
	return false
}

func (p *parser) incomplete() error {
	return &Error{Code: ErrIncomplete, Pos: p.end}
}

// binary parses the operands of next joined by the operators
func (p *parser) binary(next func() (Value, error), ops []string, apply func(op string, a, b Value) (Value, error)) (Value, error) {
	a, err := next()
	if err != nil {
		return Value{}, err
	}
	for p.isOp(ops...) {
		t := p.toks[p.pos]
		p.pos++
		b, err := next()
		if err != nil {
			return Value{}, err
		}
		if a, err = apply(t.Text, a, b); err != nil {
			return Value{}, at(err, t.Start)
		}
		p.pct = false
	}
	return a, nil
}

// The operators from the lowest precedence: or, xor, and, shifts, sums,
// products, the unary ones, the power, the postfix ones

func (p *parser) parseOr() (Value, error) {
	return p.binary(p.parseXor, []string{"|", "or"}, bitwise)
}

func (p *parser) parseXor() (Value, error) {
	return p.binary(p.parseAnd, []string{"xor"}, bitwise)
}

func (p *parser) parseAnd() (Value, error) {
	return p.binary(p.parseShift, []string{"&", "and"}, bitwise)
}

func (p *parser) parseShift() (Value, error) {
	return p.binary(p.parseSum, []string{"<<", ">>", "shl", "shr"}, bitwise)
}

func (p *parser) parseSum() (Value, error) {
	a, err := p.parseTerm()
	if err != nil {
		return Value{}, err
	}
	for p.isOp("+", "-") {
		t := p.toks[p.pos]
		p.pos++
		b, err := p.parseTerm()
		if err != nil {
			return Value{}, err
		}
		// "200 + 15%" adds 15% of 200
		if p.pct {
			if b, err = mul(a, b); err != nil {
				return Value{}, at(err, t.Start)
			}
		}
		if t.Text == "+" {
			a, err = add(a, b)
		} else {
			a, err = sub(a, b)
		}
		if err != nil {
			return Value{}, at(err, t.Start)
		}
		p.pct = false
	}
	return a, nil
}

func (p *parser) parseTerm() (Value, error) {
	a, err := p.parseUnary()
	if err != nil {
		return Value{}, err
	}
	for {
		var op Token
		switch {
		case p.isOp("*", "/", "%", "mod"):
			op = p.toks[p.pos]
			p.pos++
		case p.startsOperand(p.pos):
			op = Token{Text: "*", Start: p.toks[p.pos].Start}
		default:
			return a, nil
		}
		b, err := p.parseUnary()
		if err != nil {
			return Value{}, err
		}
		switch op.Text {
		case "*":
			a, err = mul(a, b)
		case "/":
			a, err = div(a, b)
		default:
			a, err = mod(a, b)
		}
		if err != nil {
			return Value{}, at(err, op.Start)
		}
		p.pct = false
	}
}

func (p *parser) parseUnary() (Value, error) {
	t := p.peek()
	if t == nil {
		return Value{}, p.incomplete()
	}
	if p.isOp("-", "+", "~", "not", "√") {
		p.pos++
		a, err := p.parseUnary()
		if err != nil {
			return Value{}, err
		}
		p.pct = false
		switch t.Text {
		case "-":
			return neg(a), nil
		case "+":
			return a, nil
		case "√":
			v, err := sqrt(a)
			return v, at(err, t.Start)
		}
		i, ok := a.asInt()
		if !ok {
			return Value{}, &Error{Code: ErrNotInteger, Pos: t.Start}
		}
		return IntValue(^i), nil
	}
	return p.parsePower()
}

func (p *parser) parsePower() (Value, error) {
	a, err := p.parsePostfix()
	if err != nil {
		return Value{}, err
	}
	if p.isOp("^") {
		t := p.toks[p.pos]
		p.pos++
		// Right to left: 2^3^2 = 2^9; and 2^-1
		b, err := p.parseUnary()
		if err != nil {
			return Value{}, err
		}
		p.pct = false
		v, err := pow(a, b)
		return v, at(err, t.Start)
	}
	return a, nil
}

func (p *parser) parsePostfix() (Value, error) {
	a, err := p.parsePrimary()
	if err != nil {
		return Value{}, err
	}
	p.pct = false
	for p.isOp("!", "²", "³", "%", "°") {
		t := p.toks[p.pos]
		// "7 % 3" is the remainder, "15%" the percent
		if t.Text == "%" && p.startsOperand(p.pos+1) {
			return a, nil
		}
		p.pos++
		p.pct = false
		switch t.Text {
		case "!":
			a, err = factorial(a)
		case "²":
			a, err = mul(a, a)
		case "³":
			a, err = pow(a, IntValue(3))
		case "%":
			a, err = div(a, IntValue(100))
			p.pct = true
		case "°":
			// The angle in degrees in the current mode: in radians 180° = π
			if p.c.Angle != Degrees {
				a, err = checked(p.c.Angle.fromRadians(Degrees.toRadians(a.Float())))
			}
		}
		if err != nil {
			return Value{}, at(err, t.Start)
		}
	}
	return a, nil
}

func (p *parser) parsePrimary() (Value, error) {
	t := p.peek()
	if t == nil {
		return Value{}, p.incomplete()
	}
	switch t.Kind {
	case TokNumber:
		p.pos++
		if v, ok := p.c.exact[t.Text]; ok {
			return v, nil
		}
		v, err := parseNumber(t.Text)
		return v, at(err, t.Start)
	case TokLParen:
		p.pos++
		v, err := p.parseOr()
		if err != nil {
			return Value{}, err
		}
		if err := p.closeParen(); err != nil {
			return Value{}, err
		}
		return v, nil
	case TokIdent:
		if wordOps[t.Text] {
			return Value{}, &Error{Code: ErrSyntax, Pos: t.Start}
		}
		p.pos++
		if f, ok := functions[t.Text]; ok {
			return p.call(t, f)
		}
		if v, ok := constants[t.Text]; ok {
			return FloatValue(v), nil
		}
		if v, ok := p.c.vars[t.Text]; ok {
			return v, nil
		}
		if t.Text == AnsName {
			return IntValue(0), nil
		}
		return Value{}, &Error{Code: ErrUnknownName, Pos: t.Start, Name: t.Text}
	case TokRParen:
		return Value{}, &Error{Code: ErrUnmatched, Pos: t.Start}
	}
	return Value{}, &Error{Code: ErrSyntax, Pos: t.Start}
}

// closeParen takes the closing parenthesis; at the end of the expression
// it may be missing
func (p *parser) closeParen() error {
	t := p.peek()
	switch {
	case t == nil:
		return nil
	case t.Kind == TokRParen:
		p.pos++
		return nil
	}
	return &Error{Code: ErrSyntax, Pos: t.Start}
}

// call parses the arguments of the function name and calls it. A function
// of one argument takes it without the parentheses too: "sin 30", "ln 2".
func (p *parser) call(name *Token, f function) (Value, error) {
	var args []Value
	if t := p.peek(); t != nil && t.Kind == TokLParen {
		p.pos++
		if t := p.peek(); t != nil && t.Kind == TokRParen {
			p.pos++
		} else if t != nil {
			for {
				v, err := p.parseOr()
				if err != nil {
					return Value{}, err
				}
				args = append(args, v)
				if t := p.peek(); t == nil || t.Kind != TokComma {
					break
				}
				p.pos++
			}
			if err := p.closeParen(); err != nil {
				return Value{}, err
			}
		}
	} else if f.minArgs >= 1 {
		if p.peek() == nil {
			return Value{}, p.incomplete()
		}
		v, err := p.parseUnary()
		if err != nil {
			return Value{}, err
		}
		args = []Value{v}
	}
	if len(args) < f.minArgs && p.pos >= len(p.toks) && (len(p.toks) == 0 || p.toks[len(p.toks)-1].Kind != TokRParen) {
		return Value{}, p.incomplete()
	}
	if len(args) < f.minArgs || f.maxArgs >= 0 && len(args) > f.maxArgs {
		return Value{}, &Error{Code: ErrArgCount, Pos: name.Start, Name: name.Text}
	}
	p.pct = false
	v, err := f.f(p.c, args)
	return v, at(err, name.Start)
}

// parseNumber reads the normalized text of a number token
func parseNumber(s string) (Value, error) {
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'b' || s[1] == 'o') {
		base := map[byte]int{'x': 16, 'b': 2, 'o': 8}[s[1]]
		u, err := strconv.ParseUint(s[2:], base, 64)
		if err != nil {
			return Value{}, &Error{Code: ErrOverflow, Pos: -1}
		}
		// 0xFFFFFFFFFFFFFFFF is -1, as the 64-bit integers are
		return IntValue(int64(u)), nil
	}
	if !strings.ContainsAny(s, ".e") {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return IntValue(i), nil
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil && !math.IsInf(f, 0) {
		return Value{}, &Error{Code: ErrSyntax, Pos: -1}
	}
	return checked(f)
}

// bitwise applies the operator to the integers
func bitwise(op string, a, b Value) (Value, error) {
	x, ok1 := a.asInt()
	y, ok2 := b.asInt()
	if !ok1 || !ok2 {
		return Value{}, &Error{Code: ErrNotInteger, Pos: -1}
	}
	switch op {
	case "|", "or":
		return IntValue(x | y), nil
	case "&", "and":
		return IntValue(x & y), nil
	case "xor":
		return IntValue(x ^ y), nil
	case "<<", "shl":
		if y < 0 || y > 63 {
			return Value{}, &Error{Code: ErrDomain, Pos: -1}
		}
		return IntValue(x << y), nil
	}
	if y < 0 || y > 63 {
		return Value{}, &Error{Code: ErrDomain, Pos: -1}
	}
	return IntValue(x >> y), nil
}
