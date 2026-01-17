package rtypes

import "bytes"

type SimpleError struct {
	Value []byte
}

func NewSimpleError(str string) *SimpleError {
	return &SimpleError{Value: []byte(str)}
}

func (rss *SimpleError) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte(SimpleErrorTypeId)
	buffer.Write(rss.Value)
	buffer.WriteString("\r\n")
}
