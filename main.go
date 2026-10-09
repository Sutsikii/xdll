// xdll downloads the photos and videos of Twitter/X posts and keeps them in
// a library of bookmarks, sorted into folders.
//
//	go tool mygo dev     # run with live reload
//	go tool mygo build   # build/windows-amd64/xdll.exe
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/egoist/mygo"
)

//go:embed all:web
var web embed.FS

func main() {
	app := mygo.App
	// The name decides the user data directory, where the library lives.
	app.SetName("xdll")

	dir, err := app.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}
	lib, err := OpenLibrary(dir)
	if err != nil {
		log.Fatal(err)
	}
	mygo.Bind(lib)

	// The page and the downloaded media share one origin: /media/… are the
	// files of the library, the rest is the page.
	page, _ := fs.Sub(web, "web")
	pages := mygo.FileServer(page)
	media := http.StripPrefix("/media/", mygo.FileServer(os.DirFS(lib.mediaDir())))
	mygo.Protocol.HandleFunc("mygo", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/media/") {
			media.ServeHTTP(w, r)
			return
		}
		pages.ServeHTTP(w, r)
	})

	app.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:           "xdll",
			Width:           1180,
			Height:          760,
			MinWidth:        720,
			MinHeight:       480,
			StateKey:        "main",
			BackgroundColor: "light-dark(#f6f6f4, #141416)",
			URL:             "/",
		})
	})
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
