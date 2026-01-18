package server

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
