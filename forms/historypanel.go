package forms

import (
	"strings"

	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/nui/ui"
)

// HistoryPanel lists the calculations, the newest at the top, and the
// variables under them. A click puts the result into the line, a double
// click the expression; the right click has the menu.
type HistoryPanel struct {
	ui.Widget
	main *MainForm
	list *ui.Widget

	hover int // the row under the mouse, -1 - none
	menu  *ui.ContextMenu
}

const (
	historyRowHeight = 50
	varRowHeight     = 28
	headingHeight    = 30
	historyWidth     = 270
)

// historyRow is a row of the list: an entry of the history (entry >= 0),
// the heading of the variables, or a variable
type historyRow struct {
	y, h    int
	entry   int
	heading bool
	name    string
}

func NewHistoryPanel(main *MainForm) *HistoryPanel {
	var c HistoryPanel
	c.InitWidget()
	c.main = main
	c.hover = -1
	c.SetPanelPadding(0)
	c.SetCellPadding(0)
	c.SetMinWidth(historyWidth)
	c.SetMaxWidth(historyWidth)
	c.SetYExpandable(true)

	head := ui.NewPanel()
	head.SetPanelPadding(6)
	title := ui.NewLabel("")
	title.SetTextFunc(func() string { return T().PanelHistory })
	title.SetForegroundColor(colorMuted.get())
	themeListeners = append(themeListeners, func() { title.SetForegroundColor(colorMuted.get()) })
	head.AddWidget(0, 0, title)
	head.AddWidget(0, 1, ui.NewHSpacer())
	head.AddWidget(0, 2, newLinkLabel(func() string { return T().ClearLink }, main.clearHistory))
	c.AddWidget(0, 0, head)

	c.list = &ui.Widget{}
	c.list.InitWidget()
	c.list.SetXExpandable(true)
	c.list.SetYExpandable(true)
	c.list.SetAllowScroll(false, true)
	c.list.SetAutoFillBackground(true)
	c.list.SetRole("base")
	c.list.SetOnPaint(c.paint)
	c.list.SetOnMouseMove(func(x, y int, mods ui.KeyModifiers) bool {
		if row := c.rowAt(y); row != c.hover {
			c.hover = row
			c.update()
		}
		return false
	})
	c.list.SetOnMouseLeave(func() {
		c.hover = -1
		c.update()
	})
	c.list.SetOnMouseDown(func(button ui.MouseButton, x, y int, mods ui.KeyModifiers) bool {
		c.hover = c.rowAt(y)
		if button == ui.MouseButtonLeft {
			c.use(c.hover, false)
		}
		return button == ui.MouseButtonLeft
	})
	c.list.SetOnMouseDblClick(func(button ui.MouseButton, x, y int, mods ui.KeyModifiers) bool {
		if button == ui.MouseButtonLeft {
			c.use(c.rowAt(y), true)
		}
		return true
	})
	c.menu = ui.NewContextMenu(nil)
	c.menu.SetOnShow(c.fillMenu)
	c.list.SetContextMenu(c.menu)
	c.AddWidget(1, 0, c.list)
	return &c
}

func (c *HistoryPanel) update() {
	if f := c.Form(); f != nil {
		f.Update()
	}
}

// rows lays the list out: the history from the newest, then the variables
func (c *HistoryPanel) rows() []historyRow {
	s := c.main.session
	var rows []historyRow
	y := 0
	for i := len(s.History) - 1; i >= 0; i-- {
		rows = append(rows, historyRow{y: y, h: historyRowHeight, entry: i})
		y += historyRowHeight
	}
	if names := s.Calc.VarNames(); len(names) > 0 {
		rows = append(rows, historyRow{y: y, h: headingHeight, entry: -1, heading: true})
		y += headingHeight
		for _, n := range names {
			rows = append(rows, historyRow{y: y, h: varRowHeight, entry: -1, name: n})
			y += varRowHeight
		}
	}
	return rows
}

func (c *HistoryPanel) rowAt(y int) int {
	for i, r := range c.rows() {
		if y >= r.y && y < r.y+r.h && !r.heading {
			return i
		}
	}
	return -1
}

// use puts the row into the line: the result, or with expr the expression of an entry
func (c *HistoryPanel) use(row int, expr bool) {
	rows := c.rows()
	if row < 0 || row >= len(rows) {
		return
	}
	r := rows[row]
	s := c.main.session
	switch {
	case r.name != "":
		c.main.useText(r.name)
	case expr:
		s.Line.SetText(s.History[r.entry].Expr)
		s.Edited()
		c.main.lineChanged()
	default:
		c.main.useValue(s.History[r.entry].Result)
	}
	c.main.display.Focus()
}

func (c *HistoryPanel) fillMenu() {
	c.menu.RemoveAllItems()
	rows := c.rows()
	row := c.hover
	s := c.main.session
	if row >= 0 && row < len(rows) {
		r := rows[row]
		if r.name != "" {
			name := r.name
			c.menu.AddItem(T().InsertResult, func() { c.use(row, false) })
			c.menu.AddItem(T().CopyResult, func() {
				if v, ok := s.Calc.Var(name); ok {
					c.main.copyText(calc.Format(v, copyFormat()))
				}
			})
			c.menu.AddSeparator()
			c.menu.AddItem(T().Delete, func() {
				s.Calc.DeleteVar(name)
				s.changed()
				c.main.lineChanged()
			})
		} else {
			e := s.History[r.entry]
			entry := r.entry
			c.menu.AddItem(T().InsertResult, func() { c.use(row, false) })
			c.menu.AddItem(T().InsertExpression, func() { c.use(row, true) })
			c.menu.AddSeparator()
			c.menu.AddItem(T().CopyResult, func() { c.main.copyText(c.main.copyResultText(e.Result)) })
			c.menu.AddItem(T().CmdCopyCalculation, func() {
				c.main.copyText(e.Expr + " = " + c.main.copyResultText(e.Result))
			})
			c.menu.AddSeparator()
			c.menu.AddItem(T().Delete, func() {
				s.DeleteEntry(entry)
				c.main.lineChanged()
			})
		}
		c.menu.AddSeparator()
	}
	c.menu.AddItem(T().CmdClearHistory, c.main.clearHistory)
}

// Refresh lays the list out for the history, the newest entry in sight
func (c *HistoryPanel) Refresh() {
	rows := c.rows()
	h := 0
	if n := len(rows); n > 0 {
		h = rows[n-1].y + rows[n-1].h
	}
	if c.list.InnerHeight() != h {
		c.list.SetInnerSize(c.list.Width(), h)
		c.list.SetScrollY(0)
	}
	c.update()
}

func (c *HistoryPanel) paint(cnv *ui.Canvas) {
	s := c.main.session
	pal := ui.CurrentPalette()
	w := c.list.Width()
	muted := colorMuted.get()
	rows := c.rows()
	if len(rows) == 0 {
		cnv.SetHAlign(ui.HAlignCenter)
		cnv.SetVAlign(ui.VAlignTop)
		cnv.SetFontFamily(ui.ThemeFontFamily())
		cnv.SetFontSize(ui.ThemeFontSize())
		cnv.SetColor(muted)
		cnv.DrawText(0, 24, w, 30, T().HistoryEmpty)
		return
	}
	format := displayFormat()
	inner := w - 28
	for i, r := range rows {
		if i == c.hover {
			cnv.FillRect(0, r.y, w, r.h, colorHistoryHover.get())
		}
		cnv.SetVAlign(ui.VAlignCenter)
		switch {
		case r.heading:
			cnv.FillRect(12, r.y+4, w-24, 1, pal.Divider)
			cnv.SetHAlign(ui.HAlignLeft)
			cnv.SetFontFamily(ui.ThemeFontFamily())
			cnv.SetFontSize(ui.ThemeFontSize())
			cnv.SetColor(muted)
			cnv.DrawText(14, r.y+4, inner, r.h-4, T().PanelVariables)
		case r.name != "":
			v, _ := s.Calc.Var(r.name)
			cnv.SetFontFamily(fontNumbers)
			cnv.SetFontSize(13)
			cnv.SetHAlign(ui.HAlignLeft)
			cnv.SetColor(colorTokName.get())
			cnv.DrawText(14, r.y, inner, r.h, r.name)
			nameW, _, _ := ui.MeasureText(fontNumbers, 13, r.name+" ")
			cnv.SetHAlign(ui.HAlignRight)
			cnv.SetColor(pal.Text)
			cnv.DrawText(14+nameW, r.y, inner-nameW, r.h, fitLeft(fontNumbers, 13, calc.Format(v, format), inner-nameW))
		default:
			e := s.History[r.entry]
			cnv.SetHAlign(ui.HAlignRight)
			cnv.SetFontFamily(fontNumbers)
			cnv.SetFontSize(12)
			cnv.SetColor(muted)
			expr := e.Expr
			if !strings.ContainsRune(expr, '=') {
				expr += " ="
			}
			cnv.DrawText(14, r.y+5, inner, 18, fitLeft(fontNumbers, 12, expr, inner))
			result := e.Result
			if res, err := s.Calc.Eval(e.Result); err == nil {
				result = calc.Format(res.Value, format)
			}
			cnv.SetFontSize(17)
			cnv.SetColor(pal.Text)
			cnv.DrawText(14, r.y+22, inner, 24, fitLeft(fontNumbers, 17, result, inner))
			cnv.FillRect(12, r.y+r.h-1, w-24, 1, pal.Divider)
		}
	}
}
