package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ram-the-coder/redisgo/server"
)

func main() {
	s := server.NewServer(":6379")	
	if err := s.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig // block until Ctrl+C or kill
	s.Stop()
}