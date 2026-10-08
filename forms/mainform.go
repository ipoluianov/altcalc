package forms

import (
	"regexp"
	"strings"

	"github.com/ipoluianov/altcalc/app"
	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/altcalc/config"
	"github.com/ipoluianov/nui/ui"
)

// MainForm is the window of AltCalc: the display, the keypad, the history
// on the right and the status bar
type MainForm struct {
	ui.Widget

	session *Session
	display *Display
	keypad  *Keypad
	history *HistoryPanel
	divider *ui.Widget
	status  *StatusBar
}

const (
	defaultWindowWidth  = 640
	defaultWindowHeight = 600
)

func NewMainForm() *MainForm {
	var c MainForm
	c.InitWidget()
	settings := config.GetSettings()
	c.session = NewSession()
	c.session.Load(config.LoadHistory())
	c.session.Calc.Angle = calc.AngleMode(settings.Angle)
	c.session.OnChanged = c.saveHistory

	c.SetPanelPadding(0)
	c.SetCellPadding(0)

	body := ui.NewPanel()
	body.SetPanelPadding(0)
	body.SetCellPadding(0)
	body.SetXExpandable(true)
	body.SetYExpandable(true)

	left := ui.NewPanel()
	left.SetPanelPadding(10)
	left.SetCellPadding(5)
	left.SetXExpandable(true)
	left.SetYExpandable(true)
	c.display = NewDisplay(&c)
	left.AddWidget(0, 0, c.display)
	c.keypad = NewKeypad(&c)
	c.keypad.SetScientific(settings.Scientific)
	left.AddWidget(1, 0, c.keypad)
	body.AddWidget(0, 0, left)

	c.divider = &ui.Widget{}
	c.divider.InitWidget()
	c.divider.SetMinWidth(1)
	c.divider.SetMaxWidth(1)
	c.divider.SetYExpandable(true)
	c.divider.SetOnPaint(func(cnv *ui.Canvas) {
		cnv.FillRect(0, 0, 1, c.divider.Height(), ui.CurrentPalette().Divider)
	})
	body.AddWidget(0, 1, c.divider)
	c.history = NewHistoryPanel(&c)
	body.AddWidget(0, 2, c.history)
	c.divider.SetVisible(!settings.HideHistory)
	c.history.SetVisible(!settings.HideHistory)
	c.AddWidget(0, 0, body)

	c.status = NewStatusBar(&c)
	c.AddWidget(1, 0, c.status)

	c.history.Refresh()
	c.status.Refresh()
	return &c
}

// lineChanged updates what follows the line: the history, the status bar
func (c *MainForm) lineChanged() {
	c.history.Refresh()
	c.status.Refresh()
	if f := c.Form(); f != nil {
		f.Update()
	}
}

func (c *MainForm) saveHistory() {
	if err := config.SaveHistory(c.session.Save()); err != nil {
		ui.ShowToast(c, T().Error+": "+err.Error(), ui.ToastInfo)
	}
}

// ---- What the keys do ----

// typeText types the text at the cursor
func (c *MainForm) typeText(text string) {
	c.session.Type(text)
	c.display.changed()
}

// typeFunc types the function: "sin(" for the argument to follow, or
// around the selection, or around the result after Enter: sin(30)
func (c *MainForm) typeFunc(name string) {
	s := c.session
	if s.Fresh || s.Line.HasSelection() {
		if s.Fresh {
			s.Line.SelectAll()
		}
		s.Wrap(name+"(", ")")
	} else {
		s.Type(name + "(")
	}
	c.display.changed()
}

// typeExponent types the "e" of 1.5e3; after Enter it goes on with the result
func (c *MainForm) typeExponent() {
	s := c.session
	if s.Fresh {
		s.Edited()
		s.Line.MoveTo(s.Line.Len(), false)
	}
	s.Type("e")
	c.display.changed()
}

func (c *MainForm) calculate() {
	if err := c.session.Enter(); err == nil {
		c.display.changed()
		return
	}
	// The error is shown under the line; the cursor goes to it
	if _, err, ok := c.session.Preview(); ok && err != nil {
		if e, isCalc := err.(*calc.Error); isCalc && e.Pos >= 0 {
			c.session.Line.MoveTo(e.Pos, false)
		}
	}
	c.display.changed()
}

func (c *MainForm) clear() {
	c.session.Clear()
	c.display.changed()
}

func (c *MainForm) backspace() {
	c.session.Edited()
	c.session.Line.Backspace(false)
	c.display.changed()
}

func (c *MainForm) negate() {
	c.session.Negate()
	c.display.changed()
}

// reciprocal makes the expression 1/(expression)
func (c *MainForm) reciprocal() {
	s := c.session
	switch {
	case strings.TrimSpace(s.Line.Text()) == "":
		s.Type("1/")
	case isPlainNumber(s.Line.Text()) && !s.Line.HasSelection():
		s.Line.SetText("1/" + s.Line.Text())
		s.Edited()
	default:
		s.Wrap("1/(", ")")
	}
	c.display.changed()
}

func (c *MainForm) moveCursor(dir int) {
	c.session.Edited()
	c.session.Line.Move(dir, false, false)
	c.display.changed()
}

func (c *MainForm) memoryClear() {
	c.session.MemoryClear()
	c.lineChanged()
}

func (c *MainForm) memoryRecall() {
	c.session.MemoryRecall()
	c.display.changed()
}

func (c *MainForm) memoryAdd(sign int) {
	if err := c.session.MemoryAdd(sign); err != nil {
		ui.ShowToast(c, T().ErrorText(err), ui.ToastInfo)
	}
	c.lineChanged()
}

func (c *MainForm) memoryStore() {
	c.session.MemoryStore()
	c.lineChanged()
}

// nextAngle switches to the next unit of the angles
func (c *MainForm) nextAngle() {
	c.setAngle((c.session.Calc.Angle + 1) % 3)
}

func (c *MainForm) setAngle(m calc.AngleMode) {
	c.session.Calc.Angle = m
	config.UpdateSettings(func(s *config.Settings) { s.Angle = int(m) })
	c.lineChanged()
}

func (c *MainForm) clearHistory() {
	if len(c.session.History) == 0 {
		return
	}
	ui.ShowQuestionMessageBoxOKCancel(c, T().CmdClearHistory, T().ClearHistoryAsk, func() {
		c.session.ClearHistory()
		c.lineChanged()
	}, nil)
}

// useText puts a name from the variables into the line
func (c *MainForm) useText(text string) {
	c.session.UseText(text)
	c.display.changed()
}

// useValue puts a result of the history into the line
func (c *MainForm) useValue(exact string) {
	c.session.UseValue(exact)
	c.display.changed()
}

// ---- Clipboard ----

// copyFormat is how the numbers are copied: the decimal point of the
// settings, no group separators, so that other programs read them
func copyFormat() calc.FormatOptions {
	o := calc.Plain
	if config.GetSettings().DecimalComma {
		o.Decimal = ","
	}
	return o
}

// copyResultText is the result of the history (written as it reads back) as it is copied
func (c *MainForm) copyResultText(plain string) string {
	if r, err := c.session.Calc.Eval(plain); err == nil {
		return calc.Format(r.Value, copyFormat())
	}
	return plain
}

func (c *MainForm) copyText(text string) {
	if text == "" {
		return
	}
	ui.ClipboardSetText(text)
	ui.ShowToast(c, T().Copied+": "+text, ui.ToastSuccess)
}

// copy copies the selection, or else the result
func (c *MainForm) copy() {
	s := c.session
	if s.Line.HasSelection() {
		c.copyText(s.Line.SelectedText())
		return
	}
	if v, ok := s.Value(); ok {
		c.copyText(calc.Format(v, copyFormat()))
	}
}

// copyCalculation copies "expression = result"
func (c *MainForm) copyCalculation() {
	s := c.session
	v, ok := s.Value()
	if !ok {
		return
	}
	expr := s.Line.Text()
	if s.Fresh {
		expr = s.Prev
	}
	c.copyText(expr + " = " + calc.Format(v, copyFormat()))
}

func (c *MainForm) cut() {
	s := c.session
	if !s.Line.HasSelection() {
		return
	}
	c.copyText(s.Line.SelectedText())
	s.Edited()
	s.Line.Insert("")
	c.display.changed()
}

func (c *MainForm) paste() {
	text, err := ui.ClipboardGetText()
	if err != nil {
		return
	}
	if text = normalizePasted(text); text != "" {
		c.typeText(text)
	}
}

// groupedNumber is a number with "," separating the thousands and "." the
// fraction, or the other way round: 1,234.5 or 1.234,5
var groupedNumber = regexp.MustCompile(`^-?\d{1,3}(([,.' ])\d{3})+([,.]\d+)?$`)

// normalizePasted makes a text from another program an expression: one
// line, without the "=" at its end, a number in the US or the European
// writing read as such
func normalizePasted(text string) string {
	text = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(text)
	text = strings.TrimSpace(text)
	text = strings.TrimSpace(strings.TrimSuffix(text, "="))
	if m := groupedNumber.FindStringSubmatch(text); m != nil && m[3] != "" && m[3][0] != m[2][0] {
		text = strings.ReplaceAll(text, m[2], "")
		text = strings.Replace(text, m[3][:1], ".", 1)
	}
	return text
}

// ---- View ----

// setScientific shows or hides the keys of the functions; the window
// grows or shrinks by them
func (c *MainForm) setScientific(on bool) {
	if on == c.keypad.scientific {
		return
	}
	w := c.keypad.Width()
	var delta int
	if on {
		perCol := (w - 3*keyGap) / 4
		delta = 5*perCol + 5*keyGap + blockGap - keyGap
	} else {
		perCol := (w - 8*keyGap - blockGap) / 9
		delta = -(5*perCol + 5*keyGap + blockGap - keyGap)
	}
	c.keypad.SetScientific(on)
	config.UpdateSettings(func(s *config.Settings) { s.Scientific = on })
	c.resizeBy(delta)
}

func (c *MainForm) toggleHistory() {
	show := !c.history.IsVisible()
	c.history.SetVisible(show)
	c.divider.SetVisible(show)
	config.UpdateSettings(func(s *config.Settings) { s.HideHistory = !show })
	if show {
		c.resizeBy(historyWidth + 1)
	} else {
		c.resizeBy(-historyWidth - 1)
	}
}

// resizeBy changes the width of the window that is not maximized
func (c *MainForm) resizeBy(delta int) {
	f := c.Form()
	if f == nil {
		return
	}
	if !f.IsMaximized() {
		w, h := f.Size()
		f.SetSize(max(w+delta, 280), h)
	}
	f.UpdateLayout()
	c.lineChanged()
}

func (c *MainForm) toggleAlwaysOnTop() {
	on := !config.GetSettings().AlwaysOnTop
	config.UpdateSettings(func(s *config.Settings) { s.AlwaysOnTop = on })
	c.Form().SetAlwaysOnTop(on)
}

// ---- Settings and the window ----

// ApplySettings saves the settings and applies what they change in the window
func (c *MainForm) ApplySettings(s config.Settings) {
	languageChanged := s.Language != config.GetSettings().Language
	themeChanged := s.Theme != config.GetSettings().Theme
	if err := config.SetSettings(s); err != nil {
		ui.ShowMessageBox(c, T().Error, err.Error())
	}
	if languageChanged {
		SetLanguage(s.Language)
	}
	if themeChanged {
		ApplyTheme(s.Theme)
	}
	c.lineChanged()
}

// ShowSettings opens the settings dialog
func (c *MainForm) ShowSettings() {
	c.ShowDialog(NewSettingsDialog(config.GetSettings(), c.ApplySettings))
}

// ApplyLanguage updates the texts that do not follow the language by themselves
func (c *MainForm) ApplyLanguage() {
	c.lineChanged()
}

// RestoreWindowState applies the saved window layout before the form is shown.
// Returns whether the window should be maximized once shown.
func (c *MainForm) RestoreWindowState(form *ui.Form) (maximized bool) {
	w := defaultWindowWidth
	if config.GetSettings().Scientific {
		w += 330
	}
	if config.GetSettings().HideHistory {
		w -= historyWidth
	}
	form.SetSize(w, defaultWindowHeight)
	state, ok := config.LoadWindowState()
	if !ok {
		return false
	}
	form.SetSize(state.Width, state.Height)
	// 0,0 means the position was not known (the window manager did not report it)
	if state.X != 0 || state.Y != 0 {
		form.Move(state.X, state.Y)
	}
	return state.Maximized
}

// SaveWindowState remembers the window layout for the next start
func (c *MainForm) SaveWindowState() {
	form := c.Form()
	state, _ := config.LoadWindowState()
	state.Maximized = form.IsMaximized()
	// The size of a maximized window is the screen size: keep the normal one
	if !state.Maximized {
		state.X, state.Y = form.Position()
		state.Width, state.Height = form.Size()
	}
	config.SaveWindowState(state)
}

// OnWindowClose is called when the window is being closed
func (c *MainForm) OnWindowClose() bool {
	c.SaveWindowState()
	return true
}

// quit closes the window
func (c *MainForm) quit() {
	c.SaveWindowState()
	c.Form().Close()
}

// requestExit runs then: there is nothing unsaved to ask about
func (c *MainForm) requestExit(then func()) {
	then()
}

// Activate puts the focus on the display, the window title in place
func (c *MainForm) Activate() {
	c.Form().SetTitle(app.DisplayName)
	if config.GetSettings().AlwaysOnTop {
		c.Form().SetAlwaysOnTop(true)
	}
	c.display.Focus()
}

// BringToFront shows the window on top of the others. Called when the
// application is started again.
func (c *MainForm) BringToFront() {
	raiseWindow(c.Form())
	c.display.Focus()
}

// ShowInstalled tells that this copy has just been installed and started in place of the downloaded one
func (c *MainForm) ShowInstalled() {
	ui.ShowToast(c, T().Installed, ui.ToastSuccess)
}
