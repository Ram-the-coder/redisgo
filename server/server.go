package server

import (
	"errors"
	"fmt"
	"net"

	"github.com/ram-the-coder/redisgo/internal"
	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/rs/zerolog/log"
)

type Server struct {
	address  string
	listener net.Listener
	stopCh   chan struct{}
	store    *internal.Store
}

func NewServer(address string) *Server {
	return &Server{
		address: address,
		stopCh:  make(chan struct{}),
		store:   internal.NewStore(),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.address, err)
	}
	s.listener = ln
	addressListeningOn, _ := s.getAddressListeningOn()
	log.Info().Msgf("Redisgo server started and listening on %s", addressListeningOn)

	go s.acceptConnectionLoop()
	return nil
}

func (s *Server) Stop() {
	log.Info().Msg("Stopping server...")
	close(s.stopCh)    // Stop waiting for new connections on the listener
	s.listener.Close() // Stop listening on the port
}

func (s *Server) getAddressListeningOn() (string, error) {
	if s.listener == nil {
		return "", errors.New("cannot get listener address when server is not running")
	}
	return s.listener.Addr().String(), nil
}

func (s *Server) acceptConnectionLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
				// Channel closed. Redisgo is stopping. Exit the loop.
				return
			default:
				log.Err(err).Msgf("Error accepting connection")
				continue
			}
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		command, err := resp.ReadCommand(conn)
		if err != nil {
			log.Err(err).Msg("Failed to read command")
			return
		}
		if command != nil {
			Handle(command, conn, s.store)
		}
	}
}
