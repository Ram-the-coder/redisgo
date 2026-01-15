package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

type Server struct {
	address string
	listener net.Listener
	activeConnections sync.WaitGroup
	stopCh chan struct{}
}

func NewServer(address string) *Server {
	return &Server{address: address, stopCh: make(chan struct{})}
}

func (s *Server) getAddressListeningOn() (string, error) {
	if (s.listener == nil) {
		return "", errors.New("Cannot get listener address when server is not running")
	}
	return s.listener.Addr().String(), nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.address)
	if err != nil {
		fmt.Printf("Error listening: %s\n", err.Error())
		os.Exit(1)
	}
	s.listener = ln
	addressListeningOn, _ := s.getAddressListeningOn()
	fmt.Printf("Redisgo server started and listening on %s\n", addressListeningOn)

	go s.acceptConnectionLoop()
	return nil
}

func (s *Server) Stop() {
	fmt.Println("Stopping server...")
	close(s.stopCh) // Stop waiting for new connections on the listener
	s.listener.Close() // Stop listening on the port
	os.Exit(0)
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
				fmt.Printf("Error accepting connection: %s\n", err.Error())
				continue
			}
		}
		s.activeConnections.Go(func() {
			s.handleConnection(conn)
		})
	}
}


func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if !(errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)) {
				fmt.Printf("Error reading: %s\n", err.Error())
			}
			return
		}
		fmt.Printf("Received message: %s\n", message)
		conn.Write([]byte("OK\r\n"))
	}
}
