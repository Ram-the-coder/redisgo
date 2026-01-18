package handlers

import (
	"github.com/ram-the-coder/redisgo/internal"
	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

func GetResponseForGeneralCommand() func(*internal.Command) (rtypes.RespDataType, error) {
	return func(cmd *internal.Command) (rtypes.RespDataType, error) {
		switch cmd.Name {
		case internal.CommandHello:
			kvPairs := [][2]rtypes.RespDataType{
				{rtypes.NewBulkString("server"), rtypes.NewBulkString("redis")},
				{rtypes.NewBulkString("version"), rtypes.NewBulkString("8.4.0")},
				{rtypes.NewBulkString("proto"), &rtypes.Int{Value: 3}},
				{rtypes.NewBulkString("id"), &rtypes.Int{Value: 1}},
				{rtypes.NewBulkString("mode"), rtypes.NewBulkString("standalone")},
				{rtypes.NewBulkString("role"), rtypes.NewBulkString("master")},
				{rtypes.NewBulkString("modules"), &rtypes.Array{Elements: []rtypes.RespDataType{}}},
			}
			return &rtypes.Map{KvPairs: kvPairs}, nil
		case internal.CommandPing:
			return rtypes.NewSimpleString("PONG"), nil
		default:
			log.Error().Msgf("unknown general command: %s", cmd.Name)
			return rtypes.NewSimpleError("ERR unknown command"), nil
		}
	}
}
