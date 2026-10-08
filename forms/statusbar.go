package forms

import (
	"github.com/ipoluianov/altcalc/app"
	"github.com/ipoluianov/altcalc/calc"
	"github.com/ipoluianov/altcalc/config"
	"github.com/ipoluianov/altcalc/install"
	"github.com/ipoluianov/nui/ui"
)

// StatusBar is the bottom row: the unit of the angles, the memory, and the
// links as in the other AltBins utilities
type StatusBar struct {
	ui.Widget
	main *MainForm

	lblAngle  *ui.Label
	lblMemory *statusCell

	// What the install link does: install, update or uninstall
	installStatus install.Status
}

func NewStatusBar(main *MainForm) *StatusBar {
	var c StatusBar
	c.InitWidget()
	c.main = main
	c.SetPanelPadding(4)
	col := 0
	add := func(w ui.Widgeter) {
		c.AddWidget(0, col, w)
		col++
	}
	gap := func(width int) {
		s := ui.NewSpace()
		s.SetSize(width, 0)
		add(s)
	}
	gap(6)
	// The unit of the angles: a click switches to the next one
	c.lblAngle = newLinkLabel(func() string { return T().Angles[main.session.Calc.Angle] }, main.nextAngle)
	c.lblAngle.SetTooltipFunc(func() string { return T().CmdNextAngle + " (F2)" })
	add(c.lblAngle)
	gap(16)
	c.lblMemory = newStatusCell(200)
	add(c.lblMemory)
	add(ui.NewHSpacer())

	links := []*ui.Label{
		newLinkLabel(func() string { return T().Settings }, func() { main.ShowSettings() }),
		newLinkLabel(func() string { return T().Help }, func() { openDocs(&c, "help") }),
	}
	// A downloaded copy offers to install or update itself, the installed one to be removed
	c.installStatus = install.CurrentStatus()
	if c.installStatus != install.StatusNone {
		links = append(links, newLinkLabel(c.installText, c.onInstallLink))
	}
	links = append(links, newLinkLabel(func() string { return T().About }, c.onAbout))
	for i, lbl := range links {
		if i > 0 {
			gap(12)
		}
		add(lbl)
	}
	gap(6)
	return &c
}

// Refresh shows the unit of the angles and the memory
func (c *StatusBar) Refresh() {
	s := c.main.session
	if text := T().Angles[s.Calc.Angle]; c.lblAngle.Text() != text {
		c.lblAngle.SetText(text)
	}
	if s.HasMemory {
		c.lblMemory.SetText(T().StatusMemory(calc.Format(s.Memory, displayFormat())))
	} else {
		c.lblMemory.SetText("")
	}
}

// statusCell is a text of the status bar that changes often, e.g. with the
// mouse: it has a fixed width and is just repainted, as a label would lay
// the window out again on every change
type statusCell struct {
	ui.Widget
	text string
}

func newStatusCell(width int) *statusCell {
	var c statusCell
	c.InitWidget()
	c.SetMinWidth(width)
	c.SetMaxWidth(width)
	c.SetMinHeight(ui.ThemeControlHeight())
	c.SetMaxHeight(ui.ThemeControlHeight())
	c.SetOnPaint(func(cnv *ui.Canvas) {
		cnv.SetHAlign(ui.HAlignLeft)
		cnv.SetVAlign(ui.VAlignCenter)
		cnv.SetFontFamily(ui.ThemeFontFamily())
		cnv.SetFontSize(ui.ThemeFontSize())
		cnv.SetColor(colorMuted.get())
		cnv.DrawText(0, 0, c.Width(), c.Height(), c.text)
	})
	return &c
}

func (c *statusCell) SetText(text string) {
	if c.text == text {
		return
	}
	c.text = text
	if f := c.Form(); f != nil {
		f.Update()
	}
}

// newLinkLabel creates a hyperlink-style label with the text from text() that calls onClick on left click.
// It is underlined only under the mouse, so a row of links stays quiet.
func newLinkLabel(text func() string, onClick func()) *ui.Label {
	lbl := ui.NewLabel("")
	lbl.SetTextFunc(text)
	lbl.SetForegroundColor(colorLink.get())
	lbl.SetOnMouseEnter(func() { lbl.SetUnderline(true) })
	lbl.SetOnMouseLeave(func() { lbl.SetUnderline(false) })
	linkLabels = append(linkLabels, lbl)
	lbl.SetMouseCursor(ui.MouseCursorPointer)
	lbl.SetOnMouseDown(func(button ui.MouseButton, x int, y int, mods ui.KeyModifiers) bool {
		if button != ui.MouseButtonLeft {
			return false
		}
		onClick()
		return true
	})
	return lbl
}

// openDocs opens the docs on the site; campaign tells which place in the app the visit came from
func openDocs(parent ui.Widgeter, campaign string) {
	if err := app.OpenSiteURL(app.DocsURL, campaign); err != nil {
		ui.ShowMessageBox(parent, T().Error, err.Error())
	}
}

func (c *StatusBar) onAbout() {
	c.ShowDialog(NewAboutDialog())
}

// installText is the text of the install link for what it does now
func (c *StatusBar) installText() string {
	switch c.installStatus {
	case install.StatusUpdate:
		return T().Update
	case install.StatusUninstall:
		return T().Uninstall
	}
	return T().Install
}

func (c *StatusBar) onInstallLink() {
	if c.installStatus == install.StatusUninstall {
		c.onUninstall()
		return
	}
	c.onInstall()
}

// onInstall copies the application to ~/.altbins and registers it, then
// quits for the installed copy to start (see main). Over an older installed
// version it is the same: that one is replaced.
func (c *StatusBar) onInstall() {
	ui.ShowQuestionMessageBoxOKCancel(c, c.installText(), T().InstallAsk(install.Dir()), func() {
		c.main.requestExit(func() {
			if err := install.Install(); err != nil {
				ui.ShowMessageBox(c, T().Error, T().InstallFailed(err.Error()))
				return
			}
			install.RelaunchAfterExit()
			c.main.quit()
		})
	}, nil)
}

// onUninstall removes the installed copy; the settings stay.
// When that is this copy, it quits and its binary is deleted once it has.
func (c *StatusBar) onUninstall() {
	ui.ShowQuestionMessageBoxOKCancel(c, T().Uninstall, T().UninstallAsk(config.ConfigDirectory()), func() {
		installed := install.IsInstalledCopy()
		if installed {
			c.main.requestExit(func() {
				if err := install.Uninstall(); err != nil {
					ui.ShowMessageBox(c, T().Error, err.Error())
					return
				}
				c.main.quit()
			})
			return
		}
		if err := install.Uninstall(); err != nil {
			ui.ShowMessageBox(c, T().Error, err.Error())
			return
		}
		c.installStatus = install.CurrentStatus()
		ui.ShowToast(c, T().Uninstalled, ui.ToastSuccess)
	}, nil)
}
