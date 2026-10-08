package forms

import (
	"image/color"

	"github.com/ipoluianov/nui/ui"
)

// The color themes offered in the settings (config.Settings.Theme)
const (
	themeDark  = ""
	themeLight = "light"
)

// themeColors is a color of the application in the dark and the light theme
type themeColors struct {
	dark, light color.RGBA
}

// get returns the color for the current theme
func (c themeColors) get() color.RGBA {
	if ui.IsDarkTheme {
		return c.dark
	}
	return c.light
}

var (
	colorLink  = themeColors{ui.ColorFromHex("#3d8bf2"), ui.ColorFromHex("#1d6fb5")}
	colorMuted = themeColors{ui.ColorFromHex("#8c8c8c"), ui.ColorFromHex("#6b6b6b")}

	// The accent of AltCalc: the = key, the operators
	colorAccent     = themeColors{ui.ColorFromHex("#7c5cff"), ui.ColorFromHex("#6741d9")}
	colorAccentText = themeColors{ui.ColorFromHex("#ffffff"), ui.ColorFromHex("#ffffff")}

	// The display
	colorDisplay = themeColors{ui.ColorFromHex("#1f1f1f"), ui.ColorFromHex("#ffffff")}
	colorTokOp   = themeColors{ui.ColorFromHex("#b197fc"), ui.ColorFromHex("#6741d9")}
	colorTokFunc = themeColors{ui.ColorFromHex("#74c0fc"), ui.ColorFromHex("#1c7ed6")}
	colorTokName = themeColors{ui.ColorFromHex("#63e6be"), ui.ColorFromHex("#0c8599")}
	colorTokErr  = themeColors{ui.ColorFromHex("#ff6b6b"), ui.ColorFromHex("#e03131")}
	colorParen   = themeColors{ui.ColorFromHex("#8c8c8c"), ui.ColorFromHex("#868e96")}

	// The keys: the digits, the functions, the operators, the memory
	colorKeyDigit     = themeColors{ui.ColorFromHex("#3b3b3b"), ui.ColorFromHex("#ffffff")}
	colorKeyFunc      = themeColors{ui.ColorFromHex("#323232"), ui.ColorFromHex("#f3f3f5")}
	colorKeyOp        = themeColors{ui.ColorFromHex("#352f4a"), ui.ColorFromHex("#ede7ff")}
	colorKeyMem       = themeColors{ui.ColorFromHex("#2b2b2b"), ui.ColorFromHex("#e9e9ec")}
	colorKeyText      = themeColors{ui.ColorFromHex("#f1f1f1"), ui.ColorFromHex("#1f1f1f")}
	colorKeyFuncText  = themeColors{ui.ColorFromHex("#d4d4d4"), ui.ColorFromHex("#343a40")}
	colorKeyMemText   = themeColors{ui.ColorFromHex("#a6a6a6"), ui.ColorFromHex("#5c5f66")}
	colorKeyOpText    = themeColors{ui.ColorFromHex("#c5b3ff"), ui.ColorFromHex("#5f3dc4")}
	colorKeyHoverMix  = themeColors{ui.ColorFromHex("#ffffff"), ui.ColorFromHex("#000000")}
	colorKeyBorder    = themeColors{ui.ColorFromHex("#00000000"), ui.ColorFromHex("#dcdce0")}
	colorHistoryHover = themeColors{ui.ColorFromHex("#333333"), ui.ColorFromHex("#e8e8ec")}
)

// mix returns a moved by t (0..1) towards b
func mix(a, b color.RGBA, t float64) color.RGBA {
	m := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return color.RGBA{m(a.R, b.R), m(a.G, b.G), m(a.B, b.B), m(a.A, b.A)}
}

// linkLabels are recolored when the theme changes, see newLinkLabel
var linkLabels []*ui.Label

// themeListeners are called when the theme changes
var themeListeners []func()

// ApplyTheme switches the application to the theme of the settings and
// repaints the open windows
func ApplyTheme(theme string) {
	if theme == themeLight {
		ui.ApplyLightTheme()
	} else {
		ui.ApplyDarkTheme()
	}
	for _, lbl := range linkLabels {
		lbl.SetForegroundColor(colorLink.get())
	}
	for _, f := range themeListeners {
		f()
	}
}
