package server

import (
	"bufio"
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerHandlesConnectionsConcurrently(t *testing.T) {
	s := startTestServer(t)
	hostPort, err := s.getAddressListeningOn()
	assert.Nil(t, err, "error in getting address listening on")

	done := make(chan struct{})

	go func() {
		defer close(done)
		// each makeTcpConnection takes just a little above 1s
		// if they run concurrently, it should take just a little above 1s to complete
		var wg sync.WaitGroup
		wg.Go(func() { makeTcpConnection(t, hostPort, 1) })
		wg.Go(func() { makeTcpConnection(t, hostPort, 2) })
		wg.Go(func() { makeTcpConnection(t, hostPort, 3) })
		wg.Wait()
	}()

	// set timeout to upper bound of time to complete the above goroutines
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	select {
	case <-done:
	case <-timeoutCtx.Done():
		assert.FailNow(t, "server took more time than expected")
	}
}

func makeTcpConnection(t *testing.T, hostPort string, index int) {
	conn, err := net.DialTimeout("tcp", hostPort, 1*time.Second)
	assert.Nilf(t, err, "[%d] error in connecting to redisgo", index)
	defer conn.Close()

	conn.Write([]byte("PING\r\n"))

	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	assert.Nilf(t, err, "[%d] error in reading response", index)
	assert.NotNil(t, message)

	time.Sleep(1 * time.Second)
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple PING PONG", "PING\r\n", "+PONG\r\n"},
		{"bulk string PING PONG", "*1\r\n$4\r\nPING\r\n", "+PONG\r\n"},
		{"COMMAND", "*1\r\n$7\r\nCOMMAND\r\n", "*0\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := startTestServer(t)
			hostPort, err := s.getAddressListeningOn()
			assert.Nil(t, err, "error in getting address listening on")

			conn, err := net.DialTimeout("tcp", hostPort, 1*time.Second)
			assert.Nil(t, err, "error in connecting to redisgo")
			defer conn.Close()

			conn.Write([]byte(tt.input))

			reader := bufio.NewReader(conn)
			message, err := reader.ReadString('\n')
			assert.Nil(t, err, "error in reading response")
			assert.Equal(t, tt.expected, message)
		})
	}
}

func startTestServer(t *testing.T) *Server {
	s := NewServer(":0")
	s.Start()
	t.Cleanup(func() {s.Stop()})
	return s
}