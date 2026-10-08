package forms

import (
	"errors"
	"image/color"
	"math"
	"strings"

	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/altcalc/config"
	"github.com/ipoluianov/nui/ui"
)

// Display shows the expression being typed with the cursor and its colors,
// the calculation before it, the result as it is typed and the result in
// the other bases. It takes the keyboard.
type Display struct {
	ui.Widget
	main *MainForm

	cursorOn bool
	// scroll is how far a line wider than the display is moved to the left
	scroll int
	// layout of the line as last painted, for the mouse: where it starts and
	// the x of each rune boundary
	lineX, lineY, lineH int
	runeX               []int
	selecting           bool
	// widths caches the widths of the runes by the font size
	widths map[float64]map[rune]int
}

const (
	displayPadding  = 14
	displayHeight   = 150
	exprFontSize    = 32.0
	exprFontSizeMin = 20.0
	resultFontSize  = 20.0
	smallFontSize   = 12.5
)

func NewDisplay(main *MainForm) *Display {
	var c Display
	c.InitWidget()
	c.main = main
	c.widths = map[float64]map[rune]int{}
	c.SetCanBeFocused(true)
	c.SetMinHeight(displayHeight)
	c.SetMaxHeight(displayHeight)
	c.SetXExpandable(true)
	c.SetMouseCursor(ui.MouseCursorIBeam)
	c.SetOnPaint(c.paint)
	c.SetOnKeyDown(c.keyDown)
	c.SetOnChar(c.char)
	c.SetOnMouseDown(c.mouseDown)
	c.SetOnMouseMove(c.mouseMove)
	c.SetOnMouseUp(func(button ui.MouseButton, x, y int, mods ui.KeyModifiers) bool {
		c.selecting = false
		return true
	})
	c.SetOnMouseDblClick(func(button ui.MouseButton, x, y int, mods ui.KeyModifiers) bool {
		c.session().Edited()
		c.session().Line.SelectWordAt(c.runeAt(x))
		c.changed()
		return true
	})
	c.SetOnFocused(func() { c.cursorOn = true; c.update() })
	c.SetOnFocusLost(func() { c.update() })
	c.AddTimer(530, func() {
		if c.IsFocused() {
			c.cursorOn = !c.cursorOn
			c.update()
		}
	})
	return &c
}

func (c *Display) session() *Session { return c.main.session }

func (c *Display) update() {
	if f := c.Form(); f != nil {
		f.Update()
	}
}

// changed is called after the line changed: the cursor shows, the window follows
func (c *Display) changed() {
	c.cursorOn = true
	c.main.lineChanged()
}

// ---- Keyboard ----

func (c *Display) keyDown(key ui.Key, mods ui.KeyModifiers) bool {
	s := c.session()
	line := &s.Line
	word := mods.Ctrl || mods.Alt
	edit := func(f func()) bool {
		s.Edited()
		f()
		c.changed()
		return true
	}
	switch key {
	case ui.KeyEnter:
		c.main.calculate()
		return true
	case ui.KeyEsc:
		s.Clear()
		c.changed()
		return true
	case ui.KeyBackspace:
		return edit(func() { line.Backspace(word) })
	case ui.KeyDelete:
		return edit(func() { line.Delete(word) })
	case ui.KeyArrowLeft:
		return edit(func() { line.Move(-1, word, mods.Shift) })
	case ui.KeyArrowRight:
		return edit(func() { line.Move(1, word, mods.Shift) })
	case ui.KeyHome:
		return edit(func() { line.MoveTo(0, mods.Shift) })
	case ui.KeyEnd:
		return edit(func() { line.MoveTo(line.Len(), mods.Shift) })
	case ui.KeyArrowUp, ui.KeyArrowDown:
		dir := -1
		if key == ui.KeyArrowDown {
			dir = 1
		}
		if s.Browse(dir) {
			c.changed()
		}
		return true
	case ui.KeyNumpadDot:
		// The numpad dot is "," in some layouts; here it is the decimal point
		c.main.typeText(".")
		return true
	}
	return false
}

// typedAliases are the keys typed that are written otherwise in the line
var typedAliases = map[rune]string{
	'*': "×", '/': "÷", '-': "−",
}

func (c *Display) char(r rune, mods ui.KeyModifiers) bool {
	if r < 32 || r == 127 || mods.Ctrl && !mods.Alt {
		return false
	}
	switch r {
	case '=':
		// "x =" assigns, otherwise = calculates
		if !c.typingAssignment() {
			c.main.calculate()
			return true
		}
	case '\r', '\n':
		return true
	}
	if s, ok := typedAliases[r]; ok {
		c.main.typeText(s)
		return true
	}
	c.main.typeText(string(r))
	return true
}

// typingAssignment tells whether the line is a name for "=" to assign: "rate"
func (c *Display) typingAssignment() bool {
	s := c.session()
	if s.Fresh {
		return false
	}
	text := strings.TrimSpace(s.Line.Text())
	return text != "" && s.Calc.ValidVarName(text)
}

// ---- Mouse ----

// runeAt is the rune boundary nearest to x
func (c *Display) runeAt(x int) int {
	best, dist := 0, math.MaxInt
	for i, rx := range c.runeX {
		if d := abs(c.lineX + rx - x); d < dist {
			best, dist = i, d
		}
	}
	return best
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (c *Display) mouseDown(button ui.MouseButton, x, y int, mods ui.KeyModifiers) bool {
	c.Focus()
	if button != ui.MouseButtonLeft {
		return true
	}
	s := c.session()
	if s.Fresh {
		// The result becomes an expression to edit
		s.Edited()
		c.runeX = c.lineRuneX(exprFontSize)
	}
	s.Line.MoveTo(c.runeAt(x), mods.Shift)
	c.selecting = true
	c.changed()
	return true
}

func (c *Display) mouseMove(x, y int, mods ui.KeyModifiers) bool {
	if !c.selecting {
		return false
	}
	c.session().Line.MoveTo(c.runeAt(x), true)
	c.cursorOn = true
	c.update()
	return true
}

// ---- Painting ----

func (c *Display) runeWidth(size float64, r rune) int {
	m := c.widths[size]
	if m == nil {
		m = map[rune]int{}
		c.widths[size] = m
	}
	w, ok := m[r]
	if !ok {
		w, _, _ = ui.MeasureText(fontNumbers, size, string(r))
		m[r] = w
	}
	return w
}

// lineRuneX returns the x of each rune boundary of the line in the font size
func (c *Display) lineRuneX(size float64) []int {
	text := []rune(c.session().Line.Text())
	xs := make([]int, len(text)+1)
	for i, r := range text {
		xs[i+1] = xs[i] + c.runeWidth(size, r)
	}
	return xs
}

// displayFormat is how the numbers are shown, by the settings
func displayFormat() calc.FormatOptions {
	s := config.GetSettings()
	o := calc.FormatOptions{Group: !s.NoGrouping, GroupSep: "\u00a0", Decimal: ".", Pretty: true}
	if s.DecimalComma {
		o.Decimal = ","
	}
	return o
}

// tokenColors returns the color of each rune of the line by its token; the
// parentheses around the cursor are in the accent
func (c *Display) tokenColors(text string, cursor int) []color.RGBA {
	n := len([]rune(text))
	cols := make([]color.RGBA, n)
	pal := ui.CurrentPalette()
	for i := range cols {
		cols[i] = toRGBA(pal.Text)
	}
	toks := calc.Tokenize(text)
	for _, t := range toks {
		var col color.RGBA
		switch t.Kind {
		case calc.TokNumber:
			continue
		case calc.TokOp:
			col = colorTokOp.get()
		case calc.TokLParen, calc.TokRParen, calc.TokComma:
			col = colorParen.get()
		case calc.TokIdent:
			switch c.session().Calc.KindOf(t.Text) {
			case calc.NameFunction:
				col = colorTokFunc.get()
			case calc.NameConstant, calc.NameVariable:
				col = colorTokName.get()
			case calc.NameOperator:
				col = colorTokOp.get()
			default:
				col = colorTokErr.get()
			}
		default:
			col = colorTokErr.get()
		}
		for i := t.Start; i < t.End && i < n; i++ {
			cols[i] = col
		}
	}
	// The pair of parentheses at the cursor
	if a, b, ok := matchingParens(toks, cursor); ok {
		cols[a] = colorAccent.get()
		if b >= 0 {
			cols[b] = colorAccent.get()
		}
	}
	return cols
}

// matchingParens finds the parenthesis just before or after the cursor and
// its pair (-1 when it has none)
func matchingParens(toks []calc.Token, cursor int) (a, b int, ok bool) {
	at := -1
	for i, t := range toks {
		if (t.Kind == calc.TokLParen || t.Kind == calc.TokRParen) && (t.Start == cursor || t.End == cursor) {
			at = i
			if t.Start == cursor {
				break
			}
		}
	}
	if at < 0 {
		return 0, 0, false
	}
	dir := 1
	if toks[at].Kind == calc.TokRParen {
		dir = -1
	}
	depth := 0
	for i := at; i >= 0 && i < len(toks); i += dir {
		switch toks[i].Kind {
		case calc.TokLParen:
			depth += dir
		case calc.TokRParen:
			depth -= dir
		}
		if depth == 0 {
			return toks[at].Start, toks[i].Start, true
		}
	}
	return toks[at].Start, -1, true
}

func toRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func (c *Display) paint(cnv *ui.Canvas) {
	s := c.session()
	pal := ui.CurrentPalette()
	w, h := c.Width(), c.Height()
	border := toRGBA(pal.Border)
	if c.IsFocused() {
		border = mix(colorDisplay.get(), colorAccent.get(), 0.55)
	}
	strokeRoundRect(cnv, 0, 0, w, h, 10, border, colorDisplay.get())

	inner := w - 2*displayPadding
	muted := colorMuted.get()
	cnv.SetVAlign(ui.VAlignCenter)

	// The calculation before: "12 × 3 ="
	cnv.SetHAlign(ui.HAlignRight)
	cnv.SetFontFamily(fontNumbers)
	cnv.SetFontSize(smallFontSize + 1)
	cnv.SetColor(muted)
	if s.Fresh && s.Prev != "" {
		prev := s.Prev
		if !strings.Contains(prev, "=") {
			prev += " ="
		}
		cnv.DrawText(displayPadding, 8, inner, 22, fitLeft(fontNumbers, smallFontSize+1, prev, inner))
	}

	// The line
	c.lineY, c.lineH = 32, 52
	preview, previewErr, hasPreview := s.Preview()
	var shown calc.Value
	hasValue := false
	if s.Fresh {
		shown, hasValue = s.Value()
	} else if hasPreview && previewErr == nil {
		shown, hasValue = preview.Value, true
	}
	if s.Fresh && hasValue {
		// The result as the settings show the numbers
		text := calc.Format(shown, displayFormat())
		size := fitFontSize(fontNumbers, text, inner)
		cnv.SetFontSize(size)
		cnv.SetColor(pal.Text)
		cnv.SetHAlign(ui.HAlignRight)
		cnv.DrawText(displayPadding, c.lineY, inner, c.lineH, text)
		c.runeX = nil
	} else {
		c.paintLine(cnv, inner, previewErr)
	}

	// The result as it is typed, or the error
	cnv.SetHAlign(ui.HAlignRight)
	resY := c.lineY + c.lineH + 2
	switch {
	case hasPreview && previewErr != nil:
		cnv.SetFontFamily(ui.ThemeFontFamily())
		cnv.SetFontSize(smallFontSize + 1.5)
		cnv.SetColor(colorTokErr.get())
		cnv.DrawText(displayPadding, resY, inner, 30, T().ErrorText(previewErr))
	case hasPreview && !s.Fresh:
		cnv.SetFontFamily(fontNumbers)
		text := "= " + calc.Format(preview.Value, displayFormat())
		cnv.SetFontSize(min(resultFontSize, fitFontSize(fontNumbers, text, inner)))
		cnv.SetColor(muted)
		cnv.DrawText(displayPadding, resY, inner, 30, text)
	}

	// The integer in the other bases, the fraction
	if hasValue {
		if info := valueInfo(shown); info != "" {
			cnv.SetFontFamily(fontNumbers)
			cnv.SetFontSize(smallFontSize)
			cnv.SetColor(muted)
			cnv.DrawText(displayPadding, resY+30, inner, 20, fitLeft(fontNumbers, smallFontSize, info, inner))
		}
	}
}

// paintLine draws the expression with its colors, the selection and the cursor
func (c *Display) paintLine(cnv *ui.Canvas, inner int, evalErr error) {
	s := c.session()
	line := &s.Line
	text := line.Text()
	pal := ui.CurrentPalette()
	if text == "" {
		cnv.SetFontFamily(fontNumbers)
		cnv.SetFontSize(exprFontSize)
		cnv.SetColor(colorMuted.get())
		cnv.SetHAlign(ui.HAlignRight)
		cnv.DrawText(displayPadding, c.lineY, inner, c.lineH, "0")
		c.lineX = displayPadding + inner
		c.runeX = []int{0}
		if c.IsFocused() && c.cursorOn {
			cnv.FillRect(c.lineX-1, c.lineY+10, 2, c.lineH-20, colorAccent.get())
		}
		return
	}

	// The font gets smaller for a long line, then the line scrolls
	size := fitFontSize(fontNumbers, text, inner)
	xs := c.lineRuneX(size)
	total := xs[len(xs)-1]
	cur := line.Cursor()
	if total <= inner {
		c.scroll = 0
		c.lineX = displayPadding + inner - total
	} else {
		cx := xs[cur]
		if cx-c.scroll > inner-8 {
			c.scroll = cx - inner + 8
		}
		if cx-c.scroll < 8 {
			c.scroll = max(0, cx-8)
		}
		c.scroll = min(c.scroll, total-inner+4)
		c.lineX = displayPadding - c.scroll
	}
	c.runeX = xs

	cnv.Save()
	cnv.TranslateAndClip(displayPadding-2, c.lineY, inner+4, c.lineH)
	ox := c.lineX - (displayPadding - 2)

	// The selection
	if from, to := line.Selection(); from != to {
		cnv.FillRect(ox+xs[from], 8, xs[to]-xs[from], c.lineH-16, pal.Selection)
	}

	// The text by the runs of one color
	cnv.SetFontFamily(fontNumbers)
	cnv.SetFontSize(size)
	cnv.SetHAlign(ui.HAlignLeft)
	rs := []rune(text)
	cols := c.tokenColors(text, cur)
	for i := 0; i < len(rs); {
		j := i + 1
		for j < len(rs) && cols[j] == cols[i] {
			j++
		}
		cnv.SetColor(cols[i])
		cnv.DrawText(ox+xs[i], 0, xs[j]-xs[i]+8, c.lineH, string(rs[i:j]))
		i = j
	}

	// Where the error is: a wavy line under the token
	var e *calc.Error
	if errors.As(evalErr, &e) && e.Pos >= 0 && e.Pos <= len(rs) {
		from, to := e.Pos, e.Pos+1
		for _, t := range calc.Tokenize(text) {
			if t.Start == e.Pos {
				to = t.End
			}
		}
		to = min(to, len(rs))
		x1, x2 := ox+xs[min(from, len(rs))], ox+xs[to]
		if x2 <= x1 {
			x1, x2 = x1-4, x1+4
		}
		y := c.lineH - 9
		for x := x1; x < x2; x += 4 {
			cnv.DrawLine(x, y, min(x+2, x2), y+2, 1, colorTokErr.get())
			cnv.DrawLine(min(x+2, x2), y+2, min(x+4, x2), y, 1, colorTokErr.get())
		}
	}

	if c.IsFocused() && c.cursorOn {
		cnv.FillRect(ox+xs[cur]-1, 10, 2, c.lineH-20, colorAccent.get())
	}
	cnv.Restore()
}

// fitFontSize is the font size of the line for the text to fit the width,
// down to exprFontSizeMin
func fitFontSize(family, text string, width int) float64 {
	for size := exprFontSize; size > exprFontSizeMin; size -= 2 {
		if w, _, _ := ui.MeasureText(family, size, text); w <= width {
			return size
		}
	}
	return exprFontSizeMin
}

// fitLeft cuts the beginning of the text to fit the width: "…+ 3 × 4"
func fitLeft(family string, size float64, text string, width int) string {
	if w, _, _ := ui.MeasureText(family, size, text); w <= width {
		return text
	}
	rs := []rune(text)
	for i := 1; i < len(rs); i++ {
		t := "…" + string(rs[i:])
		if w, _, _ := ui.MeasureText(family, size, t); w <= width {
			return t
		}
	}
	return ""
}

// valueInfo is the line under the result: an integer in hex, octal and
// binary, a fraction as num/den
func valueInfo(v calc.Value) string {
	if v.Int {
		i := v.I
		if i > -2 && i < 2 {
			return ""
		}
		parts := []string{"HEX " + calc.FormatBase(i, 16, ""), "OCT " + calc.FormatBase(i, 8, "")}
		if i > 0 && i < 1<<24 {
			parts = append(parts, "BIN "+calc.FormatBase(i, 2, " "))
		}
		return strings.Join(parts, "   ")
	}
	// A fraction of a short decimal tells nothing new: 0.75 is not shown as 3/4
	_, frac, _ := strings.Cut(calc.Format(v, calc.Plain), ".")
	if len(frac) < 4 || strings.Contains(frac, "e") {
		return ""
	}
	if num, den, ok := calc.Fraction(v, 10000); ok {
		return calc.Format(calc.IntValue(num), calc.Plain) + "/" + calc.Format(calc.IntValue(den), calc.Plain)
	}
	return ""
}
