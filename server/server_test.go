package server

import (
	"bufio"
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/stretchr/testify/assert"
)

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
