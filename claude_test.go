package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClaudeText(t *testing.T) {
	l := &Library{dir: `C:\data`}
	b := Bookmark{
		ID: "42", URL: "https://x.com/nasa/status/42", Author: "nasa", AuthorName: "NASA",
		Text: "Hello Mars", Note: "pour le projet", PostedAt: time.Date(2026, 10, 9, 14, 30, 0, 0, time.Local),
		Media: []Media{{Kind: "image", File: "42/1.jpg"}, {Kind: "video", File: "42/2.mp4"}},
	}
	got := l.claudeText(b)
	for _, want := range []string{
		"Post X de NASA (@nasa), 09/10/2026 14:30",
		"https://x.com/nasa/status/42",
		"Hello Mars",
		"Note : pour le projet",
		filepath.Join(`C:\data`, "media", "42", "1.jpg") + " (image)",
		filepath.Join(`C:\data`, "media", "42", "2.mp4") + " (video)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestPNGOfScalesDown(t *testing.T) {
	path := filepath.Join(t.TempDir(), "big.jpg")
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 3000, 1500)), nil); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(path, buf.Bytes(), 0o644)
	b, err := pngOf(path)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	if s := img.Bounds().Size(); s.X != 2000 || s.Y != 1000 {
		t.Errorf("size = %v, want 2000x1000", s)
	}
}

func TestClipboardDataModes(t *testing.T) {
	l := &Library{dir: t.TempDir()}
	os.MkdirAll(filepath.Join(l.mediaDir(), "42"), 0o755)
	var buf bytes.Buffer
	png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 10, 10)))
	os.WriteFile(filepath.Join(l.mediaDir(), "42", "1.png"), buf.Bytes(), 0o644)
	b := Bookmark{ID: "42", Author: "n", Media: []Media{
		{Kind: "image", File: "42/1.png"},
		{Kind: "video", File: "42/2.mp4"}, // no thumbnail
	}}
	for _, c := range []struct {
		index   int
		what    string
		copied  string
		formats string
	}{
		{0, "all", "all", "text/plain image/png"},
		{0, "text", "text", "text/plain"},
		{0, "image", "image", "image/png"},
		{1, "all", "text", "text/plain"},
		{1, "image", "", ""},
	} {
		data, copied, err := l.clipboardData(b, c.index, c.what)
		if c.copied == "" {
			if err == nil {
				t.Errorf("%d %s: no error", c.index, c.what)
			}
			continue
		}
		var formats []string
		for _, f := range data.Formats() {
			formats = append(formats, string(f))
		}
		if err != nil || copied != c.copied || strings.Join(formats, " ") != c.formats {
			t.Errorf("%d %s: %q %v %v, want %q %q", c.index, c.what, copied, formats, err, c.copied, c.formats)
		}
	}
}
