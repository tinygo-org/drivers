// +build !tinygo

package wifinina

import (
	"os"
	"testing"
)

func TestBuffer_GetFwVersion(t *testing.T) {
	buffer := NewBuffer(256)
	buffer.StartCmd(CmdGetFwVersion)
	buffer.EndCmd()
	t.Logf("% 02x", buffer.buf)
}

func TestBuffer_TestStringParam(t *testing.T) {
	buffer := NewBuffer(256)
	buffer.StartCmd(CmdReqHostByName)
	buffer.AddString("numbersapi.com")
	buffer.EndCmd()
	PrintBuffer(buffer, os.Stdout)
	if buffer.ParamLenSize() != 1 {
		t.Error("expected param len size 1")
	}
}

func TestBuffer_GetDatabufTCP(t *testing.T) {
	buffer := NewBuffer(256)
	sock := byte(0)
	buf := make([]byte, 64)
	p := uint16(len(buf))
	buffer.StartCmd(CmdGetDatabufTCP)
	buffer.AddByte(sock)
	//d.cmdbuf.AddUint16(p)
	buffer.AddData([]byte{uint8(p & 0x00FF), uint8((p) >> 8)}) // TODO: is this the right byte order?
	buffer.EndCmd()
	PrintBuffer(buffer, os.Stdout)
	if buffer.ParamLenSize() != 2 {
		t.Error("expected param len size 2")
	}
	t.Logf("% 02x", buffer.buf)
}

func TestBuffer_SetPinMode(t *testing.T) {
	buffer := NewBuffer(256)
	buffer.StartCmd(CmdSetPinMode)
	buffer.AddByte(27)
	buffer.AddUint16(128)
	buffer.AddUint32(65)
	buffer.AddString("hello world")
	buffer.EndCmd()
	PrintBuffer(buffer, os.Stdout)
	if buffer.ParamLenSize() != 1 {
		t.Error("expected param len size 1")
	}
	t.Logf("% 02x", buffer.buf)
}

func TestBuffer_SetData(t *testing.T) {
	buffer := NewBuffer(256)
	buffer.StartCmd(CmdInsertDataBuf)
	buffer.AddData([]byte{1, 2, 3, 4})
	buffer.EndCmd()
	PrintBuffer(buffer, os.Stdout)
	t.Logf("% 02x", buffer.buf)
}
