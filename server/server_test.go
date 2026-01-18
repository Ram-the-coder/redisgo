package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/stretchr/testify/assert"
)

func TestServerHandlesConnectionsConcurrently(t *testing.T) {
	_, hostPort := startTestServer(t)
	done := make(chan struct{})

	var (
		currentConnections atomic.Int32
		maxConnections     atomic.Int32
	)
	updateMax := func(current int32) {
		for {
			oldMax := maxConnections.Load()
			if current <= oldMax {
				return
			}
			// try to set new max, if someone else changed it, loop and retry
			if maxConnections.CompareAndSwap(oldMax, current) {
				return
			}
		}
	}

	makeTcpConnection := func(t *testing.T, hostPort string, index int) {
		rdb := getRedisClient(t, hostPort)
		assert := assert.New(t)
		val, err := rdb.Ping(context.Background()).Result()
		assert.Nilf(err, "[%d] error in executing Ping", index)
		assert.Equalf("PONG", val, "[%d] unexpected response", index)
		current := currentConnections.Add(1)
		updateMax(current)
		// keep connection open briefly so other connections can overlap
		time.Sleep(100 * time.Millisecond)
		currentConnections.Add(-1)
	}

	go func() {
		defer close(done)
		// each makeTcpConnection takes just a little above 100ms
		// if they run concurrently, it should take just a little above 100ms to complete
		var wg sync.WaitGroup
		wg.Go(func() { makeTcpConnection(t, hostPort, 1) })
		wg.Go(func() { makeTcpConnection(t, hostPort, 2) })
		wg.Go(func() { makeTcpConnection(t, hostPort, 3) })
		wg.Wait()
	}()

	// set timeout to upper bound of time to complete the above goroutines
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	select {
	case <-done:
		// If the calls happened concurrently, maxConnections should be 3
		assert.Equal(t, int32(3), maxConnections.Load())
	case <-timeoutCtx.Done():
		assert.FailNow(t, "server took more time than expected")
	}
}

func TestSetAndGetKey(t *testing.T) {
	_, hostPort := startTestServer(t)
	rdb := getRedisClient(t, hostPort)

	val, err := rdb.Set(context.Background(), "testkey", "testvalue", 0*time.Second).Result()
	assert.Nil(t, err, "error in setting key")
	t.Logf("val: %s", val)

	val, err = rdb.Get(context.Background(), "testkey").Result()
	assert.Nil(t, err, "error in getting key")
	assert.Equal(t, "testvalue", val)

	time.Sleep(1 * time.Second)
}

func TestIntValue(t *testing.T) {
	_, hostPort := startTestServer(t)
	rdb := getRedisClient(t, hostPort)

	val, err := rdb.Set(context.Background(), "testkey", 10, 0*time.Second).Result()
	assert.Nil(t, err, "error in setting key")
	t.Logf("val: %s", val)

	val, err = rdb.Get(context.Background(), "testkey").Result()
	assert.Nil(t, err, "error in getting key")
	assert.Equal(t, "10", val)
}

func TestConcurrentCacheAccess(t *testing.T) {
	_, hostPort := startTestServer(t)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", id)
				value := fmt.Sprintf("val-%d-%d", id, j)
				rdb := getRedisClient(t, hostPort)
				rdb.Set(context.Background(), key, value, 0)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key-%d", id)
				rdb := getRedisClient(t, hostPort)
				rdb.Get(context.Background(), key)
			}
		}(i)
	}
	wg.Wait()
}

func TestPipelineSetAndGet(t *testing.T) {
	_, hostPort := startTestServer(t)
	conn, err := net.DialTimeout("tcp", hostPort, 1*time.Second)
	assert.Nilf(t, err, "failed to connect to server: %f", err)
	t.Cleanup(func() { conn.Close() })

	// SET foo bar, then GET foo
	setCmd := concatCommands("*3", "$3", "SET", "$3", "foo", "$3", "bar")
	getCmd := concatCommands("*2", "$3", "GET", "$3", "foo")
	_, err = conn.Write([]byte(setCmd + getCmd))
	assert.Nilf(t, err, "failed to write: %f", err)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	reader := bufio.NewReader(conn)

	// Read SET response
	resp1, err := reader.ReadString('\n')
	assert.Nilf(t, err, "failed to read SET response: %f", err)
	assert.Equal(t, "+OK\r\n", resp1)

	// Read GET response
	resp2, err := reader.ReadString('\n')
	assert.Nilf(t, err, "failed to read GET response (line 1): %f", err)
	assert.Equal(t, "$3\r\n", resp2)

	resp3, err := reader.ReadString('\n')
	assert.Nilf(t, err, "failed to read GET response (line 2): %f", err)
	assert.Equal(t, "bar\r\n", resp3)
}

func TestMalformedInputDoesNotPanic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty array", "*0\r\n"},
		{"array with non-bulk-string command", concatCommands("*1", ":123")},
		{"negative array length", "*-1\r\n"},
		{"truncated bulk string", concatCommands("$10", "hello")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, hostPort := startTestServer(t)
			conn, err := net.DialTimeout("tcp", hostPort, 1*time.Second)
			assert.Nilf(t, err, "failed to connect to server: %f", err)
			t.Cleanup(func() { conn.Close() })

			_, err = conn.Write([]byte(tt.input))
			assert.Nilf(t, err, "failed to write: %f", err)

			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			reader := bufio.NewReader(conn)
			_, err = reader.ReadString('\n')

			verifyServerIsAlive(t, hostPort)
		})
	}
}

func TestVerifyServerAliveWhenClientClosesConnBeforeServerResponds(t *testing.T) {
	s, hostPort := startTestServer(t)
	conn, err := net.DialTimeout("tcp", hostPort, 1*time.Second)
	assert.Nilf(t, err, "failed to connect to server: %f", err)

	s.setHandlingDelayMsForTest(250 * time.Millisecond)
	conn.Write([]byte(concatCommands("*1", "$4", "ping")))
	conn.Close()

	s.setHandlingDelayMsForTest(0)
	verifyServerIsAlive(t, hostPort)
}

func startTestServer(t *testing.T) (*Server, string) {
	s := NewServer(":0")
	s.Start()
	t.Cleanup(func() { s.Stop() })

	hostPort, err := s.getAddressListeningOn()
	assert.Nil(t, err, "error in getting address listening on")

	return s, hostPort
}

func getRedisClient(t *testing.T, hostPort string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		// Addr: "localhost:6379",
		Addr:     hostPort,
		Password: "",
		DB:       0,
		MaintNotificationsConfig: &maintnotifications.Config{
			Mode: maintnotifications.ModeDisabled,
		},
		DisableIdentity: true,
	})
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func concatCommands(commands ...string) string {
	var sb strings.Builder
	for _, command := range commands {
		sb.WriteString(command)
		sb.WriteString("\r\n")
	}
	return sb.String()
}

func verifyServerIsAlive(t *testing.T, hostPort string) {
	// Verify the server didn't crash
	rdb := getRedisClient(t, hostPort)
	_, err := rdb.Ping(context.Background()).Result()
	assert.Nil(t, err, "server is down")
}
