package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
)

// The appearance belongs to the PC, not to the library: theme.json, in the
// app's data directory, holds it.

type themeFile struct {
	Theme mygo.ThemeSource `json:"theme"`
}

// LoadTheme returns the appearance saved in home, the system's by default.
func LoadTheme(home string) mygo.ThemeSource {
	var f themeFile
	if b, err := os.ReadFile(filepath.Join(home, "theme.json")); err == nil {
		json.Unmarshal(b, &f)
	}
	if validTheme(f.Theme) {
		return f.Theme
	}
	return mygo.ThemeSystem
}

func validTheme(t mygo.ThemeSource) bool {
	return t == mygo.ThemeSystem || t == mygo.ThemeLight || t == mygo.ThemeDark
}

// Theme returns the appearance: "system", "light" or "dark".
func (l *Library) Theme() mygo.ThemeSource {
	return mygo.Theme.Source()
}

// SetTheme changes the appearance of the app and remembers it.
func (l *Library) SetTheme(t mygo.ThemeSource) (mygo.ThemeSource, error) {
	if !validTheme(t) {
		return "", fmt.Errorf("thème inconnu : %q", t)
	}
	if err := saveTheme(l.home, t); err != nil {
		return "", err
	}
	mygo.Theme.SetSource(t)
	return t, nil
}

func saveTheme(home string, t mygo.ThemeSource) error {
	b, _ := json.Marshal(themeFile{Theme: t})
	return os.WriteFile(filepath.Join(home, "theme.json"), b, 0o644)
}
