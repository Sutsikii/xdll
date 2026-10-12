// xdll downloads the media of Twitter/X posts into a library.
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/updater"
)

//go:embed all:web
var web embed.FS

func main() {
	app := mygo.App
	app.SetName("xdll")
	mygo.Use(updater.Plugin)

	dir, err := app.Path(mygo.PathUserData)
	if err != nil {
		log.Fatal(err)
	}
	lib, err := OpenLibrary(dir)
	if err != nil {
		log.Fatal(err)
	}
	mygo.Bind(lib)

	// /media/ serves the library files, everything else the page.
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
