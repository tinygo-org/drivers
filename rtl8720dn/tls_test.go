//go:build rtl8720dn_hosttest

package rtl8720dn

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"machine"
	"net/netip"
	"strings"
	"testing"

	"tinygo.org/x/drivers/netdev"
)

type uartWriter func([]byte) (int, error)

func (f uartWriter) Write(b []byte) (int, error) { return f(b) }

type uartReader func([]byte) (int, error)

func (f uartReader) Read(b []byte) (int, error) { return f(b) }

func tlsReply(msg []byte, result int32) []byte {
	b := make([]byte, 16)
	copy(b[4:12], msg[:8])
	b[4] = 2
	binary.LittleEndian.PutUint32(b[12:], uint32(result))
	binary.LittleEndian.PutUint16(b, 12)
	binary.LittleEndian.PutUint16(b[2:], computeCRC16(b[4:]))
	return b
}

func tlsDevice(start, socketResult int32, mutate func([]byte, byte) []byte) (*rtl8720dn, *int) {
	u := &machine.UART{Reader: bytes.NewReader(nil)}
	calls, writes := 0, 0
	u.Writer = uartWriter(func(msg []byte) (int, error) {
		writes++
		if writes%2 == 0 {
			calls++
			result := start
			if msg[1] == 6 {
				result = socketResult
			}
			b := tlsReply(msg, result)
			if mutate != nil {
				b = mutate(b, msg[1])
			}
			u.Reader = bytes.NewReader(b)
		}
		return len(msg), nil
	})
	r := &rtl8720dn{uart: u, sockets: map[sock]*socket{1: newSocket(netdev.IPPROTO_TLS)}}
	return r, &calls
}

func connectTLS(r *rtl8720dn) error {
	return r.Connect(1, "example.com", netip.MustParseAddrPort("192.0.2.1:443"))
}

func TestTLSResults(t *testing.T) {
	for _, tc := range []struct {
		name          string
		start, socket int32
		want          string
		calls         int
	}{
		{"positive", 3, 3, "", 2},
		{"zero", 0, 0, "", 2},
		{"minus one", -1, 0, "code -1", 1},
		{"mbedTLS", -0x2700, 0, "code -9984", 1},
		{"closed socket", 0, -1, "socket -1", 2},
		{"negative socket", 3, -2, "socket -2", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, calls := tlsDevice(tc.start, tc.socket, nil)
			err := connectTLS(r)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.want) || !strings.Contains(err.Error(), "example.com:443") {
				t.Fatalf("got %v, want %s with host and port", err, tc.want)
			}
			if *calls != tc.calls {
				t.Fatalf("got %d RPC calls, want %d", *calls, tc.calls)
			}
			if r.captureRPCErr {
				t.Fatal("TLS transport checks remain active")
			}
		})
	}
}

func TestTLSInvalidReplies(t *testing.T) {
	for _, method := range []byte{0x12, 6} {
		for _, tc := range []struct {
			name    string
			mutate  func([]byte) []byte
			want    string
			wantErr error
		}{
			{"CRC", func(b []byte) []byte { b[2] ^= 1; return b }, "invalid RPC response CRC", nil},
			{"no header", func(b []byte) []byte { return b[:0] }, "", io.EOF},
			{"short header", func(b []byte) []byte { return b[:2] }, "", io.ErrUnexpectedEOF},
			{"short payload", func(b []byte) []byte { return b[:14] }, "", io.ErrUnexpectedEOF},
			{"long response", func(b []byte) []byte {
				b = append(b, 0, 0, 0, 0)
				binary.LittleEndian.PutUint16(b, 16)
				binary.LittleEndian.PutUint16(b[2:], computeCRC16(b[4:]))
				return b
			}, "invalid RPC response length 16", nil},
			{"zero length", func(b []byte) []byte { binary.LittleEndian.PutUint16(b, 0); return b }, "invalid RPC frame length 0", nil},
			{"oversized", func(b []byte) []byte { binary.LittleEndian.PutUint16(b, 65535); return b }, "invalid RPC frame length 65535", nil},
			{"type", func(b []byte) []byte { b[4] = 0; return b }, "unsupported incoming RPC invocation", nil},
			{"unknown type", func(b []byte) []byte { b[4] = 0xff; return b }, "invalid RPC message type 255", nil},
			{"method", func(b []byte) []byte { b[5]++; return b }, "invalid RPC response header", nil},
			{"service", func(b []byte) []byte { b[6]++; return b }, "invalid RPC response header", nil},
			{"version", func(b []byte) []byte { b[7]++; return b }, "invalid RPC response header", nil},
			{"sequence", func(b []byte) []byte { b[8]++; return b }, "invalid RPC response header", nil},
		} {
			t.Run(fmt.Sprintf("%s/method%d", tc.name, method), func(t *testing.T) {
				r, _ := tlsDevice(3, 3, func(b []byte, m byte) []byte {
					if m != method {
						return b
					}
					b = tc.mutate(b)
					if len(b) == 16 && tc.name != "CRC" {
						binary.LittleEndian.PutUint16(b[2:], computeCRC16(b[4:]))
					}
					return b
				})
				err := connectTLS(r)
				if err == nil {
					t.Fatal("invalid reply was accepted")
				}
				if r.captureRPCErr {
					t.Fatal("TLS transport checks remain active after failure")
				}
				if tc.wantErr != nil {
					if !errors.Is(err, tc.wantErr) {
						t.Fatalf("got %v, want %v", err, tc.wantErr)
					}
				} else if cause := errors.Unwrap(err); cause == nil || cause.Error() != tc.want {
					t.Fatalf("got %v, want %s", err, tc.want)
				}
				if !bytes.Equal(payload[:12], make([]byte, 12)) {
					t.Fatal("failed reply was not cleared")
				}
			})
		}
	}
}

func TestTLSUARTFailures(t *testing.T) {
	fault := errors.New("UART failure")
	for _, failWrite := range []int{1, 2, 3, 4} {
		for _, short := range []bool{false, true} {
			r, _ := tlsDevice(3, 3, nil)
			writer := r.uart.Writer
			writes := 0
			r.uart.Writer = uartWriter(func(b []byte) (int, error) {
				writes++
				if writes == failWrite {
					if short {
						return len(b) - 1, nil
					}
					return 0, fault
				}
				return writer.Write(b)
			})
			err := connectTLS(r)
			want := fault
			if short {
				want = io.ErrShortWrite
			}
			if !errors.Is(err, want) {
				t.Fatalf("write %d got %v, want %v", failWrite, err, want)
			}
		}
	}
	for _, failRead := range []int{1, 2, 3, 4} {
		r, _ := tlsDevice(3, 3, nil)
		writer := r.uart.Writer
		reads := 0
		r.uart.Writer = uartWriter(func(b []byte) (int, error) {
			n, err := writer.Write(b)
			reader := r.uart.Reader
			r.uart.Reader = uartReader(func(b []byte) (int, error) {
				reads++
				if reads == failRead {
					return 0, fault
				}
				return reader.Read(b)
			})
			return n, err
		})
		if err := connectTLS(r); !errors.Is(err, fault) {
			t.Fatalf("read %d got %v", failRead, err)
		}
	}
}

func TestTLSConsecutiveRequests(t *testing.T) {
	r, _ := tlsDevice(0, 0, nil)
	r.captureRPCErr = true
	writer := r.uart.Writer
	r.uart.Writer = uartWriter(func([]byte) (int, error) { return 0, io.ErrClosedPipe })
	r.rpc_wifi_ssl_get_socket(1)
	if !errors.Is(r.rpcErr, io.ErrClosedPipe) {
		t.Fatal(r.rpcErr)
	}
	r.uart.Writer = writer
	if result := r.rpc_wifi_ssl_get_socket(1); result != 0 || r.rpcErr != nil {
		t.Fatalf("result %d, error %v", result, r.rpcErr)
	}
	r.captureRPCErr = false
	if err := connectTLS(r); err != nil {
		t.Fatal(err)
	}
}

func TestOtherRPCPaths(t *testing.T) {
	for _, protocol := range []int{netdev.IPPROTO_TCP, netdev.IPPROTO_UDP} {
		r, calls := tlsDevice(0, 0, nil)
		r.sockets[1].protocol = protocol
		r.rpcErr = io.ErrClosedPipe
		if err := connectTLS(r); err != nil {
			t.Fatal(err)
		}
		if r.captureRPCErr || r.rpcErr != nil {
			t.Fatalf("legacy RPC left capture %t, error %v", r.captureRPCErr, r.rpcErr)
		}
		if *calls != 1 {
			t.Fatalf("got %d calls", *calls)
		}
	}
	r, _ := tlsDevice(0, 0, nil)
	r.uart.Writer = uartWriter(func(msg []byte) (int, error) {
		if len(msg) >= 8 {
			b := tlsReply(msg, 3)
			b = append(b, []byte("abc")...)
			binary.LittleEndian.PutUint16(b, uint16(len(b)-4))
			binary.LittleEndian.PutUint16(b[2:], computeCRC16(b[4:]))
			r.uart.Reader = bytes.NewReader(b)
		}
		return len(msg), nil
	})
	if result := r.rpc_system_version(); result != "abc" {
		t.Fatalf("got %q", result)
	}
}

func tlsNotification(length int) []byte {
	b := make([]byte, length+4)
	b[4] = 0x03
	b[7] = xVersion
	binary.LittleEndian.PutUint16(b, uint16(length))
	binary.LittleEndian.PutUint16(b[2:], computeCRC16(b[4:]))
	return b
}

func TestTLSNotifications(t *testing.T) {
	for _, messageType := range []byte{0x01, 0x03} {
		for _, method := range []byte{0x12, 6} {
			for _, length := range []int{8, 12, 20, len(payload)} {
				t.Run(fmt.Sprintf("type%d/method%d/length%d", messageType, method, length), func(t *testing.T) {
					r, calls := tlsDevice(0, 0, func(b []byte, m byte) []byte {
						if m != method {
							return b
						}
						frame := tlsNotification(length)
						frame[4] = messageType
						binary.LittleEndian.PutUint16(frame[2:], computeCRC16(frame[4:]))
						frames := append(frame, tlsNotification(8)...)
						return append(frames, b...)
					})
					if err := connectTLS(r); err != nil {
						t.Fatal(err)
					}
					if *calls != 2 {
						t.Fatalf("got %d RPC calls, want 2", *calls)
					}
				})
			}
		}
	}
}

func TestTLSInvalidNotifications(t *testing.T) {
	for _, messageType := range []byte{0x01, 0x03} {
		for _, method := range []byte{0x12, 6} {
			for _, tc := range []struct {
				name    string
				want    string
				wantErr error
			}{
				{"CRC", "invalid RPC response CRC", nil},
				{"version", "invalid RPC response header", nil},
				{"zero length", "invalid RPC frame length 0", nil},
				{"oversized", "invalid RPC frame length 65535", nil},
				{"unknown type", "invalid RPC message type 255", nil},
				{"short payload", "", io.ErrUnexpectedEOF},
				{"UART", "", io.ErrClosedPipe},
			} {
				t.Run(fmt.Sprintf("type%d/method%d/%s", messageType, method, tc.name), func(t *testing.T) {
					kind := tc.name
					r, _ := tlsDevice(0, 0, nil)
					writer := r.uart.Writer
					r.uart.Writer = uartWriter(func(msg []byte) (int, error) {
						n, err := writer.Write(msg)
						if len(msg) < 8 || msg[1] != method {
							return n, err
						}
						notification := tlsNotification(20)
						notification[4] = messageType
						binary.LittleEndian.PutUint16(notification[2:], computeCRC16(notification[4:]))
						switch kind {
						case "CRC":
							notification[2] ^= 1
						case "version":
							notification[7]++
							binary.LittleEndian.PutUint16(notification[2:], computeCRC16(notification[4:]))
						case "zero length":
							binary.LittleEndian.PutUint16(notification, 0)
						case "oversized":
							binary.LittleEndian.PutUint16(notification, 65535)
						case "unknown type":
							notification[4] = 0xff
							binary.LittleEndian.PutUint16(notification[2:], computeCRC16(notification[4:]))
						case "short payload":
							notification = notification[:10]
						case "UART":
							r.uart.Reader = io.MultiReader(bytes.NewReader(notification[:10]), uartReader(func([]byte) (int, error) {
								return 0, io.ErrClosedPipe
							}))
							return n, err
						}
						if kind == "short payload" {
							r.uart.Reader = bytes.NewReader(notification)
						} else {
							r.uart.Reader = io.MultiReader(bytes.NewReader(notification), r.uart.Reader)
						}
						return n, err
					})
					err := connectTLS(r)
					if err == nil {
						t.Fatal("invalid notification was accepted")
					}
					if tc.wantErr != nil {
						if !errors.Is(err, tc.wantErr) {
							t.Fatalf("got %v, want %v", err, tc.wantErr)
						}
					} else if cause := errors.Unwrap(err); cause == nil || cause.Error() != tc.want {
						t.Fatalf("got %v, want %s", err, tc.want)
					}
				})
			}
		}
	}
}

func TestTLSUARTWait(t *testing.T) {
	for _, method := range []byte{0x12, 6} {
		for _, tc := range []struct {
			name       string
			pauseAt    int
			emptyReads int
			fail       bool
		}{
			{"short wait", 0, 3, false},
			{"long wait", 0, 150, false},
			{"partial header", 2, 3, false},
			{"partial payload", 8, 3, false},
			{"error after wait", 0, 150, true},
			{"header error after wait", 2, 3, true},
			{"payload error after wait", 8, 3, true},
		} {
			t.Run(fmt.Sprintf("method%d/%s", method, tc.name), func(t *testing.T) {
				r, _ := tlsDevice(0, 0, nil)
				writer := r.uart.Writer
				emptyReads := 0
				fault := errors.New("UART failure after wait")
				r.uart.Writer = uartWriter(func(msg []byte) (int, error) {
					n, err := writer.Write(msg)
					if len(msg) < 8 || msg[1] != method {
						return n, err
					}
					reader := r.uart.Reader
					position := 0
					r.uart.Reader = uartReader(func(b []byte) (int, error) {
						if position == tc.pauseAt {
							if emptyReads < tc.emptyReads {
								emptyReads++
								return 0, nil
							}
							if tc.fail {
								return 0, fault
							}
						}
						n, err := reader.Read(b[:1])
						position += n
						return n, err
					})
					return n, err
				})
				err := connectTLS(r)
				if tc.fail {
					if !errors.Is(err, fault) {
						t.Fatalf("got %v, want %v", err, fault)
					}
				} else if err != nil {
					t.Fatal(err)
				}
				if emptyReads != tc.emptyReads {
					t.Fatalf("got %d empty reads, want %d", emptyReads, tc.emptyReads)
				}
			})
		}
	}
}

func TestBlockingUARTReader(t *testing.T) {
	calls := 0
	fault := errors.New("UART read failure")
	u := &machine.UART{Reader: uartReader(func(b []byte) (int, error) {
		calls++
		if calls <= 150 {
			return 0, nil
		}
		b[0] = 42
		return 1, fault
	})}
	reader := blockingUARTReader{uart: u}
	if n, err := reader.Read(nil); n != 0 || err != nil || calls != 0 {
		t.Fatalf("empty buffer got %d, %v, %d calls", n, err, calls)
	}
	var b [4]byte
	n, err := reader.Read(b[:])
	if n != 1 || !errors.Is(err, fault) || b[0] != 42 || calls != 151 {
		t.Fatalf("got %d, %v, data %d, %d calls", n, err, b[0], calls)
	}
}
