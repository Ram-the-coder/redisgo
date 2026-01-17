package rtypes

import (
	"bytes"
	"strconv"
)

type Array struct {
	Elements []RespDataType
}

func (ra *Array) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte(ArrayTypeId)
	
	// Array length
	buffer.WriteString(strconv.Itoa(len(ra.Elements)))
	buffer.WriteString("\r\n")
	
	// Array elements
	for _, element := range ra.Elements {
		element.WriteAsBytes(buffer)
	}
}
