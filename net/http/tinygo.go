package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"tinygo.org/x/drivers/net"
	"tinygo.org/x/drivers/net/tls"
)

var buf []byte

func SetBuf(b []byte) {
	buf = b
}

func (c *Client) Do(req *Request) (*Response, error) {
	if c.Jar != nil {
		for _, cookie := range c.Jar.Cookies(req.URL) {
			req.AddCookie(cookie)
		}
	}

	transport := c.Transport
	if transport == nil {
		transport = DefaultTransport
	}
	res, err := transport.RoundTrip(req)

	if c.Jar != nil {
		if rc := res.Cookies(); len(rc) > 0 {
			c.Jar.SetCookies(req.URL, rc)
		}
	}

	return res, err
}

type Transport struct {
}

var DefaultTransport RoundTripper

func init() {
	DefaultTransport = &Transport{}
}

func (t *Transport) RoundTrip(req *Request) (*Response, error) {
	switch req.URL.Scheme {
	case "http":
		return t.doHTTP(req)
	case "https":
		return t.doHTTPS(req)
	default:
		return nil, fmt.Errorf("invalid schemer : %s", req.URL.Scheme)
	}
}

func (t *Transport) doHTTP(req *Request) (*Response, error) {
	// make TCP connection
	ip := net.ParseIP(req.URL.Hostname())
	port := 80
	if req.URL.Port() != "" {
		p, err := strconv.ParseUint(req.URL.Port(), 0, 64)
		if err != nil {
			return nil, err
		}
		port = int(p)
	}
	raddr := &net.TCPAddr{IP: ip, Port: port}
	laddr := &net.TCPAddr{Port: 8080}

	conn, err := net.DialTCP("tcp", laddr, raddr)
	retry := 0
	for ; err != nil; conn, err = net.DialTCP("tcp", laddr, raddr) {
		retry++
		if retry > 10 {
			return nil, fmt.Errorf("Connection failed: %s", err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	p := req.URL.Path
	if p == "" {
		p = "/"
	}
	if req.URL.RawQuery != "" {
		p += "?" + req.URL.RawQuery
	}
	fmt.Fprintln(conn, req.Method+" "+p+" HTTP/1.1")
	fmt.Fprintln(conn, "Host:", req.URL.Host)

	if req.Header.get(`User-Agent`) == "" {
		fmt.Fprintln(conn, "User-Agent: TinyGo")
	}

	for k, v := range req.Header {
		if v == nil || len(v) == 0 {
			return nil, fmt.Errorf("req.Header error: %s", k)
		}
		fmt.Fprintln(conn, k+": "+v[0])
	}

	if req.Header.get(`Connection`) == "" {
		fmt.Fprintln(conn, "Connection: close")
	}

	if req.ContentLength > 0 {
		fmt.Fprintf(conn, "Content-Length: %d\n", req.ContentLength)
	}

	fmt.Fprintln(conn)

	if req.ContentLength > 0 {
		b, err := req.GetBody()
		if err != nil {
			return nil, err
		}

		n, err := b.Read(buf)
		if err != nil {
			return nil, err
		}
		conn.Write(buf[:n])

		b.Close()

	}

	return t.doResp(conn, req)
}

func (t *Transport) doHTTPS(req *Request) (*Response, error) {
	conn, err := tls.Dial("tcp", req.URL.Host, nil)
	retry := 0
	for ; err != nil; conn, err = tls.Dial("tcp", req.URL.Host, nil) {
		retry++
		if retry > 10 {
			return nil, fmt.Errorf("Connection failed: %s", err.Error())
		}
		time.Sleep(1 * time.Second)
	}

	p := req.URL.Path
	if p == "" {
		p = "/"
	}
	if req.URL.RawQuery != "" {
		p += "?" + req.URL.RawQuery
	}
	fmt.Fprintln(conn, req.Method+" "+p+" HTTP/1.1")
	fmt.Fprintln(conn, "Host:", req.URL.Host)

	if req.Header.get(`User-Agent`) == "" {
		fmt.Fprintln(conn, "User-Agent: TinyGo")
	}

	for k, v := range req.Header {
		if v == nil || len(v) == 0 {
			return nil, fmt.Errorf("req.Header error: %s", k)
		}
		fmt.Fprintln(conn, k+": "+v[0])
	}

	if req.Header.get(`Connection`) == "" {
		fmt.Fprintln(conn, "Connection: close")
	}

	if req.ContentLength > 0 {
		fmt.Fprintf(conn, "Content-Length: %d\n", req.ContentLength)
	}

	fmt.Fprintln(conn)

	if req.ContentLength > 0 {
		b, err := req.GetBody()
		if err != nil {
			return nil, err
		}

		n, err := b.Read(buf)
		if err != nil {
			return nil, err
		}
		conn.Write(buf[:n])

		b.Close()

	}

	return t.doResp(conn, req)
}

func (t *Transport) doResp(conn net.Conn, req *Request) (*Response, error) {
	resp := &Response{
		Header: map[string][]string{},
	}

	br := bufio.NewReader(conn)
	tp := textproto.NewReader(br)

	for {
		line, err := tp.ReadLine()
		if err != nil {
			if err == io.ErrNoProgress {
				// default: no timeout
				continue
			}
			conn.Close()
			return nil, err
		}

		status := strings.SplitN(line, " ", 2)
		if len(status) != 2 {
			conn.Close()
			return nil, fmt.Errorf("invalid status : %q", line)
		}
		resp.Proto = status[0]
		fmt.Sscanf(status[0], "HTTP/%d.%d", &resp.ProtoMajor, &resp.ProtoMinor)

		resp.Status = status[1]
		fmt.Sscanf(status[1], "%d", &resp.StatusCode)
		break
	}

	m, err := tp.ReadMIMEHeader()
	if err != nil {
		conn.Close()
		return nil, err
	}
	for k, v := range m {
		//fmt.Printf("%s: %s\n", k, v)

		if strings.ToLower(k) == "content-length" {
			resp.ContentLength, err = strconv.ParseInt(v[0], 10, 64)
			if err != nil {
				conn.Close()
				return nil, err
			}
		}

		if resp.Header.Get(k) == "" {
			resp.Header.Set(k, v[0])
			v = v[1:]
		}
		for _, vv := range v {
			resp.Header.Add(k, vv)
		}

	}

	if resp.Header.Get("Transfer-Encoding") == "chunked" {
		// chunked
		cur := 0
		end := 0
		for {
			length := 0
			if len(buf) < cur+6 {
				// This is not a very accurate check, but in many cases it should be fine.
				return nil, fmt.Errorf("slice out of range : use http.SetBuf() to change the allocation to %d bytes or more", cur+6)
			}
			for i := 0; ; i++ {
				buf[cur+i], err = br.ReadByte()
				if err != nil {
					conn.Close()
					return nil, err
				}
				length = i + 1
				if i > 1 && buf[cur+i-1] == '\r' && buf[cur+i] == '\n' {
					break
				}
			}
			//fmt.Printf("cur:%d length:%d\n", cur, length)

			size, err := strconv.ParseInt(string(buf[cur:cur+length-2]), 16, 64)
			if err != nil {
				conn.Close()
				return nil, err
			}
			//cur += length
			//fmt.Printf("cur:%d length:%d size:%d\n", cur, length, size)

			end = cur + int(size) + 2 // size + 2 (\r\n)
			if len(buf) < end {
				return nil, fmt.Errorf("slice out of range : use http.SetBuf() to change the allocation to %d bytes or more", end)
			}
			for i := 0; i < int(size)+2; i++ {
				buf[cur+i], err = br.ReadByte()
				if err != nil {
					conn.Close()
					return nil, err
				}
			}
			cur += int(size)

			if size == 0 {
				end = end - 2
				break
			}
		}
		//fmt.Printf("%q\n", buf[:end])
		resp.Body = io.NopCloser(bytes.NewReader(buf[:end]))
	} else {
		end := int(resp.ContentLength)
		if len(buf) < end {
			return nil, fmt.Errorf("slice out of range : use http.SetBuf() to change the allocation to %d bytes or more", end)
		}
		for i := 0; i < end; i++ {
			buf[i], err = br.ReadByte()
			if err != nil {
				conn.Close()
				return nil, err
			}
		}
		resp.Body = io.NopCloser(bytes.NewReader(buf[:end]))
	}

	return resp, conn.Close()
}
