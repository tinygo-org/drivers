package rtl8720dn

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"time"

	"tinygo.org/x/drivers/net/http"
)

func (rtl *RTL8720DN) setupHTTPServer() error {
	_, err := rtl.Rpc_lwip_close(-1)
	if err != nil {
		return err
	}

	_, err = rtl.Rpc_lwip_socket(0x00000002, 0x00000001, 0x00000000)
	if err != nil {
		return err
	}

	name := []byte{0x00, 0x02, 0x00, 0x50, 0x00, 0x00, 0x00, 0x00, 0xA5, 0x42, 0x00, 0x00, 0xC7, 0x61, 0x01, 0x00}
	_, err = rtl.Rpc_lwip_bind(0, name, uint32(len(name)))
	if err != nil {
		return err
	}

	_, err = rtl.Rpc_lwip_listen(0, 4)
	if err != nil {
		return err
	}

	_, err = rtl.Rpc_lwip_fcntl(0, 4, 1)
	if err != nil {
		return err
	}

	return nil
}

func (rtl *RTL8720DN) accept() (bool, error) {
	addr := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xE4, 0xD7, 0x00, 0x20, 0x00, 0xD8, 0x00, 0x20}
	length := uint32(len(addr))
	ret, err := rtl.Rpc_lwip_accept(0, addr, &length)
	if err != nil {
		return false, err
	}

	return ret == 1, nil
}

func (rtl *RTL8720DN) handleHTTP() error {
	socket := int32(1)
	optval := []byte{0x01, 0x00, 0x00, 0x00}
	_, err := rtl.Rpc_lwip_setsockopt(socket, 0x00000FFF, 8, optval, uint32(len(optval)))
	if err != nil {
		return nil
	}

	_, err = rtl.Rpc_lwip_setsockopt(socket, 6, 1, optval, uint32(len(optval)))
	if err != nil {
		return nil
	}

	buf := make([]byte, 4096)
	for {
		_, err = rtl.Rpc_lwip_recv(socket, &buf, uint32(len(buf)), 8, 0)
		if err != nil {
			return nil
		}
		if len(buf) > 0 {
			break
		}

		_, err = rtl.Rpc_lwip_errno()
		if err != nil {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	buf2 := make([]byte, 4096)
	result, err := rtl.Rpc_lwip_recv(socket, &buf2, uint32(len(buf2)), 8, 0)
	if err != nil {
		return nil
	}
	if result != -1 && result != 0 {
		return fmt.Errorf("Rpc_lwip_recv error")
	}

	result, err = rtl.Rpc_lwip_errno()
	if err != nil {
		return nil
	}
	if result != 11 {
		return fmt.Errorf("Rpc_lwip_errno error")
	}

	b := bufio.NewReader(bytes.NewReader(buf))
	req, err := http.ReadRequest(b)
	if err != nil {
		return err
	}
	if rtl.debug {
		fmt.Printf("%s %s %s\r\n", req.Method, req.RequestURI, req.Proto)
	}

	pos := bytes.Index(buf, []byte("\r\n\r\n"))
	if pos > 0 {
		body := bytes.NewReader(buf[pos+4:])
		req.Body = io.NopCloser(body)
	}

	handler, _ := http.DefaultServeMux.Handler(req)
	rwx := responseWriter{
		header:     http.Header{},
		statusCode: 200,
	}
	rwx.header.Add(`Content-Type`, `text/html; charset=UTF-8`)
	rwx.header.Add(`Connection`, `close`)
	handler.ServeHTTP(&rwx, req)
	rwx.header.Add(`Content-Length`, fmt.Sprintf("%d", len(rwx.Buf)))

	optval = []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xA5, 0xA5, 0xA5, 0xA5}
	_, err = rtl.Rpc_lwip_setsockopt(socket, 0x00000FFF, 0x1006, optval, uint32(len(optval)))
	if err != nil {
		return nil
	}

	optval = []byte{0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xA5, 0xA5, 0xA5, 0xA5}
	_, err = rtl.Rpc_lwip_setsockopt(socket, 0x00000FFF, 0x1005, optval, uint32(len(optval)))
	if err != nil {
		return nil
	}

	maxfdp1 := int32(2)
	writeset := []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	timeout := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x42, 0x0F, 0x00, 0x0A, 0x00, 0x00, 0x00}
	_, err = rtl.Rpc_lwip_select(maxfdp1, []byte{}, writeset, []byte{}, timeout)
	if err != nil {
		return nil
	}

	msg := rwx.Buf
	hb := bytes.Buffer{}
	err = rwx.header.Write(&hb)
	if err != nil {
		return err
	}

	data := []byte(fmt.Sprintf("HTTP/1.1 %d OK\n", rwx.statusCode))
	data = append(data, hb.Bytes()...)
	data = append(data, byte('\n'))

	_, err = rtl.Rpc_lwip_send(socket, data, 8)
	if err != nil {
		return nil
	}

	if len(msg) > 0 {
		timeout = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x42, 0x0F, 0x00, 0x54, 0x00, 0x00, 0x00}
		_, err = rtl.Rpc_lwip_select(maxfdp1, []byte{}, writeset, []byte{}, timeout)
		if err != nil {
			return nil
		}

		_, err = rtl.Rpc_lwip_send(socket, msg, 8)
		if err != nil {
			return nil
		}
	}

	for i := 0; i < 4; i++ {
		buf := make([]byte, 4096)
		_, err = rtl.Rpc_lwip_recv(socket, &buf, uint32(len(buf)), 8, 0)
		if err != nil {
			return nil
		}

		_, err = rtl.Rpc_lwip_errno()
		if err != nil {
			return nil
		}
	}

	_, err = rtl.Rpc_lwip_close(socket)
	if err != nil {
		return nil
	}
	return nil
}

func (rtl *RTL8720DN) ListenAndServe(addr string, handler http.Handler) error {
	err := rtl.setupHTTPServer()
	if err != nil {
		return err
	}

	for {
		connected, err := rtl.accept()
		if err != nil {
			return err
		}

		if connected {
			err := rtl.handleHTTP()
			if err != nil {
				return err
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

type responseWriter struct {
	Buf        []byte
	header     http.Header
	statusCode int
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) Write(b []byte) (int, error) {
	r.Buf = append(r.Buf, b...)
	return len(b), nil
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}
