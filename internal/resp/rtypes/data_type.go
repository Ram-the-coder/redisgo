package rtypes

import "bytes"

const ArrayTypeId byte = '*'
const BulkStringTypeId byte = '$'
const IntTypeId byte = ':'
const MapTypeId byte = '%'
const SimpleErrorTypeId byte = '-'
const SimpleStringTypeId byte = '+'
const NullTypeId byte = '_'

type RespDataType interface {
	WriteAsBytes(*bytes.Buffer)
}
