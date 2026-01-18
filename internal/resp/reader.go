package resp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

type Command struct {
	Name      string
	Arguments []rtypes.RespDataType
}

func ReadCommand(reader *bufio.Reader) (*Command, error) {
	element, err := readElement(reader)
	if err != nil {
		return nil, err
	}
	switch e := element.(type) {
	case *rtypes.Array:
		if len(e.Elements) == 0 {
			return nil, fmt.Errorf("empty array is not part of a recognized command")
		}
		commandName, ok := e.Elements[0].(*rtypes.BulkString)
		if !ok {
			return nil, fmt.Errorf("the first element %s was not a string", element)
		}
		return &Command{
			Name:      string(commandName.Value),
			Arguments: e.Elements[1:],
		}, nil
	// Not yet handling non-array top level elements
	default:
		return nil, fmt.Errorf("failed to read command, unhandled element: %s", element)
	}
}

func readElement(reader *bufio.Reader) (rtypes.RespDataType, error) {
	dataType, err := reader.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read first byte of the command: %w", err)
	}

	if dataType == rtypes.ArrayTypeId {
		arrayLength, err := readInteger(reader)
		if err != nil {
			return nil, err
		}
		if arrayLength < 0 {
			return nil, fmt.Errorf("invalid array length: %d", arrayLength)
		}
		values := make([]rtypes.RespDataType, arrayLength)
		for i := range arrayLength {
			values[i], err = readElement(reader)
			if err != nil {
				log.Error().Str("current_values", fmt.Sprintf("%s", values)).Msg("failed to read array element")
				return nil, fmt.Errorf("failed to read element at index %d of array: %w", i, err)
			}
		}
		return &rtypes.Array{Elements: values}, nil
	}

	if dataType == rtypes.BulkStringTypeId {
		strLength, err := readInteger(reader)
		if err != nil {
			return nil, err
		}
		str := make([]byte, strLength)
		if _, err = io.ReadFull(reader, str); err != nil {
			return nil, fmt.Errorf("failed to read string: %w", err)
		}
		// consume the next 2 bytes (\r\n)
		if _, err = io.ReadFull(reader, make([]byte, 2)); err != nil {
			return nil, fmt.Errorf("failed to read delimiter for string: %w", err)
		}
		return &rtypes.BulkString{Value: str}, nil
	}
	return nil, errors.New("failed to read unhandled element")
}

func readInteger(reader *bufio.Reader) (int, error) {
	intString, err := reader.ReadString('\r')
	if err != nil {
		return 0, fmt.Errorf("failed to read integer: %w", err)
	}
	integer, err := strconv.Atoi(intString[:len(intString)-1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse integer \"%s\": %w", intString, err)
	}

	// read the '\n'
	if _, err = reader.ReadByte(); err != nil {
		return 0, fmt.Errorf("failed to integer terminator: %w", err)
	}
	return integer, nil
}
