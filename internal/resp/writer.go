// Redis serialization protocol (RESP)
package resp

import (
	"bytes"
	"errors"

	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
)

type ResponseBuilder struct {
	Value rtypes.RespDataType
}

func (rb *ResponseBuilder) Build() ([]byte, error) {
	if rb.Value != nil {
		var result bytes.Buffer
		rb.Value.WriteAsBytes(&result)
		return result.Bytes(), nil
	}
	return nil, errors.New("no value in response")
}
