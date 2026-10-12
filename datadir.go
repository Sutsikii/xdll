package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
)

// The library lives in the app's data directory unless the user chose
// another one: data.json, in the app's data directory, then holds its path.

type pointer struct {
	Dir string `json:"dir"`
}

// OpenChosenLibrary opens the library of the directory chosen in home, or of
// home itself. When the chosen directory is gone, such as an unplugged
// drive, it opens home's and remembers the missing one.
func OpenChosenLibrary(home string) (*Library, error) {
	var p pointer
	if b, err := os.ReadFile(filepath.Join(home, "data.json")); err == nil {
		json.Unmarshal(b, &p)
	}
	if p.Dir != "" {
		if st, err := os.Stat(p.Dir); err == nil && st.IsDir() {
			l, err := OpenLibrary(p.Dir)
			if err == nil {
				l.home = home
			}
			return l, err
		}
	}
	l, err := OpenLibrary(home)
	if err == nil {
		l.missing = p.Dir
	}
	return l, err
}

// DataDirChoice is a directory picked for the library.
type DataDirChoice struct {
	Dir        string `json:"dir"`
	HasLibrary bool   `json:"hasLibrary"` // it holds a library already
}

// ChooseDataDir asks for a directory for the library. Dir is "" when
// canceled. A directory that is neither empty nor a library gets an xdll
// subdirectory.
func (l *Library) ChooseDataDir(ctx context.Context) (DataDirChoice, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent:            mygo.CallerWindow(ctx),
		Title:             "Dossier des données",
		ButtonLabel:       "Choisir",
		Directory:         true,
		CreateDirectories: true,
	})
	if err != nil || len(paths) == 0 {
		return DataDirChoice{}, err
	}
	return l.dataDirChoice(paths[0])
}

func (l *Library) dataDirChoice(dir string) (DataDirChoice, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return DataDirChoice{}, err
	}
	if isLibrary(dir) {
		return DataDirChoice{Dir: dir, HasLibrary: true}, nil
	}
	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		dir = filepath.Join(dir, "xdll")
	}
	return DataDirChoice{Dir: dir, HasLibrary: isLibrary(dir)}, nil
}

// DefaultDataDir is the app's data directory, to go back to.
func (l *Library) DefaultDataDir() DataDirChoice {
	return DataDirChoice{Dir: l.home, HasLibrary: isLibrary(l.home)}
}

// MoveData moves the library to dir and uses it from there.
func (l *Library) MoveData(dir string) (Settings, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.checkMove(dir); err != nil {
		return l.settings(), err
	}
	if isLibrary(dir) {
		return l.settings(), errors.New("ce dossier contient déjà une bibliothèque")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return l.settings(), err
	}
	from := l.root()
	var moved []string // copied, to remove from the old place at the end
	undo := func() {
		for _, name := range moved {
			os.RemoveAll(filepath.Join(dir, name))
		}
	}
	for _, name := range []string{"media", "library.json"} {
		src, dst := filepath.Join(from, name), filepath.Join(dir, name)
		if _, err := os.Stat(src); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		os.RemoveAll(dst) // an empty media directory
		if err := copyTree(src, dst); err != nil {
			undo()
			return l.settings(), err
		}
		moved = append(moved, name)
	}
	if err := l.use(dir); err != nil {
		undo()
		return l.settings(), err
	}
	for _, name := range moved {
		os.RemoveAll(filepath.Join(from, name))
	}
	return l.settings(), nil
}

// UseData switches to the library in dir, leaving the current one where it
// is.
func (l *Library) UseData(dir string) (Settings, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := l.checkMove(dir); err != nil {
		return l.settings(), err
	}
	data, err := loadLibrary(dir)
	if err != nil {
		return l.settings(), err
	}
	if err := l.use(dir); err != nil {
		return l.settings(), err
	}
	l.data = data
	return l.settings(), nil
}

// checkMove refuses a move while posts download, or into the library
// itself. The caller holds l.mu.
func (l *Library) checkMove(dir string) error {
	if len(l.inflight) > 0 {
		return errors.New("attendre la fin des téléchargements")
	}
	if !filepath.IsAbs(dir) {
		return errors.New("chemin invalide")
	}
	if sameDir(dir, l.root()) {
		return errors.New("c'est déjà le dossier des données")
	}
	if rel, err := filepath.Rel(l.root(), dir); err == nil && filepath.IsLocal(rel) {
		return errors.New("choisir un dossier en dehors du dossier actuel")
	}
	return nil
}

// use records dir as the library's directory. The caller holds l.mu.
func (l *Library) use(dir string) error {
	ptr := filepath.Join(l.home, "data.json")
	var err error
	if sameDir(dir, l.home) {
		err = os.Remove(ptr)
		if errors.Is(err, fs.ErrNotExist) {
			err = nil
		}
	} else {
		b, _ := json.Marshal(pointer{Dir: dir})
		err = os.WriteFile(ptr, b, 0o644)
	}
	if err != nil {
		return err
	}
	l.dir.Store(&dir)
	l.missing = ""
	return nil
}

func (l *Library) settings() Settings {
	return Settings{
		Dir:     l.root(),
		Default: sameDir(l.root(), l.home),
		Home:    l.home,
		Missing: l.missing,
		Version: mygo.App.Version(),
	}
}

func isLibrary(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "library.json"))
	return err == nil
}

func sameDir(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// copyTree copies a file or a directory.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(p, target)
	})
}
