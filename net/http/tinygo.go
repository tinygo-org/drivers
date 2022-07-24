package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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
	switch req.URL.Scheme {
	case "http":
		return c.doHTTP(req)
	case "https":
		return c.doHTTPS(req)
	default:
		return nil, fmt.Errorf("invalid schemer : %s", req.URL.Scheme)
	}
}

func (c *Client) doHTTP(req *Request) (*Response, error) {
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

	return c.doResp(conn, req)
}

func (c *Client) doHTTPS(req *Request) (*Response, error) {
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

	return c.doResp(conn, req)
}

func (c *Client) doResp(conn net.Conn, req *Request) (*Response, error) {
	resp := &Response{
		Header: map[string][]string{},
	}

	// Header
	var scanner *bufio.Scanner
	cont := true
	ofs := 0
	remain := int64(0)
	for cont {
		for n, err := conn.Read(buf[ofs:]); n > 0; n, err = conn.Read(buf[ofs:]) {
			if err != nil {
				println("Read error: " + err.Error())
			} else {
				// Take care of the case where "\r\n\r\n" is on the boundary of a buffer
				start := ofs
				if start > 3 {
					start -= 3
				}
				idx := bytes.Index(buf[start:ofs+n], []byte("\r\n\r\n"))
				if idx == -1 {
					ofs += n
					continue
				}
				idx += start + 4

				scanner = bufio.NewScanner(bytes.NewReader(buf[0 : ofs+n]))
				if resp.Status == "" && scanner.Scan() {
					status := strings.SplitN(scanner.Text(), " ", 2)
					if len(status) != 2 {
						conn.Close()
						return nil, fmt.Errorf("invalid status : %q", scanner.Text())
					}
					resp.Proto = status[0]
					fmt.Sscanf(status[0], "HTTP/%d.%d", &resp.ProtoMajor, &resp.ProtoMinor)

					resp.Status = status[1]
					fmt.Sscanf(status[1], "%d", &resp.StatusCode)
				}

				for scanner.Scan() {
					text := scanner.Text()
					if text == "" {
						// end of header
						if idx < n+ofs {
							ofs = ofs + n - idx
							for i := 0; i < ofs; i++ {
								buf[i] = buf[i+idx]
							}
						} else {
							ofs = 0
						}
						break
					} else {
						header := strings.SplitN(text, ": ", 2)
						if len(header) != 2 {
							conn.Close()
							return nil, fmt.Errorf("invalid header : %q", text)
						}
						if resp.Header.Get(header[0]) == "" {
							resp.Header.Set(header[0], header[1])
						} else {
							resp.Header.Add(header[0], header[1])
						}

						if strings.ToLower(header[0]) == "content-length" {
							resp.ContentLength, err = strconv.ParseInt(header[1], 10, 64)
							if err != nil {
								conn.Close()
								return nil, err
							}
							remain = resp.ContentLength
						}
					}
				}
				cont = false
				break
			}
		}
	}

	// Body
	remain -= int64(ofs)
	if remain <= 0 {
		resp.Body = io.NopCloser(bytes.NewReader(buf[:ofs]))
		if c.Jar != nil {
			if rc := resp.Cookies(); len(rc) > 0 {
				c.Jar.SetCookies(req.URL, rc)
			}
		}
		return resp, conn.Close()
	}

	cont = true
	lastRequestTime := time.Now()
	for cont {
		for {
			end := ofs + 0x400
			if len(buf) < end {
				return nil, fmt.Errorf("slice out of range : use http.SetBuf() to change the allocation to %d bytes or more", end)
			}
			n, err := conn.Read(buf[ofs : ofs+0x400])
			if err != nil {
				return nil, err
			}
			if n == 0 {
				continue
			}
			if err != nil {
				conn.Close()
				return nil, err
			} else {
				ofs += n
				remain -= int64(n)
				if remain <= 0 {
					resp.Body = io.NopCloser(bytes.NewReader(buf[:ofs]))
					cont = false
					break
				}
				if time.Now().Sub(lastRequestTime).Milliseconds() >= 1000 {
					conn.Close()
					return nil, fmt.Errorf("time out")
				}
			}
		}
	}

	if c.Jar != nil {
		if rc := resp.Cookies(); len(rc) > 0 {
			c.Jar.SetCookies(req.URL, rc)
		}
	}

	return resp, conn.Close()
}
