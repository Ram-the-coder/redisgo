package rtypes

import (
	"bytes"
	"strconv"
)

type Map struct {
	KvPairs [][2]RespDataType // slice of an array of 2 elements (kv pair) - each element being a RespDataType
}

func (rm *Map) WriteAsBytes(buffer *bytes.Buffer) {
	buffer.WriteByte(MapTypeId)
	// Map Length
	buffer.WriteString(strconv.Itoa(len(rm.KvPairs)))
	buffer.WriteString("\r\n")
	// Map Items
	for _, kvPair := range rm.KvPairs {
		kvPair[0].WriteAsBytes(buffer)
		kvPair[1].WriteAsBytes(buffer)
	}
}
