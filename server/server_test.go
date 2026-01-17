package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
	"github.com/stretchr/testify/assert"
)

func TestServerHandlesConnectionsConcurrently(t *testing.T) {
	hostPort := startTestServer(t)

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
	rdb := getRedisClient(t, hostPort)
	assert := assert.New(t)
	val, err := rdb.Ping(context.Background()).Result()
	assert.Nilf(err, "[%d] error in executing Ping", index)
	assert.Equalf("PONG", val, "[%d] unexpected response", index)
	time.Sleep(1 * time.Second)
}

func startTestServer(t *testing.T) string {
	s := NewServer(":0")
	s.Start()
	t.Cleanup(func() { s.Stop() })

	hostPort, err := s.getAddressListeningOn()
	assert.Nil(t, err, "error in getting address listening on")

	return hostPort
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
