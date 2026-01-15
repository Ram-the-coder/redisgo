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
	s := NewServer(":0")
	s.Start()
	hostPort, err := s.getAddressListeningOn()
	assert.Nil(t, err, "error in getting address listening on")

	done := make(chan struct{})

	go func() {
		defer close(done)
		// each makeTcpConnection takes just a little above 1s
		// if they run concurrently, it should take just a little above 1s to complete
		var wg sync.WaitGroup
		wg.Go(func() {makeTcpConnection(t, hostPort, 1)})
		wg.Go(func() {makeTcpConnection(t, hostPort, 2)})
		wg.Go(func() {makeTcpConnection(t, hostPort, 3)})
		wg.Wait()
	}()

	// set timeout to upper bound of time to complete the above goroutines
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	select {
	case <- done:
	case <- timeoutCtx.Done():
		assert.FailNow(t, "server took more time than expected")
	}
	
}

func makeTcpConnection(t *testing.T, hostPort string, index int) {
	conn, err := net.DialTimeout("tcp", hostPort, 1 * time.Second)
	assert.Nilf(t, err, "[%d] error in connecting to redisgo", index)
	defer conn.Close()

	conn.Write([]byte("HELLO\r\n"))

	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	assert.Nilf(t, err, "[%d] error in reading response", index)
	assert.NotNil(t, message)	
	
	time.Sleep(1 * time.Second)
}