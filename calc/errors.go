package calc

// ErrCode tells what is wrong with an expression; the UI has a text for each
type ErrCode int

const (
	// ErrIncomplete: the expression ends before it is complete, as "2+"
	// while it is being typed
	ErrIncomplete ErrCode = iota
	// ErrSyntax: a token that cannot be there
	ErrSyntax
	// ErrUnknownName: no function, constant or variable of the name
	ErrUnknownName
	ErrDivByZero
	// ErrDomain: the function is not defined there, as √-1 or ln 0
	ErrDomain
	ErrOverflow
	// ErrArgCount: a function is given too few or too many arguments
	ErrArgCount
	// ErrNotInteger: the operation works on integers only (bits, gcd)
	ErrNotInteger
	// ErrReadOnly: a constant or a function cannot be assigned
	ErrReadOnly
	// ErrUnmatched: a closing parenthesis without the opening one
	ErrUnmatched
)

// Error is an error in an expression; Pos is the index of the rune where
// it is, -1 when it is not known
type Error struct {
	Code ErrCode
	Pos  int
	// Name of the function or the variable, for ErrUnknownName, ErrArgCount and ErrReadOnly
	Name string
}

func (e *Error) Error() string {
	names := [...]string{"incomplete expression", "syntax error", "unknown name", "division by zero",
		"out of domain", "overflow", "wrong number of arguments", "integer required",
		"read-only name", "unmatched parenthesis"}
	s := "error"
	if int(e.Code) < len(names) {
		s = names[e.Code]
	}
	if e.Name != "" {
		s += ": " + e.Name
	}
	return s
}

// at sets the position of the error when it has none
func at(err error, pos int) error {
	if e, ok := err.(*Error); ok && e.Pos < 0 {
		e.Pos = pos
	}
	return err
}
