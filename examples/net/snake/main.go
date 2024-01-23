// This example is the classic snake network test.  The snake is feed a steady
// diet of pkts and the pkts work themselves thru the snake segments and exit
// the tail.  Each snake segment is a TCP socket connection to a server.  The
// server echos pkts received back to the snake, and serves each segment on a
// different port.  (See server/main.go for server).
//
//                snake     |    server
//                          |
//            head ----->---|-->--+
//             seg a        |     |
//                    +---<-|--<--+
//                    |     |
//                    +-->--|-->--+
//             seg b        |     |
//                    +---<-|--<--+
//                    |     |
//                    +-->--|-->--+
//             seg c        |     |
//                    +---<-|--<--+
//                    |     |
//                    +-->--|-->--+
//              ...         |     |
//                    +---<-|--<--+
//                    |     |
//                    +-->--|-->--+
//             seg n        |     |
//            tail -------<-|--<--+
//                          |

// The snake segments are linked by channels and each segment is run as a go
// func.  This forces segments to connect and run concurrently, which is a good
// test of the underlying driver's ability to handle concurrent connections.

//go:build ninafw || wioterminal

package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var (
	server string = "10.0.0.100:8080"
)

// The Snake, AI-generated poem
var poem = `In the swamp's green, murky depths,
A serpent slithers, silent, wreathed.
Its eyes gleam yellow, cold, and bright,
As it stalks its prey with all its might.
A frog leaps high, a tasty treat,
The snake strikes fast, its fangs meet.
The frog's last croak is all that's heard,
As the snake swallows it, whole and curbed.
Digested, the snake rests content,
In the swamp's embrace, its coils unbent.
A silent hunter, deadly, sleek,
The snake reigns supreme, its prey its peak.`

func segment(in chan []byte, out chan []byte) {
	var buf [512]byte
	for {
		c, err := net.Dial("tcp", server)
		for ; err != nil; c, err = net.Dial("tcp", server) {
			println(err.Error())
			time.Sleep(5 * time.Second)
		}
		for {
			select {
			case msg := <-in:
				_, err := c.Write(msg)
				if err != nil {
					log.Fatal(err.Error())
				}
				time.Sleep(100 * time.Millisecond)
				n, err := c.Read(buf[:])
				if err != nil {
					log.Fatal(err.Error())
				}
				out <- buf[:n]
			}
		}
	}
}

func feedit(head chan []byte) {
	for i := 0; i < 100; i++ {
		head <- []byte(fmt.Sprintf("\n---%d---", i))
		for _, line := range strings.Split(poem, "\n") {
			head <- []byte(line)
		}
	}
}

var head = make(chan []byte)
var a = make(chan []byte)
var b = make(chan []byte)
var c = make(chan []byte)
var d = make(chan []byte)
var e = make(chan []byte)
var f = make(chan []byte)
var tail = make(chan []byte)

func main() {

	// The snake
	go segment(head, a)
	go segment(a, b)
	go segment(b, c)
	go segment(c, d)
	go segment(d, e)
	go segment(e, f)
	go segment(f, tail)

	go feedit(head)

	for {
		select {
		case msg := <-tail:
			println(string(msg))
		}
	}
}
