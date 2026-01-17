package server

import (
	"fmt"
	"net"

	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

const CommandHello = "hello"
const CommandPing = "ping"

func Handle(command *resp.Command, conn net.Conn) error {
	if command.Name == CommandHello {
		log.Info().Msgf("Handling command: %s. Args: %s", command.Name, command.Arguments)
		kvPairs := [][2]rtypes.RespDataType{
			{rtypes.NewBulkString("server"), rtypes.NewBulkString("redis")},
			{rtypes.NewBulkString("version"), rtypes.NewBulkString("8.4.0")},
			{rtypes.NewBulkString("proto"), &rtypes.Int{Value: 3}},
			{rtypes.NewBulkString("id"), &rtypes.Int{Value: 1}},
			{rtypes.NewBulkString("mode"), rtypes.NewBulkString("standalone")},
			{rtypes.NewBulkString("role"), rtypes.NewBulkString("master")},
			{rtypes.NewBulkString("modules"), &rtypes.Array{Elements: []rtypes.RespDataType{}}},
		}
		response := resp.ResponseBuilder{Value: &rtypes.Map{KvPairs: kvPairs}}
		result, err := response.Build()
		if err != nil {
			return fmt.Errorf("error in building response: %w", err)
		}
		// fmt.Print("Responding with: \n" + string(result))
		log.Info().Msgf("Responding with: %s", result)
		conn.Write(result)
		return nil
	}
	if command.Name == CommandPing {
		res := resp.ResponseBuilder{Value: rtypes.NewSimpleString("PONG")}
		result, _ := res.Build()
		log.Info().Msgf("Responding with: %s", result)
		conn.Write(result)
		return nil
	}

	res := resp.ResponseBuilder{Value: rtypes.NewSimpleError("ERR unknown command")}
	result, _ := res.Build()
	conn.Write(result)
	log.Error().Msgf("Unhandled command: %s. Args: %s. Sending ERR.", command.Name, command.Arguments)
	return nil
}
