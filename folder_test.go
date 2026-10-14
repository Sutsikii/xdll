package main

import (
	"path/filepath"
	"testing"
)

func TestNestedFolders(t *testing.T) {
	l, _ := OpenLibrary(t.TempDir())
	design, _ := l.CreateFolder("Design", "")
	ui, _ := l.CreateFolder("UI/UX", design.ID)
	icons, _ := l.CreateFolder("Icônes", ui.ID)
	if _, err := l.CreateFolder("x", "nope"); err == nil {
		t.Error("created a folder in a missing parent")
	}
	if _, err := l.MoveFolder(design.ID, icons.ID); err == nil {
		t.Error("moved a folder inside itself")
	}
	if _, err := l.MoveFolder(design.ID, design.ID); err == nil {
		t.Error("moved a folder into itself")
	}
	if s, err := l.MoveFolder(icons.ID, design.ID); err != nil || s.Folders[2].Parent != design.ID {
		t.Fatalf("move: %+v %v", s.Folders, err)
	}

	// Deleting UI/UX puts its posts in Design.
	fakePost(t, l, "1", ui.ID)
	s, _ := l.DeleteFolder(ui.ID)
	if len(s.Folders) != 2 || s.Bookmarks[0].FolderID != design.ID {
		t.Errorf("delete: %+v", s)
	}
	s, _ = l.DeleteFolder(design.ID)
	if s.Folders[0].Parent != "" || s.Bookmarks[0].FolderID != "" {
		t.Errorf("delete top: %+v", s)
	}
}

func TestCleanFolders(t *testing.T) {
	got := cleanFolders([]Folder{
		{ID: "a", Parent: "b"},
		{ID: "b", Parent: "a"},
		{ID: "c", Parent: "gone"},
		{ID: "d", Parent: "c"},
	})
	if got[0].Parent != "" || got[1].Parent != "a" || got[2].Parent != "" || got[3].Parent != "c" {
		t.Errorf("%+v", got)
	}
}

func TestRestoreNestedFolders(t *testing.T) {
	src, _ := OpenLibrary(t.TempDir())
	design, _ := src.CreateFolder("Design", "")
	ui, _ := src.CreateFolder("UI", design.ID)
	src.CreateFolder("UI", "") // same name, other parent
	fakePost(t, src, "1", ui.ID)
	file := filepath.Join(t.TempDir(), "backup.zip")
	if _, err := src.writeBackup(file, ""); err != nil {
		t.Fatal(err)
	}

	dst, _ := OpenLibrary(t.TempDir())
	d2, _ := dst.CreateFolder("design", "")
	res, err := dst.restore(file, false)
	if err != nil {
		t.Fatal(err)
	}
	f := res.Library.Folders
	if len(f) != 3 {
		t.Fatalf("%+v", f)
	}
	b := res.Library.Bookmarks[0]
	i := (&Library{data: res.Library}).folder(b.FolderID)
	if i < 0 || f[i].Name != "UI" || f[i].Parent != d2.ID {
		t.Errorf("post in %+v, folders %+v", b.FolderID, f)
	}
}
