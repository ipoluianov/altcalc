package forms

import (
	"image/color"
	"strings"

	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/nui/ui"
)

type keyKind int

const (
	keyDigit keyKind = iota
	keyFunc
	keyOp
	keyEq
	keyMem
)

// calcKey is a key of the keypad; with "2nd" on, the one with label2 does act2
type calcKey struct {
	id     string // the id of its hint, see Strings.KeyHints
	label  string
	kind   keyKind
	act    func()
	label2 string
	act2   func()
	// labelFunc, when set, gives the label as it changes (the angle unit)
	labelFunc func() string
	// on tells whether a key that is a switch is on (2nd)
	on func() bool
	// enabled tells whether the key does something now (MR without memory does not)
	enabled func() bool
}

// Keypad is the grid of the keys: the standard ones, and the functions on
// the left of them in the scientific mode. It does not take the focus: the
// keyboard stays with the display.
type Keypad struct {
	ui.Widget
	main *MainForm

	std, sci   [][]*calcKey
	scientific bool
	second     bool // "2nd": the inverse functions

	hover, pressed *calcKey
	// rects of the keys as last painted
	rects map[*calcKey][4]int
}

const (
	keyGap   = 5
	blockGap = 12 // between the functions and the standard keys
	keyRows  = 6
	keyMinW  = 58
	keyMinH  = 44
)

func NewKeypad(main *MainForm) *Keypad {
	var c Keypad
	c.InitWidget()
	c.main = main
	c.rects = map[*calcKey][4]int{}
	c.SetXExpandable(true)
	c.SetYExpandable(true)
	c.build()
	c.SetOnPaint(c.paint)
	c.SetOnMouseDown(func(button ui.MouseButton, x, y int, mods ui.KeyModifiers) bool {
		if button == ui.MouseButtonLeft {
			c.pressed = c.keyAt(x, y)
			c.update()
		}
		return true
	})
	c.SetOnMouseUp(func(button ui.MouseButton, x, y int, mods ui.KeyModifiers) bool {
		k := c.pressed
		c.pressed = nil
		if k != nil && k == c.keyAt(x, y) {
			c.push(k)
		}
		c.update()
		return true
	})
	c.SetOnMouseMove(func(x, y int, mods ui.KeyModifiers) bool {
		if k := c.keyAt(x, y); k != c.hover {
			c.hover = k
			hint := ""
			if k != nil {
				hint = T().KeyHint(k.id)
			}
			c.SetTooltip(hint)
			c.update()
		}
		return true
	})
	c.SetOnMouseLeave(func() {
		c.hover = nil
		c.update()
	})
	c.updateMinSize()
	return &c
}

func (c *Keypad) update() {
	if f := c.Form(); f != nil {
		f.Update()
	}
}

// SetScientific shows or hides the keys of the functions
func (c *Keypad) SetScientific(on bool) {
	c.scientific = on
	c.second = false
	c.updateMinSize()
}

func (c *Keypad) updateMinSize() {
	cols := 4
	gaps := 3 * keyGap
	if c.scientific {
		cols += 5
		gaps += 4*keyGap + blockGap
	}
	c.SetMinWidth(cols*keyMinW + gaps)
	c.SetMinHeight(keyRows*keyMinH + (keyRows-1)*keyGap)
}

// push does what the key does
func (c *Keypad) push(k *calcKey) {
	if k.enabled != nil && !k.enabled() {
		return
	}
	act := k.act
	if c.second && k.act2 != nil {
		act = k.act2
	}
	if act == nil {
		return
	}
	// A function key used with 2nd turns it off, as on the calculators
	if c.second && k.id != "2nd" && k.act2 != nil {
		c.second = false
	}
	act()
	c.main.display.Focus()
	c.update()
}

func (c *Keypad) build() {
	m := c.main
	typ := func(s string) func() { return func() { m.typeText(s) } }
	fn := func(name string) func() { return func() { m.typeFunc(name) } }
	digit := func(d string) *calcKey { return &calcKey{label: d, kind: keyDigit, act: typ(d)} }
	op := func(label, text string) *calcKey { return &calcKey{label: label, kind: keyOp, act: typ(text)} }
	hasMemory := func() bool { return m.session.HasMemory }

	c.std = [][]*calcKey{
		{
			{id: "mc", label: "MC", kind: keyMem, act: m.memoryClear, enabled: hasMemory},
			{id: "mr", label: "MR", kind: keyMem, act: m.memoryRecall, enabled: hasMemory},
			{id: "m+", label: "M+", kind: keyMem, act: func() { m.memoryAdd(1) }},
			{id: "m-", label: "M−", kind: keyMem, act: func() { m.memoryAdd(-1) }},
		},
		{
			{id: "clear", label: "C", kind: keyFunc, act: m.clear},
			{id: "back", label: "⌫", kind: keyFunc, act: m.backspace},
			{id: "pct", label: "%", kind: keyFunc, act: typ("%")},
			op("÷", "÷"),
		},
		{digit("7"), digit("8"), digit("9"), op("×", "×")},
		{digit("4"), digit("5"), digit("6"), op("−", "−")},
		{digit("1"), digit("2"), digit("3"), op("+", "+")},
		{
			{id: "neg", label: "±", kind: keyDigit, act: m.negate},
			digit("0"),
			{label: ".", kind: keyDigit, act: typ(".")},
			{label: "=", kind: keyEq, act: m.calculate},
		},
	}

	f := func(label, name, label2, name2 string) *calcKey {
		k := &calcKey{label: label, kind: keyFunc, act: fn(name)}
		if label2 != "" {
			k.label2, k.act2 = label2, fn(name2)
		}
		return k
	}
	c.sci = [][]*calcKey{
		{
			{id: "2nd", label: "2nd", kind: keyFunc, act: func() { c.second = !c.second }, on: func() bool { return c.second }},
			{id: "angle", kind: keyFunc, act: m.nextAngle, labelFunc: func() string {
				return [...]string{"DEG", "RAD", "GRAD"}[m.session.Calc.Angle]
			}},
			{id: "deg", label: "°", kind: keyFunc, act: typ("°")},
			{label: "(", kind: keyFunc, act: typ("(")},
			{label: ")", kind: keyFunc, act: typ(")")},
		},
		{
			f("sin", "sin", "sin⁻¹", "asin"), f("cos", "cos", "cos⁻¹", "acos"), f("tan", "tan", "tan⁻¹", "atan"),
			{label: "π", kind: keyFunc, act: typ("π")},
			{label: "e", kind: keyFunc, act: typ("e")},
		},
		{
			f("sinh", "sinh", "sinh⁻¹", "asinh"), f("cosh", "cosh", "cosh⁻¹", "acosh"), f("tanh", "tanh", "tanh⁻¹", "atanh"),
			{id: "mod", label: "mod", kind: keyFunc, act: typ(" mod ")},
			{id: "fact", label: "n!", kind: keyFunc, act: typ("!")},
		},
		{
			f("ln", "ln", "log₂", "log2"), f("log", "log", "eˣ", "exp"),
			{label: "√x", kind: keyFunc, act: typ("√"), label2: "∛x", act2: fn("cbrt")},
			{label: "x²", kind: keyFunc, act: typ("²"), label2: "x³", act2: typ("³")},
			{label: "xʸ", kind: keyFunc, act: typ("^"), label2: "ʸ√x", act2: fn("root")},
		},
		{
			{id: "inv", label: "1/x", kind: keyFunc, act: m.reciprocal},
			{id: "abs", label: "|x|", kind: keyFunc, act: fn("abs")},
			{id: "ee", label: "EE", kind: keyFunc, act: m.typeExponent},
			{label: "10ˣ", kind: keyFunc, act: typ("10^"), label2: "2ˣ", act2: typ("2^")},
			{id: "ans", label: "ans", kind: keyFunc, act: typ(calc.AnsName)},
		},
		{
			{id: "comma", label: ",", kind: keyFunc, act: typ(", ")},
			{id: "left", label: "←", kind: keyFunc, act: func() { m.moveCursor(-1) }},
			{id: "right", label: "→", kind: keyFunc, act: func() { m.moveCursor(1) }},
			{id: "rand", label: "rand", kind: keyFunc, act: typ("rand()")},
			f("round", "round", "floor", "floor"),
		},
	}
}

// layout places the keys in the widget
func (c *Keypad) layout() {
	clear(c.rects)
	w, h := c.Width(), c.Height()
	cols := 4
	gaps := 3 * keyGap
	if c.scientific {
		cols += 5
		gaps += 4*keyGap + blockGap
	}
	cellW := float64(w-gaps) / float64(cols)
	cellH := float64(h-(keyRows-1)*keyGap) / keyRows
	place := func(rows [][]*calcKey, x0 float64) {
		for r, row := range rows {
			y := float64(r) * (cellH + keyGap)
			for i, k := range row {
				x := x0 + float64(i)*(cellW+keyGap)
				c.rects[k] = [4]int{int(x + 0.5), int(y + 0.5), int(x+cellW+0.5) - int(x+0.5), int(y+cellH+0.5) - int(y+0.5)}
			}
		}
	}
	x := 0.0
	if c.scientific {
		place(c.sci, 0)
		x = 5*(cellW+keyGap) - keyGap + blockGap
	}
	place(c.std, x)
}

func (c *Keypad) keyAt(x, y int) *calcKey {
	for k, r := range c.rects {
		if x >= r[0] && x < r[0]+r[2] && y >= r[1] && y < r[1]+r[3] {
			return k
		}
	}
	return nil
}

func (c *Keypad) paint(cnv *ui.Canvas) {
	c.layout()
	cnv.SetHAlign(ui.HAlignCenter)
	cnv.SetVAlign(ui.VAlignCenter)
	for k, r := range c.rects {
		c.paintKey(cnv, k, r)
	}
}

func (c *Keypad) paintKey(cnv *ui.Canvas, k *calcKey, r [4]int) {
	var bg, fg color.RGBA
	size := 15.0
	switch k.kind {
	case keyDigit:
		bg, fg, size = colorKeyDigit.get(), colorKeyText.get(), 21
	case keyOp:
		bg, fg, size = colorKeyOp.get(), colorKeyOpText.get(), 22
	case keyEq:
		bg, fg, size = colorAccent.get(), colorAccentText.get(), 24
	case keyMem:
		bg, fg, size = colorKeyMem.get(), colorKeyMemText.get(), 13.5
	default:
		bg, fg = colorKeyFunc.get(), colorKeyFuncText.get()
	}
	label := k.label
	if k.labelFunc != nil {
		label = k.labelFunc()
	}
	if c.second && k.label2 != "" {
		label = k.label2
		fg = colorKeyOpText.get()
	}
	if k.on != nil && k.on() {
		bg, fg = colorKeyOp.get(), colorKeyOpText.get()
	}
	enabled := k.enabled == nil || k.enabled()
	if !enabled {
		fg = mix(fg, bg, 0.6)
	}
	switch {
	case k == c.pressed && enabled:
		bg = mix(bg, colorKeyHoverMix.get(), 0.16)
	case k == c.hover && enabled:
		bg = mix(bg, colorKeyHoverMix.get(), 0.07)
	}
	if !ui.IsDarkTheme && k.kind != keyEq {
		strokeRoundRect(cnv, r[0], r[1], r[2], r[3], 7, colorKeyBorder.get(), bg)
	} else {
		fillRoundRect(cnv, r[0], r[1], r[2], r[3], 7, bg)
	}
	// The labels of the functions get smaller in the small keys
	if k.kind == keyFunc && r[3] < 40 {
		size = 13.5
	}
	// The symbols are drawn with the font of the system that has them
	family := fontKeys
	if strings.ContainsAny(label, symbolRunes) {
		if fontSymbols != "" {
			family = fontSymbols
		} else {
			family = fontNumbers
		}
		if label == "←" || label == "→" || label == "⌫" {
			size = 20
		}
	}
	cnv.SetFontFamily(family)
	cnv.SetFontSize(size)
	cnv.SetColor(fg)
	cnv.DrawText(r[0], r[1], r[2], r[3], label)
}
