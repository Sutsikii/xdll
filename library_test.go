package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTweetID(t *testing.T) {
	for in, want := range map[string]string{
		"https://x.com/elonmusk/status/1585341984679469056": "1585341984679469056",
		"https://twitter.com/a/status/123456789?s=20&t=abc": "123456789",
		"https://mobile.twitter.com/a/statuses/42/photo/1":  "42",
		"https://fxtwitter.com/a/status/987654321/video/1":  "987654321",
		"  1585341984679469056 ":                            "1585341984679469056",
		"https://x.com/i/web/status/1585341984679469056":    "1585341984679469056",
	} {
		got, err := parseTweetID(in)
		if err != nil || got != want {
			t.Errorf("parseTweetID(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "https://x.com/elonmusk", "hello"} {
		if _, err := parseTweetID(in); err == nil {
			t.Errorf("parseTweetID(%q) succeeded", in)
		}
	}
}

func TestParsePostID(t *testing.T) {
	for in, want := range map[string]string{
		"https://x.com/a/status/123456789":                            "123456789",
		"https://www.youtube.com/shorts/1dnMj4UqaLk":                  "yt-1dnMj4UqaLk",
		"https://youtube.com/shorts/-nt_bLCNMOc?si=abc":               "yt--nt_bLCNMOc",
		"https://www.youtube.com/watch?v=jNQXAC9IVRw&t=10s":           "yt-jNQXAC9IVRw",
		"https://m.youtube.com/watch?feature=share&v=jNQXAC9IVRw":     "yt-jNQXAC9IVRw",
		"https://youtu.be/jNQXAC9IVRw?si=x":                           "yt-jNQXAC9IVRw",
		"https://www.instagram.com/p/BsOGulcndj-/":                    "ig-BsOGulcndj-",
		"https://www.instagram.com/reel/DWAUs7jANkM/?igsh=abc":        "ig-DWAUs7jANkM",
		"https://www.instagram.com/reels/DWAUs7jANkM/":                "ig-DWAUs7jANkM",
		"https://www.instagram.com/natgeo/p/DWAUs7jANkM/?img_index=1": "ig-DWAUs7jANkM",
	} {
		got, err := parsePostID(in)
		if err != nil || got != want {
			t.Errorf("parsePostID(%q) = %q, %v; want %q", in, got, err, want)
		}
		if !validID(got) {
			t.Errorf("validID(%q) = false", got)
		}
	}
	for _, in := range []string{"https://www.youtube.com/@natgeo", "https://www.instagram.com/natgeo/", "..", "yt-../../x", "ig-a/b"} {
		if id, err := parsePostID(in); err == nil {
			t.Errorf("parsePostID(%q) = %q", in, id)
		}
		if validID(in) {
			t.Errorf("validID(%q) = true", in)
		}
	}
}

func TestOriginalPhoto(t *testing.T) {
	got := originalPhoto("https://pbs.twimg.com/media/FfzVJ1KXkAADnGn.jpg")
	if want := "https://pbs.twimg.com/media/FfzVJ1KXkAADnGn?format=jpg&name=orig"; got != want {
		t.Errorf("got %s, want %s", got, want)
	}
	if e := extOf(got, "image"); e != ".jpg" {
		t.Errorf("extOf = %s", e)
	}
}

// XDLL_NET=1 go test -run Network: downloads real posts.
func TestNetworkAdd(t *testing.T) {
	if os.Getenv("XDLL_NET") == "" {
		t.Skip("set XDLL_NET=1")
	}
	l, err := OpenLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f, _ := l.CreateFolder("Test")
	urls := []string{
		"https://x.com/elonmusk/status/1585341984679469056", // a video
		"https://x.com/NASAMars/status/2108374638355230914", // a photo
		"https://www.youtube.com/shorts/1dnMj4UqaLk",        // a Short, audio apart
		"https://www.instagram.com/reel/DWAUs7jANkM/",       // a reel
		"https://www.instagram.com/p/BsOGulcndj-/",          // a photo
	}
	for _, u := range urls {
		b, err := l.Add(context.Background(), u, f.ID, QualityMax)
		if err != nil {
			t.Fatalf("%s: %v", u, err)
		}
		t.Logf("%s @%s %q: %d media", b.Source, b.Author, b.Text, len(b.Media))
		for _, m := range b.Media {
			st, err := os.Stat(filepath.Join(l.mediaDir(), m.File))
			if err != nil || st.Size() < 10_000 {
				t.Errorf("%s: %v %v", m.File, st, err)
			}
			t.Logf("  %s %s %dx%d %d bytes thumb=%s", m.Kind, m.File, m.Width, m.Height, m.Size, m.Thumb)
		}
		if _, err := l.Add(context.Background(), u, "", QualityMax); err == nil {
			t.Error("added twice")
		}
	}
	l2, err := OpenLibrary(l.root())
	if err != nil || len(l2.List().Bookmarks) != len(urls) || l2.List().Bookmarks[0].FolderID != f.ID {
		t.Fatalf("reload: %+v %v", l2.List(), err)
	}
	s, _ := l2.DeleteFolder(f.ID)
	if s.Bookmarks[0].FolderID != "" {
		t.Error("bookmark kept a deleted folder")
	}
}

// XDLL_NET=1 go test -run Network: economy videos are lighter.
func TestNetworkEconomy(t *testing.T) {
	if os.Getenv("XDLL_NET") == "" {
		t.Skip("set XDLL_NET=1")
	}
	max, err := OpenLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	eco, err := OpenLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{
		"https://x.com/i/status/2108580996761493865", // 1080p
		"https://www.youtube.com/shorts/1dnMj4UqaLk",
	} {
		a, err := max.Add(context.Background(), u, "", QualityMax)
		if err != nil {
			t.Fatalf("%s: %v", u, err)
		}
		b, err := eco.Add(context.Background(), u, "", QualityEconomy)
		if err != nil {
			t.Fatalf("%s: %v", u, err)
		}
		ma, mb := a.Media[0], b.Media[0]
		t.Logf("%s: max %dx%d %d bytes, economy %dx%d %d bytes", u, ma.Width, ma.Height, ma.Size, mb.Width, mb.Height, mb.Size)
		if ma.Light || !mb.Light || mb.Size >= ma.Size {
			t.Errorf("%s: economy not lighter", u)
		}
	}
}

// XDLL_NET=1 go test -run Network: Optimize makes maximum quality videos lighter.
func TestNetworkOptimize(t *testing.T) {
	if os.Getenv("XDLL_NET") == "" {
		t.Skip("set XDLL_NET=1")
	}
	l, err := OpenLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{
		"https://x.com/i/status/2108580996761493865",
		"https://www.youtube.com/shorts/1dnMj4UqaLk",
		"https://www.instagram.com/reel/DWAUs7jANkM/", // left as it is
	} {
		if _, err := l.Add(context.Background(), u, "", QualityMax); err != nil {
			t.Fatalf("%s: %v", u, err)
		}
	}
	before := l.List().Bookmarks
	st, err := l.Optimize(context.Background())
	if err != nil || st.Total != 2 || st.Done != 2 || st.Failed != 0 || st.Saved <= 0 {
		t.Fatalf("Optimize = %+v, %v", st, err)
	}
	for i, b := range l.List().Bookmarks {
		m, old := b.Media[0], before[i].Media[0]
		st, err := os.Stat(filepath.Join(l.mediaDir(), m.File))
		if err != nil || st.Size() != m.Size {
			t.Errorf("%s: %v %v", m.File, st, err)
		}
		t.Logf("%s: %d -> %d bytes, %dx%d", b.ID, old.Size, m.Size, m.Width, m.Height)
		if b.Source != "instagram" && (!m.Light || m.Size >= old.Size) {
			t.Errorf("%s not lighter", b.ID)
		}
		entries, _ := os.ReadDir(filepath.Join(l.mediaDir(), b.ID))
		for _, e := range entries {
			if n := e.Name(); strings.HasPrefix(n, ".") || strings.Contains(n, ".light") || strings.HasSuffix(n, ".part") {
				t.Errorf("left %s/%s", b.ID, n)
			}
		}
	}
	if ids := l.heavyPosts(); len(ids) != 0 {
		t.Errorf("still heavy: %v", ids)
	}
}

// XDLL_NET=1 go test -run Network: exports economy videos in maximum quality.
func TestNetworkExportOriginal(t *testing.T) {
	if os.Getenv("XDLL_NET") == "" {
		t.Skip("set XDLL_NET=1")
	}
	l, err := OpenLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	b, err := l.Add(context.Background(), "https://x.com/i/status/2108580996761493865", "", QualityEconomy)
	if err != nil {
		t.Fatal(err)
	}
	for _, original := range []bool{false, true} {
		dir := t.TempDir()
		r, err := l.exportTo(context.Background(), []string{b.ID}, dir, original)
		if err != nil || r.Files != 1 || r.Light != 0 {
			t.Fatalf("exportTo(%v) = %+v, %v", original, r, err)
		}
		st, err := os.Stat(filepath.Join(dir, b.Author+"_"+b.ID+"_1.mp4"))
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("original=%v: %d bytes, stored %d", original, st.Size(), b.Media[0].Size)
		if (st.Size() > b.Media[0].Size) != original || (r.Downloaded == 1) != original {
			t.Errorf("original=%v: %+v, %d bytes", original, r, st.Size())
		}
	}
}
