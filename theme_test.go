package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo"
)

func TestTheme(t *testing.T) {
	home := t.TempDir()
	if got := LoadTheme(home); got != mygo.ThemeSystem {
		t.Errorf("default: %q", got)
	}
	if err := saveTheme(home, mygo.ThemeDark); err != nil {
		t.Fatal(err)
	}
	if got := LoadTheme(home); got != mygo.ThemeDark {
		t.Errorf("saved: %q", got)
	}

	// A damaged or unknown choice falls back to the system's.
	os.WriteFile(filepath.Join(home, "theme.json"), []byte(`{"theme":"sepia"}`), 0o644)
	if got := LoadTheme(home); got != mygo.ThemeSystem {
		t.Errorf("unknown: %q", got)
	}
	l, _ := OpenChosenLibrary(home)
	if _, err := l.SetTheme("sepia"); err == nil {
		t.Error("unknown theme accepted")
	}
}
