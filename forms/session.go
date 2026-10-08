package forms

import (
	"strings"
	"unicode"

	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/altcalc/config"
)

// Session is the work of the calculator without the UI: the line being
// typed, what Enter does to it, the history, the variables and the memory
type Session struct {
	Calc *calc.Calc
	Line EditLine

	// Fresh: the line shows the result of Enter. An operator typed goes on
	// with it, a number or a name starts a new expression.
	Fresh bool
	// Prev is the calculation shown above the line after Enter: "12×3"
	Prev string

	History []config.HistoryEntry
	// browse is the entry of the history Up and Down are at; len(History) - none
	browse int
	// draft is the line as it was before Up went into the history
	draft string

	Memory    calc.Value
	HasMemory bool

	// OnChanged is called when the history, the variables or the memory change, to save them
	OnChanged func()
}

func NewSession() *Session {
	return &Session{Calc: calc.New()}
}

// Load takes the history, the variables and the memory that were saved
func (s *Session) Load(h config.History) {
	s.History = h.Entries
	s.browse = len(s.History)
	for name, text := range h.Vars {
		if r, err := s.Calc.Eval(text); err == nil && s.Calc.ValidVarName(name) {
			s.Calc.SetVar(name, r.Value)
		}
	}
	if r, err := s.Calc.Eval(h.Ans); err == nil {
		s.Calc.Commit(calc.Result{Value: r.Value})
	}
	if r, err := s.Calc.Eval(h.Memory); err == nil {
		s.Memory, s.HasMemory = r.Value, true
	}
}

// Save returns what is kept between the starts
func (s *Session) Save() config.History {
	h := config.History{Entries: s.History, Vars: map[string]string{}}
	for _, name := range s.Calc.VarNames() {
		v, _ := s.Calc.Var(name)
		h.Vars[name] = calc.FormatExact(v)
	}
	h.Ans = calc.FormatExact(s.Calc.Ans())
	if s.HasMemory {
		h.Memory = calc.FormatExact(s.Memory)
	}
	return h
}

func (s *Session) changed() {
	if s.OnChanged != nil {
		s.OnChanged()
	}
}

// Preview evaluates the line as it is typed. ok is false when there is
// nothing to show: the line is empty, shows the result, or is not complete yet.
func (s *Session) Preview() (r calc.Result, err error, ok bool) {
	text := s.Line.Text()
	if s.Fresh || strings.TrimSpace(text) == "" {
		return r, nil, false
	}
	r, err = s.Calc.Eval(text)
	if e, isCalcErr := err.(*calc.Error); isCalcErr && e.Code == calc.ErrIncomplete {
		return r, nil, false
	}
	return r, err, true
}

// Value is the number the line means now: its result, or the one Enter gave
func (s *Session) Value() (calc.Value, bool) {
	if strings.TrimSpace(s.Line.Text()) == "" {
		return calc.Value{}, false
	}
	r, err := s.Calc.Eval(s.Line.Text())
	return r.Value, err == nil
}

// Enter evaluates the line: the result takes its place and goes into the
// history. Returns the error of the expression, nil when there is nothing to do.
func (s *Session) Enter() error {
	text := strings.TrimSpace(s.Line.Text())
	if text == "" || s.Fresh {
		return nil
	}
	r, err := s.Calc.Eval(text)
	if err != nil {
		return err
	}
	s.Calc.Commit(r)
	s.Calc.Remember(r.Value)
	s.History = append(s.History, config.HistoryEntry{Expr: text, Result: calc.FormatExact(r.Value)})
	if len(s.History) > config.HistoryLimit {
		s.History = s.History[len(s.History)-config.HistoryLimit:]
	}
	s.browse = len(s.History)
	s.Prev = text
	s.Line.SetText(calc.Format(r.Value, calc.Plain))
	s.Fresh = true
	s.changed()
	return nil
}

// continuesExpression tells whether the typed text goes on with the result
// before it, as "+5" does, instead of starting a new expression
func continuesExpression(text string) bool {
	rs := []rune(strings.TrimLeft(text, " "))
	if len(rs) == 0 {
		return false
	}
	if strings.ContainsRune("+-−*×·/÷^!%&|²³°)<>=,;:", rs[0]) {
		return true
	}
	// " mod 3", " xor 1": a word operator after a space
	if text[0] != ' ' {
		return false
	}
	end := 0
	for end < len(rs) && unicode.IsLetter(rs[end]) {
		end++
	}
	return calc.IsWordOp(string(rs[:end]))
}

// Type puts the text typed or clicked at the cursor. After Enter, a number
// or a name replaces the result, an operator goes on with it.
func (s *Session) Type(text string) {
	if s.Fresh {
		s.Fresh = false
		if continuesExpression(text) {
			s.Line.MoveTo(s.Line.Len(), false)
		} else {
			s.Line.SelectAll()
		}
	}
	s.browse = len(s.History)
	s.Line.Insert(text)
}

// Edited is called on any other change of the line: it is an expression again
func (s *Session) Edited() {
	s.Fresh = false
	s.browse = len(s.History)
}

// Clear empties the line; a second time, also what is shown above it
func (s *Session) Clear() {
	if s.Line.Len() == 0 {
		s.Prev = ""
	}
	s.Line.SetText("")
	s.Fresh = false
	s.browse = len(s.History)
}

// Browse puts the expression of the previous (dir < 0) or the next entry
// of the history in the line; past the last one the line is as it was typed
func (s *Session) Browse(dir int) bool {
	if len(s.History) == 0 {
		return false
	}
	if s.browse == len(s.History) {
		if dir > 0 {
			return false
		}
		s.draft = s.Line.Text()
	}
	pos := max(0, min(len(s.History), s.browse+dir))
	if pos == s.browse {
		return false
	}
	s.browse = pos
	s.Fresh = false
	if pos == len(s.History) {
		s.Line.SetText(s.draft)
	} else {
		s.Line.SetText(s.History[pos].Expr)
	}
	return true
}

// Negate changes the sign of the line: of the number when it is only one,
// else of the whole expression
func (s *Session) Negate() {
	text := s.Line.Text()
	if strings.TrimSpace(text) == "" {
		s.Type("−")
		return
	}
	for _, minus := range []string{"−", "-"} {
		if rest, ok := strings.CutPrefix(text, minus); ok && isPlainNumber(rest) {
			s.Line.SetText(rest)
			s.Edited()
			return
		}
	}
	switch {
	case strings.HasPrefix(text, "−(") && strings.HasSuffix(text, ")") && parensBalanced(text[len("−("):len(text)-1]):
		s.Line.SetText(text[len("−(") : len(text)-1])
	case isPlainNumber(text):
		s.Line.SetText("−" + text)
	default:
		s.Line.SetText("−(" + text + ")")
	}
	s.Edited()
}

// parensBalanced tells whether every parenthesis of the text is closed in
// it: "(1)+(2)" is, so "−((1)+(2))" is the negated "(1)+(2)"
func parensBalanced(text string) bool {
	depth := 0
	for _, r := range text {
		switch r {
		case '(':
			depth++
		case ')':
			if depth--; depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

// isPlainNumber tells whether the text is a single number
func isPlainNumber(text string) bool {
	toks := calc.Tokenize(text)
	return len(toks) == 1 && toks[0].Kind == calc.TokNumber
}

// Wrap puts the expression (the selection, or the whole line) into
// before and after: "1/(" ")"
func (s *Session) Wrap(before, after string) {
	s.Fresh = false
	s.Line.Wrap(before, after)
	s.browse = len(s.History)
}

// Memory operations: MC, MR, M+, M-, MS

func (s *Session) MemoryClear() {
	s.HasMemory = false
	s.Memory = calc.Value{}
	s.changed()
}

func (s *Session) MemoryRecall() {
	if s.HasMemory {
		s.Calc.Remember(s.Memory)
		s.Type(calc.Format(s.Memory, calc.Plain))
	}
}

// MemoryAdd adds the value of the line to the memory, or subtracts it (sign -1)
func (s *Session) MemoryAdd(sign int) error {
	v, ok := s.Value()
	if !ok {
		return nil
	}
	if sign < 0 {
		v = calc.Neg(v)
	}
	if s.HasMemory {
		var err error
		if v, err = calc.Add(s.Memory, v); err != nil {
			return err
		}
	}
	s.Memory, s.HasMemory = v, true
	s.changed()
	return nil
}

func (s *Session) MemoryStore() {
	if v, ok := s.Value(); ok {
		s.Memory, s.HasMemory = v, true
		s.changed()
	}
}

// ClearHistory forgets the calculations
func (s *Session) ClearHistory() {
	s.History = nil
	s.browse = 0
	s.changed()
}

// DeleteEntry removes an entry of the history
func (s *Session) DeleteEntry(i int) {
	if i >= 0 && i < len(s.History) {
		s.History = append(s.History[:i:i], s.History[i+1:]...)
		s.browse = len(s.History)
		s.changed()
	}
}

// UseValue puts a number saved with all its digits (a result of the
// history) into the line as it is shown; it still means the exact value
func (s *Session) UseValue(exact string) {
	r, err := s.Calc.Eval(exact)
	if err != nil {
		s.UseText(exact)
		return
	}
	s.Calc.Remember(r.Value)
	s.UseText(calc.Format(r.Value, calc.Plain))
}

// UseText puts a text from the history or the variables into the line:
// in place of the result after Enter, else at the cursor
func (s *Session) UseText(text string) {
	if s.Fresh {
		s.Fresh = false
		s.Line.SelectAll()
	}
	s.browse = len(s.History)
	s.Line.Insert(text)
}
