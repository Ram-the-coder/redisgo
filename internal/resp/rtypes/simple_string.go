package rtypes

import "bytes"

type SimpleString struct {
	Value []byte
}

func NewSimpleString(str string) *SimpleString {
	return &SimpleString{Value: []byte(str)}
}

func (rss *SimpleString) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte(SimpleStringTypeId)
	buffer.Write(rss.Value)
	buffer.WriteString("\r\n")
}

func (rss *SimpleString) String() string {
	return string(rss.Value)
}
