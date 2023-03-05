// This example is an HTTP server serving up a static file system
//
// Note: It may be necessary to increase the stack size when using "net/http".
// Use the -stack-size=4KB command line option.

package main

import (
	"embed"
	"log"
	"net/http"
	"time"
)

var (
	ssid string
	pass string
	port string = ":80"
)

//go:embed index.html main.go images
var fs embed.FS

func main() {
	// wait a bit for console
	time.Sleep(2 * time.Second)

	if err := netdev.NetConnect(); err != nil {
		log.Fatal(err)
	}

	hfs := http.FileServer(http.FS(fs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hfs.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(port, nil))
}
