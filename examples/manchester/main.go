package main

import (
	"errors"
	"time"

	"machine"

	"tinygo.org/x/drivers/manchester"
	"tinygo.org/x/drivers/syn115"
	"tinygo.org/x/drivers/syn480r"
)

var (
	txPin = machine.GP4
	rxPin = machine.GP5
)

func main() {
	tx := syn115.New(txPin)
	tx.Configure()
	rx := syn480r.New(rxPin)
	rx.Configure()

	enc := manchester.NewEncoder(tx, manchester.Config{BitRate: 2000})
	dec := manchester.NewDecoder(rx, manchester.Config{
		BitRate: 2000,
		Timeout: 500 * time.Millisecond,
	})

	buf := make([]byte, 8)
	for i := 0; ; i++ {
		msg := []byte{'t', 'i', 'n', 'y', byte('0' + i%10)}
		if _, err := enc.Write(msg); err != nil {
			println("write:", err.Error())
		}
		time.Sleep(100 * time.Millisecond)

		n, err := dec.Read(buf)
		switch {
		case err == nil:
			println("received:", string(buf[:n]))
		case errors.Is(err, manchester.ErrTimeout):
			println("no frame received")
		default:
			println("read:", err.Error())
		}
	}
}
