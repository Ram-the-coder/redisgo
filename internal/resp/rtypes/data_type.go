package rtypes

import "bytes"

type RespDataType interface {
	WriteAsBytes(*bytes.Buffer)
}
