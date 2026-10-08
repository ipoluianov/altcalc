package calc

import (
	"strings"
	"unicode"
)

// TokenKind is the kind of a piece of an expression
type TokenKind int

const (
	TokNumber TokenKind = iota
	TokIdent            // a function, a constant, a variable or a word operator (mod, and...)
	TokOp
	TokLParen
	TokRParen
	TokComma // separates the arguments of a function: "," in a call, or ";"
	TokInvalid
)

// Token is a piece of an expression: Start and End are rune indexes
type Token struct {
	Kind       TokenKind
	Start, End int
	// Text is the normalized text: the operator ("−" and "×" become "-" and
	// "*"), the lower-case name, the number without the group separators
	// and with "." for the decimal point
	Text string
}

// wordOps are the operators written as words
var wordOps = map[string]bool{"mod": true, "and": true, "or": true, "xor": true, "not": true, "shl": true, "shr": true}

// IsWordOp tells whether the name is an operator, as "mod"
func IsWordOp(name string) bool {
	return wordOps[strings.ToLower(name)]
}

// opAliases are the other ways to type the operators
var opAliases = map[rune]string{
	'+': "+", '-': "-", '−': "-", '–': "-", '*': "*", '×': "*", '·': "*", '⋅': "*", '∙': "*",
	'/': "/", '÷': "/", ':': "/", '^': "^", '!': "!", '%': "%", '&': "&", '|': "|", '~': "~",
	'=': "=", '√': "√", '²': "²", '³': "³", '°': "°",
}

// isSpace: the spaces, including the ones that group the digits
func isSpace(r rune) bool {
	return unicode.IsSpace(r) || r == '\u00a0' || r == '\u202f' || r == '\u2009'
}

func isIdentStart(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func isIdentPart(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// Tokenize splits the expression into tokens, skipping the spaces
func Tokenize(expr string) []Token {
	rs := []rune(expr)
	var toks []Token
	// parens tells for each open parenthesis whether it is a function call:
	// there "," separates the arguments, elsewhere it is the decimal point
	var parens []bool
	inCall := func() bool { return len(parens) > 0 && parens[len(parens)-1] }
	for i := 0; i < len(rs); {
		r := rs[i]
		switch {
		case isSpace(r):
			i++
		case isDigit(r) || (r == '.' && i+1 < len(rs) && isDigit(rs[i+1])):
			end, text := lexNumber(rs, i, !inCall())
			toks = append(toks, Token{Kind: TokNumber, Start: i, End: end, Text: text})
			i = end
		case isIdentStart(r) || r == 'π' || r == 'τ' || r == 'φ':
			end := i + 1
			for end < len(rs) && isIdentPart(rs[end]) {
				end++
			}
			toks = append(toks, Token{Kind: TokIdent, Start: i, End: end, Text: strings.ToLower(string(rs[i:end]))})
			i = end
		case r == '(':
			n := len(toks)
			parens = append(parens, n > 0 && toks[n-1].Kind == TokIdent && isFunction(toks[n-1].Text))
			toks = append(toks, Token{Kind: TokLParen, Start: i, End: i + 1, Text: "("})
			i++
		case r == ')':
			if len(parens) > 0 {
				parens = parens[:len(parens)-1]
			}
			toks = append(toks, Token{Kind: TokRParen, Start: i, End: i + 1, Text: ")"})
			i++
		case r == ',' || r == ';':
			toks = append(toks, Token{Kind: TokComma, Start: i, End: i + 1, Text: ","})
			i++
		case (r == '<' || r == '>') && i+1 < len(rs) && rs[i+1] == r:
			toks = append(toks, Token{Kind: TokOp, Start: i, End: i + 2, Text: string([]rune{r, r})})
			i += 2
		case r == '*' && i+1 < len(rs) && rs[i+1] == '*':
			toks = append(toks, Token{Kind: TokOp, Start: i, End: i + 2, Text: "^"})
			i += 2
		default:
			if op, ok := opAliases[r]; ok {
				toks = append(toks, Token{Kind: TokOp, Start: i, End: i + 1, Text: op})
			} else {
				toks = append(toks, Token{Kind: TokInvalid, Start: i, End: i + 1, Text: string(r)})
			}
			i++
		}
	}
	return toks
}

// lexNumber reads the number at i: decimal ("1.5", "1,5" when commaDecimal,
// "2e-3", "1 000 000", "1_000"), hex "0x1F", binary "0b101" or octal "0o17".
// Returns where it ends and its normalized text.
func lexNumber(rs []rune, i int, commaDecimal bool) (int, string) {
	if rs[i] == '0' && i+2 < len(rs) {
		base := unicode.ToLower(rs[i+1])
		valid := map[rune]func(r rune) bool{
			'x': func(r rune) bool { return isDigit(r) || strings.ContainsRune("abcdefABCDEF", r) },
			'b': func(r rune) bool { return r == '0' || r == '1' },
			'o': func(r rune) bool { return r >= '0' && r <= '7' },
		}[base]
		if valid != nil && valid(rs[i+2]) {
			end := i + 2
			var sb strings.Builder
			sb.WriteRune('0')
			sb.WriteRune(base)
			for end < len(rs) && (valid(rs[end]) || rs[end] == '_' && end+1 < len(rs) && valid(rs[end+1])) {
				if rs[end] != '_' {
					sb.WriteRune(rs[end])
				}
				end++
			}
			return end, sb.String()
		}
	}

	var sb strings.Builder
	end := i
	digits := func() {
		for end < len(rs) {
			switch {
			case isDigit(rs[end]):
				sb.WriteRune(rs[end])
				end++
			case rs[end] == '_' && end+1 < len(rs) && isDigit(rs[end+1]):
				end++
			default:
				return
			}
		}
	}
	digits()
	// "1 000 000": a space and a group of exactly three digits continue the integer part
	for sb.Len() > 0 && end+3 < len(rs) && isSpace(rs[end]) && rs[end] != '\n' &&
		isDigit(rs[end+1]) && isDigit(rs[end+2]) && isDigit(rs[end+3]) &&
		(end+4 >= len(rs) || !isDigit(rs[end+4])) {
		sb.WriteString(string(rs[end+1 : end+4]))
		end += 4
	}
	if end < len(rs) && (rs[end] == '.' || rs[end] == ',' && commaDecimal && end+1 < len(rs) && isDigit(rs[end+1])) {
		sb.WriteRune('.')
		end++
		digits()
	}
	if end < len(rs) && (rs[end] == 'e' || rs[end] == 'E') {
		j := end + 1
		if j < len(rs) && (rs[j] == '+' || rs[j] == '-' || rs[j] == '−') {
			j++
		}
		if j < len(rs) && isDigit(rs[j]) {
			sb.WriteRune('e')
			if rs[end+1] == '-' || rs[end+1] == '−' {
				sb.WriteRune('-')
			}
			end = j
			digits()
		}
	}
	return end, sb.String()
}
