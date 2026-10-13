package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
)

// Media is a downloaded photo, video or GIF.
type Media struct {
	Kind   string `json:"kind"` // "image", "video" or "gif"
	File   string `json:"file"` // relative to the media directory, e.g. "1234/1.mp4"
	Thumb  string `json:"thumb,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Size   int64  `json:"size"`
	Light  bool   `json:"light,omitempty"` // a video downloaded in economy quality
}

// Bookmark is a saved post and its media.
type Bookmark struct {
	ID         string    `json:"id"`
	Source     string    `json:"source,omitempty"` // "x", "youtube" or "instagram"; X when empty (older posts)
	URL        string    `json:"url"`
	Author     string    `json:"author"`
	AuthorName string    `json:"authorName"`
	Text       string    `json:"text"`
	PostedAt   time.Time `json:"postedAt"`
	AddedAt    time.Time `json:"addedAt"`
	FolderID   string    `json:"folderId"` // "" when not filed
	Favorite   bool      `json:"favorite"`
	Note       string    `json:"note"`
	Media      []Media   `json:"media"`
}

// Folder groups bookmarks.
type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Snapshot is the whole library.
type Snapshot struct {
	Folders   []Folder   `json:"folders"`
	Bookmarks []Bookmark `json:"bookmarks"`
}

// Library stores bookmarks in library.json and media in media/.
type Library struct {
	mu       sync.Mutex
	dir      atomic.Pointer[string] // changes when the user moves the data
	home     string                 // the app's data directory, which holds data.json
	missing  string                 // the chosen directory, when it was not found at launch
	data     Snapshot
	inflight map[string]bool
	opt      optimizer
}

// OpenLibrary loads or creates the library in dir.
func OpenLibrary(dir string) (*Library, error) {
	l := &Library{home: dir, inflight: map[string]bool{}}
	data, err := loadLibrary(dir)
	if err != nil {
		return nil, err
	}
	l.data = data
	l.dir.Store(&dir)
	return l, nil
}

// loadLibrary reads the library in dir, creating its media directory.
func loadLibrary(dir string) (Snapshot, error) {
	var data Snapshot
	if err := os.MkdirAll(filepath.Join(dir, "media"), 0o755); err != nil {
		return data, err
	}
	b, err := os.ReadFile(filepath.Join(dir, "library.json"))
	if errors.Is(err, os.ErrNotExist) {
		return data, nil
	}
	if err == nil {
		err = json.Unmarshal(b, &data)
	}
	if err != nil {
		return data, fmt.Errorf("library.json : %w", err)
	}
	return data, nil
}

func (l *Library) root() string     { return *l.dir.Load() }
func (l *Library) file() string     { return filepath.Join(l.root(), "library.json") }
func (l *Library) mediaDir() string { return filepath.Join(l.root(), "media") }

// save writes the library; the caller holds l.mu.
func (l *Library) save() error {
	b, err := json.MarshalIndent(l.data, "", "  ")
	if err != nil {
		return err
	}
	tmp := l.file() + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, l.file())
}

func (l *Library) snapshot() Snapshot {
	return Snapshot{
		Folders:   append([]Folder{}, l.data.Folders...),
		Bookmarks: append([]Bookmark{}, l.data.Bookmarks...),
	}
}

func (l *Library) find(id string) int {
	return slices.IndexFunc(l.data.Bookmarks, func(b Bookmark) bool { return b.ID == id })
}

func (l *Library) hasFolder(id string) bool {
	return id == "" || slices.ContainsFunc(l.data.Folders, func(f Folder) bool { return f.ID == id })
}

// update applies fn to a bookmark and saves.
func (l *Library) update(id string, fn func(*Bookmark) error) (Snapshot, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	i := l.find(id)
	if i < 0 {
		return l.snapshot(), errors.New("post introuvable")
	}
	if err := fn(&l.data.Bookmarks[i]); err != nil {
		return l.snapshot(), err
	}
	return l.snapshot(), l.save()
}

// List returns the folders and bookmarks.
func (l *Library) List() Snapshot {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.snapshot()
}

// Video qualities: the best the site has, or lighter videos.
const (
	QualityMax     = "max"
	QualityEconomy = "economy"
)

// Add downloads the media of the post at url into folderID, in quality.
func (l *Library) Add(ctx context.Context, url, folderID, quality string) (Bookmark, error) {
	id, err := parsePostID(url)
	if err != nil {
		return Bookmark{}, err
	}
	l.mu.Lock()
	switch {
	case l.find(id) >= 0:
		l.mu.Unlock()
		return Bookmark{}, errors.New("déjà enregistré")
	case l.inflight[id]:
		l.mu.Unlock()
		return Bookmark{}, errors.New("déjà en cours")
	}
	l.inflight[id] = true
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		delete(l.inflight, id)
		l.mu.Unlock()
	}()

	t, err := fetchPost(ctx, id, quality == QualityEconomy)
	if err != nil {
		return Bookmark{}, err
	}
	if len(t.Media) == 0 {
		return Bookmark{}, errors.New("aucun média")
	}
	dir := filepath.Join(l.mediaDir(), id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Bookmark{}, err
	}
	b := Bookmark{
		ID: id, Source: t.Source, URL: t.URL, Author: t.Author, AuthorName: t.AuthorName,
		Text: t.Text, PostedAt: t.Created, AddedAt: time.Now().UTC(), FolderID: folderID,
	}
	for i, m := range t.Media {
		name := strconv.Itoa(i+1) + extOf(m.URL, m.Kind)
		n, err := downloadMedia(ctx, m, filepath.Join(dir, name))
		if err != nil {
			os.RemoveAll(dir)
			return Bookmark{}, err
		}
		media := Media{Kind: m.Kind, File: id + "/" + name, Width: m.Width, Height: m.Height, Size: n, Light: m.Light}
		if media.Width == 0 {
			media.Width, media.Height = mediaSize(filepath.Join(dir, name), m.Kind)
		}
		if m.Thumb != "" {
			thumb := strconv.Itoa(i+1) + ".thumb" + extOf(m.Thumb, "image")
			if _, err := download(ctx, m.Thumb, filepath.Join(dir, thumb)); err == nil {
				media.Thumb = id + "/" + thumb
			}
		}
		b.Media = append(b.Media, media)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.hasFolder(b.FolderID) {
		b.FolderID = ""
	}
	l.data.Bookmarks = append(l.data.Bookmarks, b)
	return b, l.save()
}

// Delete removes a bookmark and its files.
func (l *Library) Delete(id string) (Snapshot, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if i := l.find(id); i >= 0 {
		l.data.Bookmarks = slices.Delete(l.data.Bookmarks, i, i+1)
		os.RemoveAll(filepath.Join(l.mediaDir(), id))
	}
	return l.snapshot(), l.save()
}

// Move files bookmarks into a folder ("" for none).
func (l *Library) Move(ids []string, folderID string) (Snapshot, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.hasFolder(folderID) {
		return l.snapshot(), errors.New("dossier introuvable")
	}
	for i := range l.data.Bookmarks {
		if slices.Contains(ids, l.data.Bookmarks[i].ID) {
			l.data.Bookmarks[i].FolderID = folderID
		}
	}
	return l.snapshot(), l.save()
}

// SetFavorite stars or unstars a bookmark.
func (l *Library) SetFavorite(id string, favorite bool) (Snapshot, error) {
	return l.update(id, func(b *Bookmark) error { b.Favorite = favorite; return nil })
}

// SetNote sets a bookmark's note.
func (l *Library) SetNote(id, note string) (Snapshot, error) {
	return l.update(id, func(b *Bookmark) error { b.Note = strings.TrimSpace(note); return nil })
}

// CreateFolder adds a folder.
func (l *Library) CreateFolder(name string) (Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Folder{}, errors.New("nom vide")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	f := Folder{ID: strconv.FormatInt(time.Now().UnixNano(), 36), Name: name}
	l.data.Folders = append(l.data.Folders, f)
	return f, l.save()
}

// RenameFolder renames a folder.
func (l *Library) RenameFolder(id, name string) (Snapshot, error) {
	name = strings.TrimSpace(name)
	l.mu.Lock()
	defer l.mu.Unlock()
	if name == "" {
		return l.snapshot(), errors.New("nom vide")
	}
	for i := range l.data.Folders {
		if l.data.Folders[i].ID == id {
			l.data.Folders[i].Name = name
		}
	}
	return l.snapshot(), l.save()
}

// DeleteFolder removes a folder; its bookmarks become unfiled.
func (l *Library) DeleteFolder(id string) (Snapshot, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.data.Folders = slices.DeleteFunc(l.data.Folders, func(f Folder) bool { return f.ID == id })
	for i := range l.data.Bookmarks {
		if l.data.Bookmarks[i].FolderID == id {
			l.data.Bookmarks[i].FolderID = ""
		}
	}
	return l.snapshot(), l.save()
}

// ExportResult describes an export.
type ExportResult struct {
	Files      int `json:"files"`
	Downloaded int `json:"downloaded"` // videos downloaded again in maximum quality
	Light      int `json:"light"`      // economy videos exported as they are: gone from their site
}

// Export copies media to a chosen directory. With original, economy videos
// are downloaded again in maximum quality instead, when still online.
func (l *Library) Export(ctx context.Context, ids []string, original bool) (ExportResult, error) {
	paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent:            mygo.CallerWindow(ctx),
		Title:             "Exporter",
		ButtonLabel:       "Exporter ici",
		Directory:         true,
		CreateDirectories: true,
	})
	if err != nil || len(paths) == 0 {
		return ExportResult{}, err
	}
	r, err := l.exportTo(ctx, ids, paths[0], original)
	if r.Files > 0 {
		mygo.Shell.OpenPath(paths[0])
	}
	return r, err
}

func (l *Library) exportTo(ctx context.Context, ids []string, dir string, original bool) (ExportResult, error) {
	l.mu.Lock()
	var bookmarks []Bookmark
	for _, b := range l.data.Bookmarks {
		if slices.Contains(ids, b.ID) {
			bookmarks = append(bookmarks, b)
		}
	}
	l.mu.Unlock()
	var r ExportResult
	for _, b := range bookmarks {
		var p *post // the post in maximum quality, when needed
		if original && slices.ContainsFunc(b.Media, func(m Media) bool { return m.Light }) {
			p, _ = fetchPost(ctx, b.ID, false)
		}
		for i, m := range b.Media {
			dst := filepath.Join(dir, fmt.Sprintf("%s_%s_%d%s", cmpOr(b.Author, "x"), b.ID, i+1, filepath.Ext(m.File)))
			if m.Light && original {
				if p != nil && len(p.Media) == len(b.Media) && p.Media[i].Kind == m.Kind {
					if _, err := downloadMedia(ctx, p.Media[i], dst); err == nil {
						r.Files++
						r.Downloaded++
						continue
					}
				}
				if ctx.Err() != nil {
					return r, ctx.Err()
				}
				r.Light++
			}
			if err := copyFile(filepath.Join(l.mediaDir(), filepath.FromSlash(m.File)), dst); err != nil {
				return r, err
			}
			r.Files++
		}
	}
	return r, nil
}

// Reveal shows a bookmark's files in the explorer.
func (l *Library) Reveal(id string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if i := l.find(id); i >= 0 && len(l.data.Bookmarks[i].Media) > 0 {
		mygo.Shell.ShowItemInFolder(filepath.Join(l.mediaDir(), filepath.FromSlash(l.data.Bookmarks[i].Media[0].File)))
	}
}

// OpenPost opens the post in the browser.
func (l *Library) OpenPost(id string) error {
	l.mu.Lock()
	i := l.find(id)
	var u string
	if i >= 0 {
		u = l.data.Bookmarks[i].URL
	}
	l.mu.Unlock()
	if u == "" {
		return errors.New("post introuvable")
	}
	return mygo.Shell.OpenExternal(u)
}

// OpenMedia opens a media file.
func (l *Library) OpenMedia(file string) error {
	p := filepath.Join(l.mediaDir(), filepath.FromSlash(file))
	if rel, err := filepath.Rel(l.mediaDir(), p); err != nil || strings.HasPrefix(rel, "..") {
		return errors.New("chemin invalide")
	}
	return mygo.Shell.OpenPath(p)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
