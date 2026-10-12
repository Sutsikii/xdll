package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// fakePost adds a bookmark with one image and its file, as Add would.
func fakePost(t *testing.T, l *Library, id, folderID string) {
	t.Helper()
	dir := filepath.Join(l.mediaDir(), id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1.jpg"), []byte("image "+id), 0o644); err != nil {
		t.Fatal(err)
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.data.Bookmarks = append(l.data.Bookmarks, Bookmark{
		ID: id, Author: "a", AddedAt: time.Now().UTC(), FolderID: folderID,
		Media: []Media{{Kind: "image", File: id + "/1.jpg", Size: 7}},
	})
	if err := l.save(); err != nil {
		t.Fatal(err)
	}
}

func TestBackupRestore(t *testing.T) {
	src, _ := OpenLibrary(t.TempDir())
	f, _ := src.CreateFolder("Art")
	fakePost(t, src, "1", f.ID)
	fakePost(t, src, "2", "")
	src.SetNote("2", "note")
	file := filepath.Join(t.TempDir(), "backup.zip")
	r, err := src.writeBackup(file, `{"zoom":300}`)
	if err != nil || r.Posts != 2 || r.Files != 2 {
		t.Fatalf("backup: %+v %v", r, err)
	}

	// Merge into a library that already has post 2 and a folder named "art".
	dst, _ := OpenLibrary(t.TempDir())
	art, _ := dst.CreateFolder("art")
	fakePost(t, dst, "2", "")
	fakePost(t, dst, "3", "")
	res, err := dst.restore(file, false)
	if err != nil || res.Added != 1 || res.Skipped != 1 || res.Prefs != `{"zoom":300}` {
		t.Fatalf("merge: %+v %v", res, err)
	}
	if n := len(res.Library.Folders); n != 1 {
		t.Errorf("merge: %d folders, want 1", n)
	}
	b := res.Library.Bookmarks
	if len(b) != 3 || b[2].ID != "1" || b[2].FolderID != art.ID || b[0].Note != "" {
		t.Errorf("merge: %+v", b)
	}
	if got, _ := os.ReadFile(filepath.Join(dst.mediaDir(), "1", "1.jpg")); string(got) != "image 1" {
		t.Errorf("merge: media %q", got)
	}

	// Replace.
	res, err = dst.restore(file, true)
	if err != nil || res.Added != 2 || len(res.Library.Bookmarks) != 2 || res.Library.Bookmarks[1].Note != "note" {
		t.Fatalf("replace: %+v %v", res, err)
	}
	if _, err := os.Stat(filepath.Join(dst.mediaDir(), "3")); !os.IsNotExist(err) {
		t.Error("replace kept the media of post 3")
	}
	if _, err := os.Stat(filepath.Join(dst.root(), "media.old")); !os.IsNotExist(err) {
		t.Error("media.old left behind")
	}
	re, err := OpenLibrary(dst.root())
	if err != nil || len(re.List().Bookmarks) != 2 || re.List().Bookmarks[0].FolderID != re.List().Folders[0].ID {
		t.Fatalf("reload: %+v %v", re.List(), err)
	}
}

func TestRestoreRejectsBadBackups(t *testing.T) {
	l, _ := OpenLibrary(t.TempDir())
	fakePost(t, l, "5", "")
	notZip := filepath.Join(t.TempDir(), "x.zip")
	os.WriteFile(notZip, []byte("nope"), 0o644)
	if _, err := l.restore(notZip, true); err == nil {
		t.Error("restored a non-zip file")
	}

	// Bad media paths and missing media leave the library untouched.
	for name, lib := range map[string]string{
		"missing": `{"bookmarks":[{"id":"9","media":[{"kind":"image","file":"9/1.jpg"}]}]}`,
		"escape":  `{"bookmarks":[{"id":"9","media":[{"kind":"image","file":"9/../../evil"}]}]}`,
		"other":   `{"bookmarks":[{"id":"9","media":[{"kind":"image","file":"9/../5/1.jpg"}]}]}`,
		"bad id":  `{"bookmarks":[{"id":"..","media":[]}]}`,
	} {
		file := filepath.Join(t.TempDir(), "b.zip")
		out, _ := os.Create(file)
		zw := zip.NewWriter(out)
		w, _ := zw.Create("library.json")
		w.Write([]byte(lib))
		w, _ = zw.Create("media/9/../../evil")
		w.Write([]byte("x"))
		zw.Close()
		out.Close()
		if _, err := l.restore(file, true); err == nil {
			t.Errorf("%s: restored", name)
		}
		if s := l.List(); len(s.Bookmarks) != 1 || s.Bookmarks[0].ID != "5" {
			t.Errorf("%s: library changed: %+v", name, s)
		}
		if _, err := os.Stat(filepath.Join(l.mediaDir(), "5", "1.jpg")); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}
