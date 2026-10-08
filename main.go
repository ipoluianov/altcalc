package main

import (
	"github.com/ipoluianov/altcalc/app"
	"github.com/ipoluianov/altcalc/config"
	"github.com/ipoluianov/altcalc/forms"
	"github.com/ipoluianov/altcalc/install"
	"github.com/ipoluianov/altcalc/instance"
	"github.com/ipoluianov/nui/ui"
)

func main() {
	install.SetIcon(iconPNG)
	// Started from "Installed apps" to remove it
	if install.HasArg(install.UninstallArg) {
		uninstall(install.HasArg(install.QuietArg))
		return
	}

	// One copy runs: a second start shows its window
	inst, ok := instance.Acquire(config.ConfigDirectory(), nil)
	if !ok {
		return // the running copy shows its window instead
	}
	defer inst.Close()

	config.LoadSettings()
	forms.SetLanguage(config.GetSettings().Language)
	forms.ApplyTheme(config.GetSettings().Theme)
	forms.SetupFonts()
	ui.SetAppIcon(appIcon())
	form := ui.NewForm()
	mainForm := forms.NewMainForm()
	form.Panel().SetPanelPadding(0)
	form.Panel().AddWidget(0, 0, mainForm)
	form.SetMenuBar(mainForm.BuildMenuBar())
	mainForm.AddShortcuts(form)
	maximized := mainForm.RestoreWindowState(form)
	form.OnClose = mainForm.OnWindowClose
	form.SetOnLanguageChanged(mainForm.ApplyLanguage)
	form.Show()
	inst.Serve(func([]string) { form.Invoke(mainForm.BringToFront) })
	// The form is handled by its own goroutine once shown
	form.Invoke(func() {
		if maximized {
			form.Maximize()
		}
		mainForm.Activate()
		if install.HasArg(install.InstalledArg) {
			mainForm.ShowInstalled()
		}
	})
	form.Exec()

	// Installed from this copy: the installed one takes over, so the lock goes first
	if install.RelaunchPending() {
		inst.Close()
		if err := install.StartInstalled(); err != nil {
			install.Inform(app.DisplayName, err.Error())
		}
	}
}
