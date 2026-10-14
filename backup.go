package main

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo"
)

// A backup is a zip holding library.json, prefs.json (the page's settings)
// and media/<post id>/<file>.

// Settings is what the settings page shows.
type Settings struct {
	Dir     string `json:"dir"`
	Default bool   `json:"default"` // Dir is the app's data directory
	Home    string `json:"home"`    // the app's data directory
	Missing string `json:"missing"` // the chosen directory, not found at launch
	Version string `json:"version"`
}

// BackupResult describes a written backup.
type BackupResult struct {
	File  string `json:"file"`
	Posts int    `json:"posts"`
	Files int    `json:"files"`
}

// RestoreResult describes an imported backup.
type RestoreResult struct {
	Library Snapshot `json:"library"`
	Added   int      `json:"added"`
	Skipped int      `json:"skipped"` // already in the library
	Prefs   string   `json:"prefs"`
}

// Settings returns the data directory and the app version.
func (l *Library) Settings() Settings {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.settings()
}

// OpenDataDir opens the data directory in the explorer.
func (l *Library) OpenDataDir() error {
	return mygo.Shell.OpenPath(l.root())
}

// Backup asks where to save and writes the whole library to a zip. prefs is
// stored as is and given back by Restore. File is "" when canceled.
func (l *Library) Backup(ctx context.Context, prefs string) (BackupResult, error) {
	file, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      mygo.CallerWindow(ctx),
		Title:       "Sauvegarder la bibliothèque",
		DefaultPath: "xdll-" + time.Now().Format("2006-01-02") + ".zip",
		ButtonLabel: "Sauvegarder",
		Filters:     []mygo.FileFilter{{Name: "Sauvegarde xdll", Extensions: []string{"zip"}}},
	})
	if err != nil || file == "" {
		return BackupResult{}, err
	}
	if filepath.Ext(file) == "" {
		file += ".zip"
	}
	r, err := l.writeBackup(file, prefs)
	if err == nil {
		mygo.Shell.ShowItemInFolder(file)
	}
	return r, err
}

// Restore asks for a backup and imports it. replace drops the current
// library first; otherwise posts already present are kept as they are.
// It returns nil when canceled.
func (l *Library) Restore(ctx context.Context, replace bool) (*RestoreResult, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent:      mygo.CallerWindow(ctx),
		Title:       "Importer une sauvegarde",
		ButtonLabel: "Importer",
		Filters:     []mygo.FileFilter{{Name: "Sauvegarde xdll", Extensions: []string{"zip"}}},
	})
	if err != nil || len(paths) == 0 {
		return nil, err
	}
	r, err := l.restore(paths[0], replace)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (l *Library) writeBackup(file, prefs string) (BackupResult, error) {
	l.mu.Lock()
	snap := l.snapshot()
	lib, err := json.MarshalIndent(l.data, "", "  ")
	l.mu.Unlock()
	if err != nil {
		return BackupResult{}, err
	}

	tmp := file + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return BackupResult{}, err
	}
	defer os.Remove(tmp)
	zw := zip.NewWriter(f)
	r := BackupResult{File: file, Posts: len(snap.Bookmarks)}
	write := func(name string, method uint16, src io.Reader) error {
		w, err := zw.CreateHeader(&zip.FileHeader{Name: name, Method: method, Modified: time.Now()})
		if err == nil {
			_, err = io.Copy(w, src)
		}
		return err
	}
	err = write("library.json", zip.Deflate, strings.NewReader(string(lib)))
	if err == nil && prefs != "" {
		err = write("prefs.json", zip.Deflate, strings.NewReader(prefs))
	}
	for _, b := range snap.Bookmarks {
		for _, name := range mediaFiles(b) {
			if err != nil {
				break
			}
			var in *os.File
			in, err = os.Open(filepath.Join(l.mediaDir(), filepath.FromSlash(name)))
			if err != nil {
				err = fmt.Errorf("%s : %w", name, err)
				break
			}
			// Media are already compressed.
			err = write("media/"+name, zip.Store, in)
			in.Close()
			r.Files++
		}
	}
	if err == nil {
		err = zw.Close()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return BackupResult{}, err
	}
	return r, os.Rename(tmp, file)
}

func (l *Library) restore(file string, replace bool) (RestoreResult, error) {
	zr, err := zip.OpenReader(file)
	if err != nil {
		return RestoreResult{}, fmt.Errorf("sauvegarde illisible : %w", err)
	}
	defer zr.Close()
	entries := map[string]*zip.File{}
	for _, f := range zr.File {
		entries[f.Name] = f
	}
	var in Snapshot
	if err := readJSON(entries["library.json"], &in); err != nil {
		return RestoreResult{}, err
	}
	res := RestoreResult{}
	if p := entries["prefs.json"]; p != nil {
		if b, err := readAll(p); err == nil {
			res.Prefs = string(b)
		}
	}

	// Extract the media into a staging directory, then move them in place
	// under the lock.
	l.mu.Lock()
	var todo []Bookmark
	for _, b := range in.Bookmarks {
		if !validID(b.ID) {
			l.mu.Unlock()
			return RestoreResult{}, fmt.Errorf("post invalide : %q", b.ID)
		}
		if !replace && l.find(b.ID) >= 0 {
			res.Skipped++
			continue
		}
		todo = append(todo, b)
	}
	l.mu.Unlock()

	staging := filepath.Join(l.root(), "import.tmp")
	os.RemoveAll(staging)
	defer os.RemoveAll(staging)
	for _, b := range todo {
		for _, name := range mediaFiles(b) {
			if path.Clean(name) != name || !strings.HasPrefix(name, b.ID+"/") || !filepath.IsLocal(filepath.FromSlash(name)) {
				return RestoreResult{}, fmt.Errorf("chemin invalide : %q", name)
			}
			f := entries["media/"+name]
			if f == nil {
				return RestoreResult{}, fmt.Errorf("sauvegarde incomplète : %s manque", name)
			}
			if err := extract(f, filepath.Join(staging, filepath.FromSlash(name))); err != nil {
				return RestoreResult{}, err
			}
		}
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	prev := l.data
	old := filepath.Join(l.root(), "media.old")
	if replace {
		os.RemoveAll(old)
		if err := os.Rename(l.mediaDir(), old); err != nil {
			return RestoreResult{}, err
		}
		l.data = Snapshot{}
	}
	var moved []string
	// fail puts the previous library back.
	fail := func(err error) (RestoreResult, error) {
		l.data = prev
		if replace {
			os.RemoveAll(l.mediaDir())
			os.Rename(old, l.mediaDir())
		}
		for _, d := range moved {
			os.RemoveAll(d)
		}
		return RestoreResult{}, err
	}
	if err := os.MkdirAll(l.mediaDir(), 0o755); err != nil {
		return fail(err)
	}
	l.data.Folders = slices.Clone(l.data.Folders)
	l.data.Bookmarks = slices.Clone(l.data.Bookmarks)
	folders := l.mergeFolders(in.Folders)
	for _, b := range todo {
		if l.find(b.ID) >= 0 { // added meanwhile
			res.Skipped++
			continue
		}
		if len(b.Media) > 0 {
			dst := filepath.Join(l.mediaDir(), b.ID)
			os.RemoveAll(dst)
			if err := os.Rename(filepath.Join(staging, b.ID), dst); err != nil {
				return fail(err)
			}
			moved = append(moved, dst)
		}
		b.FolderID = folders[b.FolderID]
		l.data.Bookmarks = append(l.data.Bookmarks, b)
		res.Added++
	}
	if err := l.save(); err != nil {
		return fail(err)
	}
	os.RemoveAll(old)
	res.Library = l.snapshot()
	return res, nil
}

// mergeFolders adds the imported folders, reusing folders with the same
// name in the same parent, and maps imported folder IDs to library ones.
// The caller holds l.mu.
func (l *Library) mergeFolders(in []Folder) map[string]string {
	in = cleanFolders(slices.Clone(in))
	ids := map[string]string{"": ""}
	var merge func(f Folder) string
	merge = func(f Folder) string {
		if id, ok := ids[f.ID]; ok {
			return id
		}
		parent := ""
		if p := slices.IndexFunc(in, func(g Folder) bool { return g.ID == f.Parent }); p >= 0 {
			parent = merge(in[p])
		}
		name := strings.TrimSpace(f.Name)
		if name == "" {
			ids[f.ID] = parent
			return parent
		}
		i := slices.IndexFunc(l.data.Folders, func(g Folder) bool { return g.Parent == parent && strings.EqualFold(g.Name, name) })
		if i >= 0 {
			ids[f.ID] = l.data.Folders[i].ID
			return ids[f.ID]
		}
		id := f.ID
		if l.hasFolder(id) { // empty or taken
			id = l.newFolderID()
		}
		l.data.Folders = append(l.data.Folders, Folder{ID: id, Name: name, Parent: parent})
		ids[f.ID] = id
		return id
	}
	for _, f := range in {
		merge(f)
	}
	return ids
}

// mediaFiles lists the files of a bookmark, relative to the media directory.
func mediaFiles(b Bookmark) []string {
	var out []string
	for _, m := range b.Media {
		out = append(out, m.File)
		if m.Thumb != "" {
			out = append(out, m.Thumb)
		}
	}
	return out
}

func readAll(f *zip.File) ([]byte, error) {
	r, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}

func readJSON(f *zip.File, v any) error {
	if f == nil {
		return errors.New("ce fichier n'est pas une sauvegarde xdll")
	}
	b, err := readAll(f)
	if err == nil {
		err = json.Unmarshal(b, v)
	}
	if err != nil {
		return fmt.Errorf("library.json : %w", err)
	}
	return nil
}

func extract(f *zip.File, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return fmt.Errorf("%s : %w", f.Name, err)
	}
	return out.Close()
}
