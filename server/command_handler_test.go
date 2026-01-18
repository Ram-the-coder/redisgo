package server

import (
	"net"
	"testing"

	"github.com/ram-the-coder/redisgo/internal"
	"github.com/ram-the-coder/redisgo/internal/resp"
	"github.com/stretchr/testify/assert"
)

func TestHandleReturnsErrorOnWriteFailure(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	clientConn.Close()

	store := internal.NewStore()
	cmd := &resp.Command{Name: "PING", Arguments: nil}

	err := Handle(cmd, serverConn, store)

	assert.NotNilf(t, err, "expected error when writing to closed connection, got nil")
}
