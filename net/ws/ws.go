// Package is mostly inspired by https://github.com/gopherjs/websocket/blob/master/websocketjs/websocketjs.go
// Websocket RFC: https://tools.ietf.org/html/rfc6455
package ws

import (
	"errors"
	"syscall/js"
	"time"

	"github.com/Nerzal/drivers/net"
)

// Close codes defined in RFC 6455, section 11.7.
const (
	// 1000 indicates a normal closure, meaning that the purpose for
	// which the connection was established has been fulfilled.
	closeNormalClosure = 1000
)

type Websocket struct {
	socket   js.Value
	url      string
	messages chan []byte
}

var ErrCouldNotGetWebsocket = errors.New("could not get websocket")

func New(url string) (*Websocket, error) {
	ws := js.Global().Get("WebSocket")
	if ws.IsNull() {
		return nil, ErrCouldNotGetWebsocket
	}

	return &Websocket{
		socket:   ws,
		url:      url,
		messages: make(chan []byte),
	}, nil
}

func (w *Websocket) onMessage(this js.Value, args []js.Value) {
	go func() {
		bytes := []byte(args[0].String())
		w.messages <- bytes
	}()
}

func (w *Websocket) Open() {
	w.socket.New(w.url)
	w.socket.Call("addEventListener", "open", w.OnOpen)
	w.socket.Call("addEventListener", "close", w.OnClose)
	w.socket.Call("addEventListener", "error", w.OnError)
}

func (w *Websocket) OnOpen(this js.Value, args []js.Value) {
	println("Connection successfully established")
}

func (w *Websocket) OnError(this js.Value, args []js.Value) {
	println("Connection has been unexpectedly terminated")
}

func (w *Websocket) OnClose(this js.Value, args []js.Value) {
	println("Connection has been closed")
}

func (w *Websocket) AddEventListener(name string, eventListener func(this js.Value, args []js.Value)) {
	w.socket.Call("addEventListener", name, eventListener)
}

func (w *Websocket) RemoveEventListener(name string, eventListener func(this js.Value, args []js.Value)) {
	w.socket.Call("removeEventListener", name, eventListener)
}

func (w *Websocket) Send(data interface{}) {
	w.socket.Call("send", data)
}

func (w *Websocket) Close() error {
	w.socket.Call("close", closeNormalClosure)
	return nil
}

func (w *Websocket) Read(b []byte) (n int, err error) {
	message := <-w.messages

	for i := range b {
		if i >= len(message) {
			return i, nil
		}

		b[i] = message[i]
	}

	return 0, nil
}
func (w *Websocket) Write(b []byte) (n int, err error) {
	w.Send(b)
	return len(b), nil
}

func (w *Websocket) LocalAddr() net.Addr {
	return nil
}

func (w *Websocket) RemoteAddr() net.Addr {
	return nil
}

func (w *Websocket) SetDeadline(t time.Time) error {
	return nil
}

func (w *Websocket) SetReadDeadline(t time.Time) error {
	return nil
}

func (w *Websocket) SetWriteDeadline(t time.Time) error {
	return nil
}
