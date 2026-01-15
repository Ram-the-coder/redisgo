package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"sync"
)

type Server struct {
	port string
	listener net.Listener
	activeConnections sync.WaitGroup
	stopCh chan struct{}
}

func NewServer(port string) *Server {
	return &Server{port: port, stopCh: make(chan struct{})}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.port)
	if err != nil {
		fmt.Printf("Error listening: %s\n", err.Error())
		os.Exit(1)
	}
	s.listener = ln
	fmt.Printf("Redisgo server started and listening on port %s\n", s.port)

	go s.acceptConnectionLoop()
	return nil
}

func (s *Server) Stop() error {
	fmt.Println("Stopping server...")
	close(s.stopCh)
	s.listener.Close()
	s.activeConnections.Wait()
	fmt.Println("Redisgo server successfully stopped")
	return nil
}

func (s *Server) acceptConnectionLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopCh:
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
			fmt.Printf("Error reading: %s\n", err.Error())
			return
		}
		fmt.Printf("Received message: %s\n", message)
	}
}
