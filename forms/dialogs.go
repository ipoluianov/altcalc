package forms

import (
	"github.com/ipoluianov/altcalc/app"
	"github.com/ipoluianov/altcalc/config"
	"github.com/ipoluianov/nui/ui"
)

// dialogButtons adds a row of buttons at the right of the panel
func dialogButtons(panel *ui.Panel, buttons ...*ui.Button) {
	panel.AddWidget(0, 0, ui.NewHSpacer())
	for i, b := range buttons {
		panel.AddWidget(0, i+1, b)
	}
}

// dialogLayout adds the content panel, a spacer and the buttons panel to the dialog
func dialogLayout(c *ui.DialogContent) (content, buttons *ui.Panel) {
	content = ui.NewPanel()
	c.AddWidget(0, 0, content)
	c.AddWidget(1, 0, ui.NewVSpacer())
	buttons = ui.NewPanel()
	c.AddWidget(2, 0, buttons)
	return content, buttons
}

// okCancel makes the OK and Cancel buttons of a dialog
func okCancel(c *ui.DialogContent, buttons *ui.Panel, onOK func()) (ok, cancel *ui.Button) {
	ok = ui.NewButton(ui.UIText().OK)
	ok.SetOnClick(onOK)
	cancel = ui.NewButton(ui.UIText().Cancel)
	cancel.SetOnClick(func() { c.Form().RequestClose() })
	dialogButtons(buttons, ok, cancel)
	return ok, cancel
}

// showDialog sets the title, the size and the buttons of a dialog when it is shown
func showDialog(c *ui.DialogContent, title string, w, h int, ok, cancel *ui.Button, focus ui.Widgeter) {
	c.OnDialogShow = func() {
		c.Form().SetTitle(title)
		c.Form().SetSize(w, h)
		c.Form().MoveToCenterOfParent()
		c.Form().SetAcceptButton(ok)
		c.Form().SetCancelButton(cancel)
		if focus != nil {
			focus.Focus()
		}
	}
}

// ---- Settings ----

// NewSettingsDialog edits the application options
func NewSettingsDialog(settings config.Settings, onAccept func(config.Settings)) *ui.DialogContent {
	var c ui.DialogContent
	c.InitWidget()
	content, buttons := dialogLayout(&c)

	cmbLanguage := ui.NewComboBox()
	cmbLanguage.AddItem(T().LanguageSystem, "")
	cmbLanguage.SetSelectedIndex(0)
	for i, l := range languages {
		cmbLanguage.AddItem(l.name, l.tag)
		if l.tag == settings.Language {
			cmbLanguage.SetSelectedIndex(i + 1)
		}
	}
	cmbTheme := ui.NewComboBox()
	cmbTheme.AddItem(T().ThemeDark, themeDark)
	cmbTheme.AddItem(T().ThemeLight, themeLight)
	cmbTheme.SetSelectedIndex(0)
	if settings.Theme == themeLight {
		cmbTheme.SetSelectedIndex(1)
	}
	cmbDecimal := ui.NewComboBox()
	cmbDecimal.AddItem(T().DecimalPoint, false)
	cmbDecimal.AddItem(T().DecimalComma, true)
	cmbDecimal.SetSelectedIndex(0)
	if settings.DecimalComma {
		cmbDecimal.SetSelectedIndex(1)
	}
	chkGrouping := ui.NewCheckbox(T().Grouping)
	chkGrouping.SetChecked(!settings.NoGrouping)

	content.AddWidget(0, 0, ui.NewLabel(T().Language))
	content.AddWidget(0, 1, cmbLanguage)
	content.AddWidget(1, 0, ui.NewLabel(T().Theme))
	content.AddWidget(1, 1, cmbTheme)
	content.AddWidget(2, 0, ui.NewLabel(T().DecimalSeparator))
	content.AddWidget(2, 1, cmbDecimal)
	content.AddWidget(3, 1, chkGrouping)
	content.AddWidget(0, 2, ui.NewHSpacer())

	ok, cancel := okCancel(&c, buttons, func() {
		s := config.GetSettings()
		s.Language, _ = cmbLanguage.SelectedItemData().(string)
		s.Theme, _ = cmbTheme.SelectedItemData().(string)
		s.DecimalComma, _ = cmbDecimal.SelectedItemData().(bool)
		s.NoGrouping = !chkGrouping.Checked()
		c.Form().Close()
		c.RunInParent(func() { onAccept(s) })
	})
	showDialog(&c, T().Settings, 460, 260, ok, cancel, nil)
	return &c
}

// ---- Quick reference ----

// NewReferenceDialog shows what can be typed: the operators, the functions, the keys
func NewReferenceDialog() *ui.DialogContent {
	var c ui.DialogContent
	c.InitWidget()
	text := ui.NewTextBox()
	text.SetMultiline(true)
	text.SetReadOnly(true)
	text.SetFontFamily(fontNumbers)
	text.SetFontSize(13)
	text.SetText(T().Reference)
	text.SetXExpandable(true)
	text.SetYExpandable(true)
	// The text takes all the height, the button stays under it
	c.AddWidget(0, 0, text)
	buttons := ui.NewPanel()
	c.AddWidget(1, 0, buttons)
	btnClose := ui.NewButton(T().Close)
	btnClose.SetOnClick(func() { c.Form().Close() })
	dialogButtons(buttons, btnClose)
	showDialog(&c, T().ReferenceTitle, 720, 600, btnClose, btnClose, btnClose)
	return &c
}

// ---- About ----

func NewAboutDialog() *ui.DialogContent {
	var c ui.DialogContent
	c.InitWidget()
	content, buttons := dialogLayout(&c)
	lines := []string{
		T().Version + " " + app.Version,
		T().Author + " " + app.Author,
		app.Copyright(),
		T().License + " " + app.License,
		app.Website,
	}
	lblName := ui.NewLabel(app.DisplayName)
	lblName.SetFontSize(24)
	lblName.SetTextAlign(ui.HAlignCenter)
	content.AddWidget(0, 0, lblName)
	for i, s := range lines {
		l := ui.NewLabel(s)
		l.SetTextAlign(ui.HAlignCenter)
		content.AddWidget(i+1, 0, l)
	}
	btnWebsite := ui.NewButton(T().VisitWebsite)
	btnWebsite.SetOnClick(func() {
		if err := app.OpenSiteURL(app.Website, "about_dialog"); err != nil {
			ui.ShowMessageBox(&c, T().Error, err.Error())
		}
	})
	btnClose := ui.NewButton(T().Close)
	btnClose.SetOnClick(func() { c.Form().Close() })
	dialogButtons(buttons, btnWebsite, btnClose)
	showDialog(&c, T().AboutTitle(app.DisplayName), 400, 300, btnClose, btnClose, btnClose)
	return &c
}
