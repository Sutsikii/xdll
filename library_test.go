package main

import (
	"context"
	"os"
	"path/filepath"
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
		b, err := l.Add(context.Background(), u, f.ID)
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
		if _, err := l.Add(context.Background(), u, ""); err == nil {
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
