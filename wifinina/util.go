package wifinina

import (
	"bytes"
	"fmt"
	"io"
)

func PrintBuffer(ninabuf *Buffer, w io.Writer) error {
	var buf bytes.Buffer
	if _, err := ninabuf.WriteTo(&buf); err != nil {
		return err
	}
	commandReply := "Command"
	if ninabuf.IsReply() {
		commandReply = "Reply"
	}
	sl := buf.Bytes()
	pLenSize := int(ninabuf.ParamLenSize())
	fmt.Fprintln(w, "Start command: ", sl[0] == CmdStart)
	fmt.Fprintln(w, "  Command/Reply:  ", commandReply)
	fmt.Fprintf(w, "  Command Byte:    %02X\n", ninabuf.Command())
	fmt.Fprintln(w, "  Param Len Size: ", pLenSize)
	fmt.Fprintln(w, "  Number Params:  ", ninabuf.NumParams())
	var pos = 3
	for i := 0; i < int(ninabuf.NumParams()); i++ {
		pLen := int(sl[pos])
		if pLenSize == 2 {
			pLen = (pLen << 8) | int(sl[pos+1])
		}
		pslice := sl[pos+pLenSize : pos+pLenSize+pLen]
		fmt.Fprintf(w, "    Parameter %d (pos %d, length %d): %v\n", i, pos+pLenSize, pLen, pslice)
		pos += pLenSize + pLen
	}
	fmt.Fprintln(w, "End command: ", sl[pos] == CmdEnd)
	for i := pos + 1; i < len(sl); i++ {
		fmt.Fprintln(w, "Padding: ", sl[i] == 0xFF)
	}
	return nil
}
