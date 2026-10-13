//go:build windows

package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/egoist/mygo"
)

func TestMain(m *testing.M) {
	if os.Getenv("XDLL_CLIP") == "" {
		os.Exit(m.Run())
	}
	code := 1
	mygo.App.WhenReady(func() {
		go func() {
			code = m.Run()
			mygo.App.Quit()
		}()
	})
	mygo.App.Run()
	os.Exit(code)
}

func powershell(t *testing.T, script string) string {
	out, err := exec.Command("powershell", "-NoProfile", "-STA", "-Command", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestClipboardForClaude(t *testing.T) {
	if os.Getenv("XDLL_CLIP") == "" {
		t.Skip("set XDLL_CLIP=1")
	}
	l, err := OpenLibrary(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{
		"https://x.com/NASAMars/status/2108374638355230914",
		"https://x.com/elonmusk/status/1585341984679469056",
	} {
		b, err := l.Add(context.Background(), u, "", QualityMax)
		if err != nil {
			t.Fatal(err)
		}
		copied, err := l.CopyForClaude(b.ID, 0, "all")
		if err != nil || copied != "all" {
			t.Fatalf("CopyForClaude: %v %v", copied, err)
		}
		if err := mygo.Clipboard.Flush(); err != nil {
			t.Fatal(err)
		}
		text := powershell(t, "Get-Clipboard -Raw")
		size := powershell(t, "Add-Type -AssemblyName System.Windows.Forms; $i = [System.Windows.Forms.Clipboard]::GetImage(); if ($i) { \"$($i.Width)x$($i.Height)\" } else { 'none' }")
		t.Logf("%s\nimage: %s\n---\n%s", u, size, text)
		if !strings.Contains(text, b.URL) || !strings.Contains(text, b.Media[0].File[len(b.ID)+1:]) || size == "none" {
			t.Errorf("clipboard is missing the text or the image")
		}
	}
}
