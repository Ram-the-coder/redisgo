package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ram-the-coder/redisgo/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	s := server.NewServer(":6379")
	if err := s.Start(); err != nil {
		log.Fatal().Msgf("server failed: %v", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig // block until Ctrl+C or kill
	s.Stop()
}
