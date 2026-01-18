package server

import (
	"fmt"
	"net"
	"strings"

	"github.com/ram-the-coder/redisgo/internal"
	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

const CommandHello = "hello"
const CommandPing = "ping"
const CommandSet = "set"
const CommandGet = "get"

func Handle(command *resp.Command, conn net.Conn, store *internal.Store) error {
	log.Trace().Msgf("Handling command: %s. Args: %s", command.Name, command.Arguments)
	switch strings.ToLower(command.Name) {
	case CommandHello:
		kvPairs := [][2]rtypes.RespDataType{
			{rtypes.NewBulkString("server"), rtypes.NewBulkString("redis")},
			{rtypes.NewBulkString("version"), rtypes.NewBulkString("8.4.0")},
			{rtypes.NewBulkString("proto"), &rtypes.Int{Value: 3}},
			{rtypes.NewBulkString("id"), &rtypes.Int{Value: 1}},
			{rtypes.NewBulkString("mode"), rtypes.NewBulkString("standalone")},
			{rtypes.NewBulkString("role"), rtypes.NewBulkString("master")},
			{rtypes.NewBulkString("modules"), &rtypes.Array{Elements: []rtypes.RespDataType{}}},
		}
		return resp.WriteResponse(&rtypes.Map{KvPairs: kvPairs}, conn)

	case CommandPing:
		return resp.WriteResponse(rtypes.NewSimpleString("PONG"), conn)

	case CommandSet:
		keyStr, err := getString(command.Arguments[0])
		if err != nil {
			log.Err(err).Msg("failed to parse key")
			return writeErrorResponse(conn)
		}
		valueStr, err := getString(command.Arguments[1])
		if err != nil {
			log.Err(err).Msg("failed to parse value")
			return writeErrorResponse(conn)
		}
		store.Set(keyStr, valueStr)
		return resp.WriteResponse(rtypes.NewSimpleString("OK"), conn)

	case CommandGet:
		keyStr, err := getString(command.Arguments[0])
		if err != nil {
			log.Err(err).Msg("failed to parse key")
			return writeErrorResponse(conn)
		}

		value, ok, err := store.Get(keyStr)
		log.Trace().Msgf("Got value: %s", value)
		if err != nil {
			return writeErrorResponse(conn)
		}
		if !ok {
			return resp.WriteResponse(&rtypes.Null{}, conn)
		}
		resp.WriteResponse(rtypes.NewBulkString(string(value)), conn)

	default:
		return writeErrorResponse(conn)
	}

	return writeErrorResponse(conn)
}

func writeErrorResponse(conn net.Conn) error {
	return resp.WriteResponse(rtypes.NewSimpleError("ERR unknown command"), conn)
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
