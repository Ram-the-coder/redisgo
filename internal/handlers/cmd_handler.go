package handlers

import (
	"github.com/ram-the-coder/redisgo/internal"
	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/ram-the-coder/redisgo/internal/resp/rtypes"
	"github.com/rs/zerolog/log"
)

func HandleCommands(
		commandCh <-chan *internal.Command,
		stopCh <-chan struct{},
		getResponse func(*internal.Command) (rtypes.RespDataType, error)) {
	for {
		select {
		case <-stopCh:
			return
		case command := <-commandCh:
			log.Trace().Msgf("Handling command: %s. Args: %s", command.Name, command.Arguments)
			response, err := getResponse(command)
			if err != nil {
				log.Err(err).Msgf("failed to build response for command %q", command.Name)
				continue
			}
			if err := resp.WriteResponse(response, command.Metadata.Conn); err != nil {
				log.Err(err).Msgf("failed to write response for command %q", command.Name)
				continue
			}
		}
	}
}
