package forms

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ipoluianov/nui/ui"
)

// The fonts of the calculator are the system's own where it has good ones:
// the numbers in its monospace font, the keys in its interface font. The
// characters a font lacks (√, ⌫, the superscripts) come from a font of the
// system that has them.
var (
	// fontNumbers draws the expression, the results and the history
	fontNumbers = ui.FontFamilyMono
	// fontKeys draws the labels of the keys
	fontKeys = ui.FontFamilySans
	// fontSymbols draws the labels with the symbols, "" when the system has
	// no font for them
	fontSymbols = ""
)

const (
	familyNumbers = "altcalc-numbers"
	familyKeys    = "altcalc-keys"
	familySymbols = "altcalc-symbols"
)

// symbolRunes are the characters of the keys and the display that the
// fonts built into nui lack
const symbolRunes = "√∛⌫←→ˣʸ⁻"

// monospaceFonts are the monospace fonts tried for the numbers, the best first
func monospaceFonts() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"Cascadia Mono", "Consolas"}
	case "darwin":
		return []string{"SF Mono", "Menlo"}
	}
	// The monospace font of the desktop, then the common ones
	return append(fcFamilies("monospace"), "DejaVu Sans Mono", "Noto Sans Mono", "Ubuntu Mono", "Liberation Mono")
}

// SetupFonts picks the fonts of the system; called once before the window is made
func SetupFonts() {
	for _, name := range monospaceFonts() {
		if ui.RegisterSystemFont(familyNumbers, name) == nil {
			fontNumbers = familyNumbers
			break
		}
	}
	if ui.RegisterSystemFont(familyKeys, ui.SystemUIFontName()) == nil {
		fontKeys = familyKeys
	}
	// The fonts built into nui fall back to a font of the system with the symbols
	if path := symbolFontFile(); path != "" {
		if data, err := os.ReadFile(path); err == nil && ui.RegisterFont(familySymbols, data) == nil {
			ui.AddFallbackFont(familySymbols)
			fontSymbols = familySymbols
		}
	}
}

// symbolFontFile is a font file of the system that has the symbols, "" if none is known
func symbolFontFile() string {
	var candidates []string
	switch runtime.GOOS {
	case "windows":
		dir := filepath.Join(os.Getenv("WINDIR"), "Fonts")
		candidates = []string{filepath.Join(dir, "seguisym.ttf")}
	case "darwin":
		candidates = []string{"/System/Library/Fonts/Apple Symbols.ttf", "/System/Library/Fonts/Supplemental/Arial Unicode.ttf"}
	default:
		// fontconfig finds the font with all the characters
		var hex []string
		for _, r := range symbolRunes {
			hex = append(hex, strconv.FormatInt(int64(r), 16))
		}
		if out, err := exec.Command("fc-match", "-f", "%{file}", ":charset="+strings.Join(hex, " ")).Output(); err == nil {
			candidates = []string{strings.TrimSpace(string(out))}
		}
	}
	for _, p := range candidates {
		// A collection (.ttc) is not read by RegisterFont
		if strings.HasSuffix(strings.ToLower(p), ".ttf") || strings.HasSuffix(strings.ToLower(p), ".otf") {
			if _, err := os.Stat(p); err == nil {
				return p
			}
		}
	}
	return ""
}

// fcFamilies returns the family fontconfig gives for the pattern, as
// "monospace": the one the desktop is set up with
func fcFamilies(pattern string) []string {
	out, err := exec.Command("fc-match", "-f", "%{family[0]}", pattern).Output()
	if err != nil || strings.TrimSpace(string(out)) == "" {
		return nil
	}
	return []string{strings.TrimSpace(string(out))}
}
