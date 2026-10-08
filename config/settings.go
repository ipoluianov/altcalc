// Package config keeps the settings and the history of AltCalc in ~/.altbins/.altcalc
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Settings are the application options and what is remembered between starts
type Settings struct {
	// Language of the interface as a tag like "ru"; "" - the system's
	Language string `json:",omitempty"`

	// Color theme: "light"; "" - the dark one
	Theme string `json:",omitempty"`

	// DecimalComma shows the numbers with "," as the decimal point
	DecimalComma bool `json:",omitempty"`

	// NoGrouping shows the numbers without the thousands separated
	NoGrouping bool `json:",omitempty"`

	// Scientific shows the keys of the functions
	Scientific bool `json:",omitempty"`

	// HideHistory hides the history panel
	HideHistory bool `json:",omitempty"`

	// AlwaysOnTop keeps the window above the others
	AlwaysOnTop bool `json:",omitempty"`

	// Angle is the unit of the angles: 0 degrees, 1 radians, 2 gradians
	Angle int `json:",omitempty"`
}

var (
	settingsMtx sync.Mutex
	settings    = defaultSettings()
)

func defaultSettings() Settings {
	return Settings{}
}

// ConfigDirectory returns ~/.altbins/.altcalc
func ConfigDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".altbins", ".altcalc")
}

func settingsPath() string {
	return filepath.Join(ConfigDirectory(), "settings.json")
}

// LoadSettings reads the settings file; missing values get the defaults
func LoadSettings() {
	s := defaultSettings()
	if bs, err := os.ReadFile(settingsPath()); err == nil {
		_ = json.Unmarshal(bs, &s)
	}
	if s.Angle < 0 || s.Angle > 2 {
		s.Angle = 0
	}
	settingsMtx.Lock()
	settings = s
	settingsMtx.Unlock()
}

// GetSettings returns the current settings; safe to call from any goroutine
func GetSettings() Settings {
	settingsMtx.Lock()
	defer settingsMtx.Unlock()
	return settings
}

// SetSettings applies and saves the settings
func SetSettings(s Settings) error {
	settingsMtx.Lock()
	settings = s
	settingsMtx.Unlock()

	bs, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(ConfigDirectory(), 0755); err != nil {
		return err
	}
	return writeFileAtomic(settingsPath(), bs)
}

// UpdateSettings changes the settings with f and saves them
func UpdateSettings(f func(s *Settings)) error {
	s := GetSettings()
	f(&s)
	return SetSettings(s)
}

// writeFileAtomic writes to a temporary file and renames it over the target,
// so a crash leaves either the old file or the new one, never a torn one
func writeFileAtomic(fullPath string, bs []byte) error {
	tmpPath := fullPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	_, err = f.Write(bs)
	if err == nil {
		err = f.Sync()
	}
	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(tmpPath, fullPath)
	}
	if err != nil {
		os.Remove(tmpPath)
	}
	return err
}
