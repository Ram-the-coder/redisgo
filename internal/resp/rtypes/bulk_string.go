package rtypes

import (
	"bytes"
	"strconv"
)

type BulkString struct {
	Value []byte
}

func NewBulkString(str string) *BulkString {
	return &BulkString{Value: []byte(str)}
}

func (rbs *BulkString) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte('$')
	buffer.WriteString(strconv.Itoa(len(rbs.Value)))
	buffer.WriteString("\r\n")
	buffer.Write(rbs.Value)
	buffer.WriteString("\r\n")
}

func (rbs *BulkString) String() string {
	return string(rbs.Value)
}
