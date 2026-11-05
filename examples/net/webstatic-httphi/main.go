// This example is the webstatic example using the heapless httphi router
// from github.com/soypat/lneto instead of "net/http".

//go:build ninafw || wioterminal

package main

import (
	"embed"
	"log"
	"net"
	"path"
	"time"

	"github.com/soypat/lneto/http/httphi"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	port string = ":80"
)

const (
	numWorkers    = 2
	memoryPerConn = 1024
)

//go:embed index.html main.go images
var fs embed.FS

func main() {
	// wait a bit for console
	time.Sleep(2 * time.Second)

	link, _ := probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	var mux httphi.MuxSlice
	mux.Handle("GET /", serveFile)

	var router httphi.Router
	cfg := httphi.DefaultRouterConfig(numWorkers, memoryPerConn, mux.MaxPathValues())
	err = router.Configure(&mux, cfg)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			println("accept:", err.Error())
			time.Sleep(time.Second)
			continue
		}
		if err := router.Handle(conn); err != nil {
			conn.Close()
		}
	}
}

func serveFile(ex *httphi.Exchange) {
	name := string(ex.RequestPath()[1:])
	if name == "" {
		name = "index.html"
	}

	f, err := fs.Open(name)
	if err != nil {
		ex.RespondString(httphi.StatusNotFound, "text/plain", "404 page not found")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.IsDir() {
		ex.RespondString(httphi.StatusNotFound, "text/plain", "404 page not found")
		return
	}

	ex.StageHeader("Content-Type", contentType(name))
	ex.StageHeaderInt("Content-Length", info.Size())
	ex.StageHeader("Connection", "close")

	var buf [512]byte
	for {
		n, err := f.Read(buf[:])
		if n > 0 {
			if _, werr := ex.WriteBody(buf[:n]); werr != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

func contentType(name string) string {
	switch path.Ext(name) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".png":
		return "image/png"
	default:
		return "text/plain; charset=utf-8"
	}
}
