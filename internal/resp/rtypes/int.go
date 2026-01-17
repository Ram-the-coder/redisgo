package rtypes

import (
	"bytes"
	"strconv"
)

type Int struct {
	Value int
}

func (ri *Int) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte(IntTypeId)
	buffer.WriteString(strconv.Itoa(ri.Value))
	buffer.WriteString("\r\n")
}
