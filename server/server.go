package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/ram-the-coder/redisgo/internal"
	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/rs/zerolog/log"
)

type Server struct {
	address                string
	listener               net.Listener
	stopCh                 chan struct{}
	store                  *internal.Store
	handlingDelayMsForTest atomic.Int64
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
	reader := bufio.NewReader(conn)
	for {
		command, err := resp.ReadCommand(reader)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				log.Trace().Msgf("connection closed: %s", err)
			} else {
				log.Err(err).Msgf("failed to read command")
			}
			return
		}
		if command != nil {
			s.addDelayForTesting()
			if err := Handle(command, conn, s.store); err != nil {
				log.Err(err).Msgf("failed to handle command: %v", command)
			}
		} else {
			log.Trace().Msg("No command")
		}
	}
}

func (s *Server) setHandlingDelayMsForTest(delay time.Duration) {
	s.handlingDelayMsForTest.Store(delay.Milliseconds())
}

func (s *Server) addDelayForTesting() {
	delay := s.handlingDelayMsForTest.Load()
	if delay > 0 {
		log.Trace().Msg("Sleeping...")
		time.Sleep(time.Duration(delay) * time.Millisecond)
		log.Trace().Msg("Awakening...")
	}
}
