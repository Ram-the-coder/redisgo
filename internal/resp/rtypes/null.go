package rtypes

import "bytes"

type Null struct{}

func (rn *Null) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte(NullTypeId)
	buffer.WriteString("\r\n")
}
