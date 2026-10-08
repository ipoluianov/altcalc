package forms

import (
	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/altcalc/config"
	"github.com/ipoluianov/nui/ui"
)

// checked puts a check mark before the text of an item that is on
func checked(on bool, text string) string {
	if on {
		return "✓ " + text
	}
	return text
}

// BuildMenuBar makes the main menu with its shortcuts
func (c *MainForm) BuildMenuBar() *ui.MenuBar {
	bar := ui.NewMenuBar()
	t := T
	// The texts are set again when a menu opens: the check marks follow the state
	texts := map[*ui.ContextMenu][]func(){}
	item := func(m *ui.ContextMenu, text func() string, shortcut string, f func()) *ui.ContextMenuItem {
		it := m.AddItem(text(), f)
		it.SetTextFunc(text)
		if shortcut != "" {
			it.SetShortcut(shortcut)
		}
		texts[m] = append(texts[m], func() { it.SetText(text()) })
		return it
	}
	menu := func(text func() string) *ui.ContextMenu {
		m := ui.NewContextMenu(nil)
		bar.AddMenuItem(text(), m).SetTextFunc(text)
		m.SetOnShow(func() {
			for _, f := range texts[m] {
				f()
			}
		})
		return m
	}
	lineEdit := func(f func(l *EditLine) bool) func() {
		return func() {
			c.session.Edited()
			f(&c.session.Line)
			c.display.changed()
		}
	}

	// ---- Edit ----
	edit := menu(func() string { return t().MenuEdit })
	item(edit, func() string { return t().CmdUndo }, "Mod+Z", lineEdit((*EditLine).Undo))
	item(edit, func() string { return t().CmdRedo }, "Mod+Y", lineEdit((*EditLine).Redo))
	edit.AddSeparator()
	item(edit, func() string { return t().CmdCut }, "Mod+X", c.cut)
	item(edit, func() string { return t().CmdCopy }, "Mod+C", c.copy)
	item(edit, func() string { return t().CmdCopyCalculation }, "Mod+Shift+C", c.copyCalculation)
	item(edit, func() string { return t().CmdPaste }, "Mod+V", c.paste)
	edit.AddSeparator()
	item(edit, func() string { return t().CmdSelectAll }, "Mod+A", lineEdit(func(l *EditLine) bool { l.SelectAll(); return true }))
	item(edit, func() string { return t().CmdClear }, "", c.clear)
	edit.AddSeparator()
	item(edit, func() string { return t().CmdClearHistory }, "Mod+Shift+Delete", c.clearHistory)

	// ---- Memory ----
	mem := menu(func() string { return t().MenuMemory })
	item(mem, func() string { return t().CmdMemoryClear }, "Mod+L", c.memoryClear)
	item(mem, func() string { return t().CmdMemoryRecall }, "Mod+R", c.memoryRecall)
	item(mem, func() string { return t().CmdMemoryAdd }, "Mod+P", func() { c.memoryAdd(1) })
	item(mem, func() string { return t().CmdMemorySubtract }, "Mod+Q", func() { c.memoryAdd(-1) })
	item(mem, func() string { return t().CmdMemoryStore }, "Mod+M", c.memoryStore)

	// ---- View ----
	view := menu(func() string { return t().MenuView })
	item(view, func() string { return checked(!c.keypad.scientific, t().CmdStandard) }, "Mod+1", func() { c.setScientific(false) })
	item(view, func() string { return checked(c.keypad.scientific, t().CmdScientific) }, "Mod+2", func() { c.setScientific(true) })
	view.AddSeparator()
	item(view, func() string { return checked(c.history.IsVisible(), t().CmdHistory) }, "Mod+H", c.toggleHistory)
	item(view, func() string { return checked(config.GetSettings().AlwaysOnTop, t().CmdAlwaysOnTop) }, "Mod+T", c.toggleAlwaysOnTop)
	view.AddSeparator()
	for m := calc.Degrees; m <= calc.Gradians; m++ {
		item(view, func() string { return checked(c.session.Calc.Angle == m, t().Angles[m]) }, "", func() { c.setAngle(m) })
	}
	item(view, func() string { return t().CmdNextAngle }, "F2", c.nextAngle)

	// ---- Help ----
	help := menu(func() string { return t().MenuHelp })
	item(help, func() string { return t().CmdReference }, "F1", func() { c.ShowDialog(NewReferenceDialog()) })
	item(help, func() string { return t().Help }, "", func() { openDocs(c, "help_menu") })
	item(help, func() string { return t().Settings + "..." }, "", c.ShowSettings)
	help.AddSeparator()
	item(help, func() string { return t().About + "..." }, "", func() { c.ShowDialog(NewAboutDialog()) })
	return bar
}

// AddShortcuts adds the keys that are not in the menu
func (c *MainForm) AddShortcuts(form *ui.Form) {
	form.AddShortcut("F9", c.negate)
	form.AddShortcut("Shift+Insert", c.paste)
	form.AddShortcut("Mod+Insert", c.copy)
	// The keyboard always types into the display: a click on the history or
	// the status bar takes the focus from it, the next key gives it back.
	// An open menu keeps its keys.
	form.SetOnGlobalKeyDown(func(key ui.Key, mods ui.KeyModifiers) bool {
		if form.TopPopupWidget() == nil && !c.display.IsFocused() {
			c.display.Focus()
		}
		return false
	})
}
