package forms

import (
	"testing"

	"github.com/ipoluianov/altcalc/calc"
)

func typeAll(s *Session, keys ...string) {
	for _, k := range keys {
		s.Type(k)
	}
}

func TestSessionEnterAndGoOn(t *testing.T) {
	s := NewSession()
	typeAll(s, "1", "2", "×", "3")
	if r, err, ok := s.Preview(); !ok || err != nil || calc.Format(r.Value, calc.Plain) != "36" {
		t.Fatalf("preview: %v %v %v", r, err, ok)
	}
	if err := s.Enter(); err != nil {
		t.Fatal(err)
	}
	if s.Line.Text() != "36" || !s.Fresh || s.Prev != "12×3" || len(s.History) != 1 {
		t.Fatalf("after Enter: %q fresh=%v prev=%q", s.Line.Text(), s.Fresh, s.Prev)
	}
	// An operator goes on with the result
	typeAll(s, "+", "4")
	if s.Line.Text() != "36+4" {
		t.Errorf("operator after Enter: %q", s.Line.Text())
	}
	s.Enter()
	// A number starts anew
	typeAll(s, "7")
	if s.Line.Text() != "7" {
		t.Errorf("number after Enter: %q", s.Line.Text())
	}
	// " mod" goes on, a name starts anew
	s.Enter()
	typeAll(s, " mod ")
	if s.Line.Text() != "7 mod " {
		t.Errorf("mod after Enter: %q", s.Line.Text())
	}
}

func TestSessionExactResult(t *testing.T) {
	s := NewSession()
	typeAll(s, "1÷3")
	s.Enter()
	typeAll(s, "×3")
	s.Enter()
	if s.Line.Text() != "1" {
		t.Errorf("1/3 then ×3 = %q", s.Line.Text())
	}
}

func TestSessionBrowse(t *testing.T) {
	s := NewSession()
	for _, e := range []string{"1+1", "2+2"} {
		s.Type(e)
		s.Enter()
		s.Clear()
	}
	s.Type("draft")
	s.Browse(-1)
	if s.Line.Text() != "2+2" {
		t.Errorf("up: %q", s.Line.Text())
	}
	s.Browse(-1)
	s.Browse(-1)
	if s.Line.Text() != "1+1" {
		t.Errorf("up up up: %q", s.Line.Text())
	}
	s.Browse(1)
	s.Browse(1)
	if s.Line.Text() != "draft" {
		t.Errorf("back to the draft: %q", s.Line.Text())
	}
}

func TestSessionNegate(t *testing.T) {
	s := NewSession()
	for _, tt := range []struct{ in, want string }{
		{"5", "−5"},
		{"−5", "5"},
		{"-5", "5"},
		{"2+3", "−(2+3)"},
		{"−(2+3)", "2+3"},
		{"−(1)+(2)", "−(−(1)+(2))"},
	} {
		s.Line.SetText(tt.in)
		s.Negate()
		if s.Line.Text() != tt.want {
			t.Errorf("negate %q = %q, want %q", tt.in, s.Line.Text(), tt.want)
		}
	}
}

func TestSessionMemoryAndSave(t *testing.T) {
	s := NewSession()
	s.Type("2^60")
	s.MemoryAdd(1)
	s.MemoryAdd(1)
	s.MemoryAdd(-1)
	if !s.HasMemory || calc.Format(s.Memory, calc.Plain) != "1152921504606846976" {
		t.Errorf("memory = %v", s.Memory)
	}
	s.Clear()
	s.Type("x = 7")
	s.Enter()
	h := s.Save()

	s2 := NewSession()
	s2.Load(h)
	if v, ok := s2.Calc.Var("x"); !ok || calc.Format(v, calc.Plain) != "7" {
		t.Errorf("variable not loaded: %v", h.Vars)
	}
	if calc.Format(s2.Calc.Ans(), calc.Plain) != "7" || !s2.HasMemory || len(s2.History) != 1 {
		t.Errorf("not loaded: %+v", h)
	}
}

func TestNormalizePasted(t *testing.T) {
	for in, want := range map[string]string{
		"1,234.56":    "1234.56",
		"1.234,5":     "1234.5",
		" 2 + 3 =\n":  "2 + 3",
		"12\r\n× 3":   "12 × 3",
		"1 234 567,8": "1234567.8",
	} {
		if got := normalizePasted(in); got != want {
			t.Errorf("normalizePasted(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEditLine(t *testing.T) {
	var e EditLine
	e.Insert("sin(30)")
	e.Move(-1, false, false)
	e.Backspace(false)
	if e.Text() != "sin(3)" || e.Cursor() != 5 {
		t.Errorf("backspace: %q %d", e.Text(), e.Cursor())
	}
	e.Undo()
	if e.Text() != "sin(30)" {
		t.Errorf("undo: %q", e.Text())
	}
	e.Redo()
	e.SelectAll()
	e.Wrap("1/(", ")")
	if e.Text() != "1/(sin(3))" {
		t.Errorf("wrap: %q", e.Text())
	}
	e.MoveTo(e.Len(), false)
	e.Backspace(true)
	if e.Text() != "1/(sin(3)" {
		t.Errorf("word backspace: %q", e.Text())
	}
	e.SelectWordAt(1)
	if e.SelectedText() != "1" {
		t.Errorf("select word: %q", e.SelectedText())
	}
}
