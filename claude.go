package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// maxSide bounds copied images: Claude downsizes larger ones anyway.
const maxSide = 2000

// CopyForClaude puts a post on the clipboard. what is "text" (the post and
// the paths of its files), "image" (media index, a video's thumbnail) or
// "all" (both). It returns what was copied: "text", "image" or "all".
func (l *Library) CopyForClaude(id string, index int, what string) (string, error) {
	l.mu.Lock()
	i := l.find(id)
	var b Bookmark
	if i >= 0 {
		b = l.data.Bookmarks[i]
	}
	l.mu.Unlock()
	if i < 0 {
		return "", errors.New("post introuvable")
	}

	data, copied, err := l.clipboardData(b, index, what)
	if err != nil {
		return "", err
	}
	return copied, writeClipboard(data)
}

// CopyFiles puts the media files of posts on the clipboard, to paste them
// in the explorer or a chat. index picks one media of a single post; -1
// takes them all. It returns how many files were copied.
func (l *Library) CopyFiles(ids []string, index int) (int, error) {
	l.mu.Lock()
	var paths []string
	for _, id := range ids {
		i := l.find(id)
		if i < 0 {
			continue
		}
		for j, m := range l.data.Bookmarks[i].Media {
			if index < 0 || index == j {
				paths = append(paths, filepath.Join(l.mediaDir(), filepath.FromSlash(m.File)))
			}
		}
	}
	l.mu.Unlock()
	if len(paths) == 0 {
		return 0, errors.New("aucun fichier")
	}
	data, err := transfer.FileData(paths...)
	if err != nil {
		return 0, err
	}
	return len(paths), writeClipboard(data)
}

// ClipboardText returns the text on the clipboard, "" when there is none.
func (l *Library) ClipboardText() string {
	b, _ := mygo.Clipboard.ReadFormat(transfer.Text)
	return string(b)
}

// writeClipboard replaces the clipboard. Another program may hold it open
// for a moment.
func writeClipboard(data transfer.Data) error {
	for try := 0; ; try++ {
		err := mygo.Clipboard.Write(data)
		if err == nil {
			return nil
		}
		if try == 9 {
			return fmt.Errorf("presse-papier occupé : %w", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// clipboardData builds what CopyForClaude writes.
func (l *Library) clipboardData(b Bookmark, index int, what string) (transfer.Data, string, error) {
	var reps []transfer.Representation
	copied := ""
	if what != "image" {
		reps = append(reps, transfer.Bytes(transfer.Text, []byte(l.claudeText(b))))
		copied = "text"
	}
	if what != "text" && index >= 0 && index < len(b.Media) {
		m := b.Media[index]
		file := m.File
		if m.Kind != "image" {
			file = m.Thumb
		}
		if file != "" {
			img, err := pngOf(filepath.Join(l.mediaDir(), filepath.FromSlash(file)))
			if err != nil {
				return transfer.Data{}, "", err
			}
			reps = append(reps, transfer.Bytes(transfer.PNG, img))
			if copied == "text" {
				copied = "all"
			} else {
				copied = "image"
			}
		}
	}
	if len(reps) == 0 {
		return transfer.Data{}, "", errors.New("pas d'image pour ce média")
	}
	return transfer.New(transfer.NewItem(reps...)), copied, nil
}

func (l *Library) claudeText(b Bookmark) string {
	var s strings.Builder
	kind := map[string]string{"youtube": "Vidéo YouTube", "instagram": "Post Instagram", "pinterest": "Épingle Pinterest"}[b.Source]
	fmt.Fprintf(&s, "%s de %s (@%s)", cmpOr(kind, "Post X"), cmpOr(b.AuthorName, b.Author), b.Author)
	if !b.PostedAt.IsZero() {
		fmt.Fprintf(&s, ", %s", b.PostedAt.Local().Format("02/01/2006 15:04"))
	}
	fmt.Fprintf(&s, "\n%s\n", b.URL)
	if b.Text != "" {
		fmt.Fprintf(&s, "\n%s\n", b.Text)
	}
	if b.Note != "" {
		fmt.Fprintf(&s, "\nNote : %s\n", b.Note)
	}
	s.WriteString("\nFichiers :\n")
	for _, m := range b.Media {
		fmt.Fprintf(&s, "- %s (%s)\n", filepath.Join(l.mediaDir(), filepath.FromSlash(m.File)), m.Kind)
	}
	return s.String()
}

// pngOf decodes an image file and encodes it as PNG, scaled down to maxSide.
func pngOf(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%s : %w", filepath.Base(path), err)
	}
	if w, h := img.Bounds().Dx(), img.Bounds().Dy(); max(w, h) > maxSide {
		scale := float64(maxSide) / float64(max(w, h))
		dst := image.NewRGBA(image.Rect(0, 0, max(1, int(float64(w)*scale)), max(1, int(float64(h)*scale))))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
		img = dst
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
