package main

import (
	"io"
	"log"
	"net"
)

func main() {
	// Listen for connections
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer l.Close()
	println("Listening on port", ":8080")
	for {
		// Wait for a connection
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		println("Accepted connection from", conn.RemoteAddr().String())
		// Service the new connection in a goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently
		go func(c net.Conn) {
			// Echo all incoming data
			io.Copy(c, c)
			// Shut down the connection
			c.Close()
		}(conn)
	}
}
