package rtypes

import (
	"bytes"
	"strconv"
)

type Array struct {
	Elements []RespDataType
}

func (ra *Array) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte('*')
	buffer.WriteString(strconv.Itoa(len(ra.Elements)))
	buffer.WriteString("\r\n")
	for _, element := range ra.Elements {
		element.WriteAsBytes(buffer)
	}
}
