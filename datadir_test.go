package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMoveData(t *testing.T) {
	home := t.TempDir()
	l, _ := OpenChosenLibrary(home)
	fakePost(t, l, "1", "")
	other := filepath.Join(t.TempDir(), "xdll data")

	s, err := l.MoveData(other)
	if err != nil || s.Dir != other || s.Default {
		t.Fatalf("move: %+v %v", s, err)
	}
	if got, _ := os.ReadFile(filepath.Join(other, "media", "1", "1.jpg")); string(got) != "image 1" {
		t.Errorf("media not moved: %q", got)
	}
	if isLibrary(home) || exists(filepath.Join(home, "media")) {
		t.Error("old data left behind")
	}
	fakePost(t, l, "2", "") // saved in the new place

	// The next launch opens the chosen directory.
	re, err := OpenChosenLibrary(home)
	if err != nil || re.root() != other || len(re.List().Bookmarks) != 2 {
		t.Fatalf("reopen: %s %+v %v", re.root(), re.List(), err)
	}

	// Back to the default directory.
	if s, err := re.MoveData(home); err != nil || !s.Default {
		t.Fatalf("move back: %+v %v", s, err)
	}
	if exists(filepath.Join(home, "data.json")) || !exists(filepath.Join(home, "media", "2", "1.jpg")) {
		t.Error("move back")
	}
}

func TestUseDataAndChoice(t *testing.T) {
	home, shared := t.TempDir(), t.TempDir()
	other, _ := OpenLibrary(shared)
	fakePost(t, other, "7", "")

	l, _ := OpenChosenLibrary(home)
	fakePost(t, l, "1", "")
	c, err := l.dataDirChoice(shared)
	if err != nil || !c.HasLibrary || c.Dir != shared {
		t.Fatalf("choice: %+v %v", c, err)
	}
	if _, err := l.MoveData(shared); err == nil {
		t.Error("moved onto a library")
	}
	if _, err := l.UseData(shared); err != nil {
		t.Fatal(err)
	}
	if b := l.List().Bookmarks; len(b) != 1 || b[0].ID != "7" {
		t.Errorf("use: %+v", b)
	}
	if !isLibrary(home) {
		t.Error("the previous library was removed")
	}

	// A folder with other files gets an xdll subfolder.
	busy := t.TempDir()
	os.WriteFile(filepath.Join(busy, "notes.txt"), nil, 0o644)
	if c, _ := l.dataDirChoice(busy); c.Dir != filepath.Join(busy, "xdll") || c.HasLibrary {
		t.Errorf("busy folder: %+v", c)
	}
	if _, err := l.MoveData(filepath.Join(shared, "media", "sub")); err == nil {
		t.Error("moved inside the library")
	}
}

func TestMissingDataDir(t *testing.T) {
	home := t.TempDir()
	gone := filepath.Join(t.TempDir(), "unplugged")
	os.WriteFile(filepath.Join(home, "data.json"), []byte(`{"dir":`+quote(gone)+`}`), 0o644)
	l, err := OpenChosenLibrary(home)
	if err != nil || l.root() != home || l.Settings().Missing != gone {
		t.Fatalf("%s %+v %v", l.root(), l.Settings(), err)
	}
	if exists(gone) {
		t.Error("created the missing directory")
	}
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
