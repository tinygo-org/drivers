package main

import (
	"bytes"
	"fmt"
	"log"
	"machine"
	"net"
	"time"

	"tinygo.org/x/drivers/netdev"
	"tinygo.org/x/drivers/espat"
)

var (
	ssid string
	pass string
	addr string = "10.0.0.100:8080"
)

var (
	netcfg = espat.Config{
		Ssid:       ssid,
		Passphrase: pass,
		Uart:       machine.UART2,
		Tx:         machine.TX1,
		Rx:         machine.RX0,
	}

	dev = espat.New(&netcfg)
)

var buf = &bytes.Buffer{}

func main() {

	waitSerial()

	netdev.Use(dev)
	if err := dev.NetConnect(); err != nil {
		log.Fatal(err)
	}

	for {
		sendBatch()
		time.Sleep(500 * time.Millisecond)
	}
}

func sendBatch() {

	// make TCP connection
	message("---------------\r\nDialing TCP connection")
	conn, err := net.Dial("tcp", addr)
	for ; err != nil; conn, err = net.Dial("tcp", addr) {
		message(err.Error())
		time.Sleep(5 * time.Second)
	}

	n := 0
	w := 0
	start := time.Now()

	// send data
	message("Sending data")

	for i := 0; i < 1000; i++ {
		buf.Reset()
		fmt.Fprint(buf,
			"\r---------------------------- i == ", i, " ----------------------------"+
				"\r---------------------------- i == ", i, " ----------------------------")
		if w, err = conn.Write(buf.Bytes()); err != nil {
			println("error:", err.Error(), "\r")
			break
		}
		n += w
	}

	buf.Reset()
	ms := time.Now().Sub(start).Milliseconds()
	fmt.Fprint(buf, "\nWrote ", n, " bytes in ", ms, " ms\r\n")
	message(buf.String())

	if _, err := conn.Write(buf.Bytes()); err != nil {
		println("error:", err.Error(), "\r")
	}

	println("Disconnecting TCP...")
	conn.Close()
}

func message(msg string) {
	println(msg, "\r")
}

// Wait for user to open serial console
func waitSerial() {
	for !machine.Serial.DTR() {
		time.Sleep(100 * time.Millisecond)
	}
}
