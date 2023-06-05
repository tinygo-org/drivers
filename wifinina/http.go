package wifinina

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/drivers/net/http"
)

func (d *Device) ListenAndServe(addr string, handler http.Handler) error {

	if handler == nil {
		handler = http.DefaultServeMux
	}

	server := newServer(d, handler)

	if err := server.listen(addr); err != nil {
		return err
	}

	for {
		client, err := server.accept()
		if err != nil {
			return err
		}

		if err := client.handleHTTP(); err != nil {
			return err
		}

		if err = client.stop(); err != nil {
			return err
		}
	}

	return nil
}

// Server stuff

type server struct {
	device  *Device
	handler http.Handler
	sock    uint8
	clients map[uint8]*client // keyed by client sock
}

func newServer(device *Device, handler http.Handler) *server {
	return &server{
		device:  device,
		handler: handler,
		sock:    NoSocketAvail,
		clients: make(map[uint8]*client),
	}
}

func portFromAddr(addr string) (uint16, error) {
	// ignore anything before ':' in address
	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return 0, fmt.Errorf("Missing ':' in address")
	}
	v, err := strconv.ParseUint(addr[i+1:], 10, 16)
	if err != nil {
		return 0, fmt.Errorf("Parsing address err: %s", err)
	}
	return uint16(v), nil
}

func (s *server) listen(addr string) error {
	port, err := portFromAddr(addr)
	if err != nil {
		return fmt.Errorf("Getting port err: %s", err)
	}

	s.sock, err = s.device.GetSocket()
	if err != nil {
		return fmt.Errorf("Getting socket err: %s", err)
	}
	if s.sock == NoSocketAvail {
		return fmt.Errorf("No socket available")
	}

	return s.device.StartServer(port, s.sock, ProtoModeTCP)
}

func (s *server) availServer(sock uint8) (uint8, error) {
	d := s.device

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return NoSocketAvail, fmt.Errorf("Wait for CS: %s", err)
	}

	l := d.sendCmd(CmdAvailDataTCP, 1)
	l += d.sendParam8(sock, true)
	d.addPadding(l)
	d.spiChipDeselect()
	_, err := d.waitRspCmd1(CmdAvailDataTCP)
	if err != nil {
		return NoSocketAvail, fmt.Errorf("Wait for Rsp: %s", err)
	}
	newsock, err := d.getUint16(2, err)
	if err != nil {
		return NoSocketAvail, fmt.Errorf("getUint16: %s", err)
	}
	return uint8(newsock >> 8), nil
}

func (s *server) accept() (*client, error) {

	for {
		sock, err := s.availServer(s.sock)
		if err != nil {
			return nil, fmt.Errorf("accept: %w", err)
		}

		if sock == NoSocketAvail {
			continue
		}

		if client, ok := s.clients[sock]; ok {
			return client, nil
		}

		client := newClient(s, sock)
		s.clients[sock] = client

		return client, nil
	}
}

// client stuff

type client struct {
	server *server
	device *Device
	sock   uint8

	// HTTP request
	req     *http.Request
	reqBuf  bytes.Buffer
	readBuf [256]byte

	// HTTP response
	res        bytes.Buffer
	resHdr     http.Header
	resBuf     bytes.Buffer
	statusCode int
}

func newClient(server *server, sock uint8) *client {
	return &client{
		server: server,
		device: server.device,
		sock:   sock,
	}
}

// client implements http.ResponseWriter interface

func (c *client) Header() http.Header {
	return c.resHdr
}

func (c *client) Write(b []byte) (int, error) {
	return c.resBuf.Write(b)
}

func (c *client) WriteHeader(statusCode int) {
	c.statusCode = statusCode
}

func (c *client) status() uint8 {
	d := c.device

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.waitForChipSelect(); err != nil {
		d.spiChipDeselect()
		return 0
	}

	l := d.sendCmd(CmdGetClientStateTCP, 1)
	l += d.sendParam8(c.sock, true)
	d.addPadding(l)
	d.spiChipDeselect()
	_, err := d.waitRspCmd1(CmdGetClientStateTCP)
	if err != nil {
		return 0
	}
	status, err := d.getUint8(1, err)
	if err != nil {
		return 0
	}
	return status
}

func (c *client) stop() error {
	if err := c.device.StopClient(c.sock); err != nil {
		return err
	}

	// Wait max 5 secs for the connection to close
	for i := 0; i < 50 && c.status() != uint8(TCPStateClosed); i++ {
		time.Sleep(100 * time.Millisecond)
	}

	if c.status() != uint8(TCPStateClosed) {
		return fmt.Errorf("stop failed, client status %x", c.status())
	}

	return nil
}

func (c *client) handleHTTP() error {

	c.reqBuf.Reset()
	end := -1

	// read the request

	start := time.Now()
	for {

		// TODO use Server.ReadTimeout
		if time.Since(start) > 1*time.Second {
			return fmt.Errorf("ReadTimeout")
		}

		n, err := c.device.GetDataBuf(c.sock, c.readBuf[:])
		if err != nil {
			return fmt.Errorf("GetDataBuf: %s", err)
		}
		if n == 0 {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		c.reqBuf.Write(c.readBuf[:n])
		bytesSoFar := c.reqBuf.Bytes()

		if end == -1 {

			// search for blank line marking end-of-header
			end = bytes.Index(bytesSoFar, []byte("\r\n\r\n"))
			if end == -1 {
				continue
			}

			// found end-of-header; parse header
			end += len([]byte("\r\n\r\n"))
			bufio := bufio.NewReader(bytes.NewReader(bytesSoFar[:end]))
			c.req, err = http.ReadRequest(bufio)
			if err != nil {
				return err
			}
		}

		v := c.req.Header.Get("Content-Length")
		if v == "" {
			// no body; we're done reading request
			break
		}

		length, _ := strconv.Atoi(v)
		if end+length == len(bytesSoFar) {
			// got the whole body
			body := bytes.NewReader(bytesSoFar[end:])
			c.req.Body = io.NopCloser(body)
			break
		}

		// continue reading request...
	}

	// build the response

	c.statusCode = 200

	c.resHdr = http.Header{}
	c.resHdr.Add(`Content-Type`, `text/html; charset=UTF-8`)
	c.resHdr.Add(`Connection`, `close`)

	c.resBuf.Reset()
	c.server.handler.ServeHTTP(c, c.req)

	c.resHdr.Add(`Content-Length`, fmt.Sprintf("%d", c.resBuf.Len()))

	c.res.Reset()
	fmt.Fprintf(&c.res, "HTTP/1.1 %d %s\r\n", c.statusCode,
		http.StatusText(c.statusCode))
	if err := c.resHdr.Write(&c.res); err != nil {
		return err
	}
	c.res.WriteByte(byte('\n'))
	c.res.Write(c.resBuf.Bytes())

	// send the response

	written, err := c.device.SendData(c.res.Bytes(), c.sock)
	if err != nil {
		return err
	}
	if written == 0 {
		return ErrDataNotWritten
	}
	if sent, _ := c.device.CheckDataSent(c.sock); !sent {
		return ErrCheckDataError
	}

	return nil
}
