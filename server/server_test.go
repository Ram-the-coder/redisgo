package server

import (
	"context"
	"fmt"
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

	makeTcpConnection := func(t *testing.T, hostPort string, index int) {
		rdb := getRedisClient(t, hostPort)
		assert := assert.New(t)
		val, err := rdb.Ping(context.Background()).Result()
		assert.Nilf(err, "[%d] error in executing Ping", index)
		assert.Equalf("PONG", val, "[%d] unexpected response", index)
		time.Sleep(100 * time.Millisecond)
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
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	select {
	case <-done:
	case <-timeoutCtx.Done():
		assert.FailNow(t, "server took more time than expected")
	}
}

func TestSetAndGetKey(t *testing.T) {
	hostPort := startTestServer(t)
	rdb := getRedisClient(t, hostPort)

	val, err := rdb.Set(context.Background(), "testkey", "testvalue", 0*time.Second).Result()
	assert.Nil(t, err, "error in setting key")
	fmt.Printf("val: %s", val)

	val, err = rdb.Get(context.Background(), "testkey").Result()
	assert.Nil(t, err, "error in getting key")
	assert.Equal(t, "testvalue", val)
}

func TestIntValue(t *testing.T) {
	hostPort := startTestServer(t)
	rdb := getRedisClient(t, hostPort)

	val, err := rdb.Set(context.Background(), "testkey", 10, 0*time.Second).Result()
	assert.Nil(t, err, "error in setting key")
	fmt.Printf("val: %s", val)

	val, err = rdb.Get(context.Background(), "testkey").Result()
	assert.Nil(t, err, "error in getting key")
	assert.Equal(t, "10", val)
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
