// Redis serialization protocol (RESP)
package resp

import (
	"bytes"

	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
)

func GetResponse(rdt rtypes.RespDataType) []byte {
	var bytesBuffer bytes.Buffer
	rdt.WriteAsBytes(&bytesBuffer)
	return bytesBuffer.Bytes()
}
