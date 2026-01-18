package handlers

import (
	"fmt"

	"github.com/ram-the-coder/redisgo/internal"
	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

func GetResponseForStoreCommand(store *internal.Store) func(*internal.Command) (rtypes.RespDataType, error) {
	return func(cmd *internal.Command) (rtypes.RespDataType, error) {
		switch cmd.Name {
		case internal.CommandSet:
			keyStr, err := getString(cmd.Arguments[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse key")
			}
			valueStr, err := getString(cmd.Arguments[1])
			if err != nil {
				return nil, fmt.Errorf("failed to parse value")

			}
			store.Set(keyStr, valueStr)
			return rtypes.NewSimpleString("OK"), nil

		case internal.CommandGet:
			keyStr, err := getString(cmd.Arguments[0])
			if err != nil {
				return nil, fmt.Errorf("failed to parse key")
			}

			value, ok, err := store.Get(keyStr)
			log.Trace().Msgf("Got value: %s", value)
			if err != nil {
				return nil, err
			}
			if !ok {
				return &rtypes.Null{}, nil
			}
			return rtypes.NewBulkString(string(value)), nil
		default:
			log.Error().Msgf("unknown store command: %s", cmd.Name)
			return rtypes.NewSimpleError("ERR unknown command"), nil
		}
	}
}

func getString(rdt rtypes.RespDataType) (string, error) {
	if key, ok := rdt.(*rtypes.BulkString); ok {
		return string(key.Value), nil
	} else if key, ok := rdt.(*rtypes.SimpleString); ok {
		return string(key.Value), nil
	} else {
		return "", fmt.Errorf("failed to get string fromL %s", rdt)
	}
}
