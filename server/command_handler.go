package server

import (
	"net"

	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

const CommandHello = "hello"
const CommandPing = "ping"

func Handle(command *resp.Command, conn net.Conn) error {
	log.Info().Msgf("Handling command: %s. Args: %s", command.Name, command.Arguments)
	if command.Name == CommandHello {
		kvPairs := [][2]rtypes.RespDataType{
			{rtypes.NewBulkString("server"), rtypes.NewBulkString("redis")},
			{rtypes.NewBulkString("version"), rtypes.NewBulkString("8.4.0")},
			{rtypes.NewBulkString("proto"), &rtypes.Int{Value: 3}},
			{rtypes.NewBulkString("id"), &rtypes.Int{Value: 1}},
			{rtypes.NewBulkString("mode"), rtypes.NewBulkString("standalone")},
			{rtypes.NewBulkString("role"), rtypes.NewBulkString("master")},
			{rtypes.NewBulkString("modules"), &rtypes.Array{Elements: []rtypes.RespDataType{}}},
		}
		response := resp.GetResponse(&rtypes.Map{KvPairs: kvPairs})
		// fmt.Print("Responding with: \n" + string(result))
		log.Info().Msgf("Responding with: %s", response)
		conn.Write(response)
		return nil
	}
	if command.Name == CommandPing {
		res := resp.GetResponse(rtypes.NewSimpleString("PONG"))
		log.Info().Msgf("Responding with: %s", res)
		conn.Write(res)
		return nil
	}

	res := resp.GetResponse(rtypes.NewSimpleError("ERR unknown command"))
	log.Info().Msgf("Responding with: %s", res)
	conn.Write(res)
	return nil
}
