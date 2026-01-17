package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/rs/zerolog/log"
)

type Server struct {
	address  string
	listener net.Listener
	stopCh   chan struct{}
}

func NewServer(address string) *Server {
	return &Server{address: address, stopCh: make(chan struct{})}
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
	messageCh := make(chan string)
	defer close(messageCh)
	go handleMessages(conn, messageCh)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if !(errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)) {
				log.Err(err).Msg("Error reading messsage")
			}
			return
		}
		log.Info().Msgf("Received message: %s", message)
		messageCh <- message[:len(message)-2]
	}
}

func handleMessages(conn net.Conn, messageCh <-chan string) {
	for message := range messageCh {
		if message == "PING" {
			log.Info().Msg("Got PING, replying with PONG")
			conn.Write([]byte(resp.SimpleString("PONG")))
			continue
		}

		index := 0
		if message[index] == '*' {
			index++
			arrayLength, err := strconv.Atoi(string(message[index]))
			if err != nil {
				log.Err(err).Msg("failed to read array length")
				continue
			}
			message = <-messageCh
			index = 0
			for i := 0; i < arrayLength; i++ {
				if message[index] == '$' {
					message = <-messageCh
					command := message
					if command == "COMMAND" {
						log.Info().Msg("Got COMMAND, replying with empty array")
						conn.Write([]byte(resp.Array(0)))
						continue
					}

					if command == "PING" {
						log.Info().Msg("Got PING, replying with PONG")
						conn.Write([]byte(resp.SimpleString("PONG")))
						continue
					}
				}
			}
		}
	}
}
