// Redis serialization protocol (RESP)
package resp

import (
	"bytes"
	"net"

	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

func WriteResponse(rdt rtypes.RespDataType, conn net.Conn) error {
	var bytesBuffer bytes.Buffer
	rdt.WriteAsBytes(&bytesBuffer)
	response := bytesBuffer.Bytes()
	// fmt.Print("Responding with: \n" + string(response))
	log.Trace().Msgf("Responding with: %s", response)
	_, err := conn.Write(response)
	if err != nil {
		log.Err(err).Msgf("error writing response (%q) to connection", response)
	}
	return err
}
