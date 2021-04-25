package bytealg

import (
	"strings"
	"testing"
)

func TestRabinKarpBytes(t *testing.T) {
	testCases := []struct {
		target, substr string
	}{
		{
			target: "GET /arp HTTP/1.1\r\ncontent-type: text/plain\r\nencoding: gzip\r\n\r\n<h1>Hello!</h1>",
			substr: "\r\n",
		},
		{
			target: "kasdfkl;sdfakl;'asdfkl;'asdfkl;'asdfkl;'gavkl;'vbkml'afop'rkfgopakfgopaf,'sdf;kwepokf",
			substr: "rkf",
		},
		{
			target: "",
			substr: "",
		},
		{
			target: "asdasdasdasd",
			substr: "",
		},
		{
			target: "",
			substr: "bad",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.substr, func(t *testing.T) {
			expect := strings.Index(tC.target, tC.substr)
			got := IdxRabinKarpBytes([]byte(tC.target), []byte(tC.substr))
			if got != expect {
				t.Errorf("expected %v, got %v", expect, got)
			}
		})
	}
}
