package internal

import (
	"fmt"
	"net"

	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
)

const (
	CommandGet   = "get"
	CommandHello = "hello"
	CommandPing  = "ping"
	CommandSet   = "set"
)

const (
	CommandTypeStore   = "store"
	CommandTypeGeneral = "general"
)

var commandToTypeMap = map[string]string{
	CommandGet:   CommandTypeStore,
	CommandHello: CommandTypeGeneral,
	CommandPing:  CommandTypeGeneral,
	CommandSet:   CommandTypeStore,
}

type CommandMeta struct {
	Conn net.Conn
}

type Command struct {
	Name      string
	Arguments []rtypes.RespDataType
	Metadata  CommandMeta
}

func (c *Command) GetType() (string, error) {
	if c == nil {
		return "", fmt.Errorf("c *Command is nil")
	}
	cmdType, ok := commandToTypeMap[c.Name]
	if !ok {
		return "", fmt.Errorf("command %q does not have a type", c.Name)
	}
	return cmdType, nil
}
