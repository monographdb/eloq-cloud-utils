package grpcwrapper

import (
	"sync"

	"github.com/puzpuzpuz/xsync/v4"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	rpcInitOnce sync.Once
	instance    *RpcConnWrapper
)

type ClientConnInterface interface {
	grpc.ClientConnInterface
}

// RpcConnWrapper provides a thread-safe singleton pool for gRPC connections.
// It ensures that only one physical TCP connection is established per destination address.
type RpcConnWrapper struct {
	// rpcConnSet caches established connections using a high-performance concurrent map.
	// Key:   The destination server address (e.g., "127.0.0.1:8080" or "proxy-service.namespace:9090").
	// Value: The established gRPC *ClientConn.
	// Purpose: To ensure all components reuse the same physical connection for a given endpoint.
	rpcConnSet *xsync.MapOf[string, *grpc.ClientConn]

	// group ensures that concurrent dial requests to the same address are collapsed
	// into a single execution, preventing "thundering herd" or "cache miss storm".
	group singleflight.Group
}

func NewRpcConnWrapper() *RpcConnWrapper {
	rpcInitOnce.Do(func() {
		instance = &RpcConnWrapper{
			rpcConnSet: xsync.NewMapOf[string, *grpc.ClientConn](),
			group:      singleflight.Group{},
		}
	})
	return instance
}

// GetConn lazily creates or retrieves a cached gRPC connection for the given address.
// The 'serverAddr' parameter is the target host:port string which serves as the cache key.
func (g *RpcConnWrapper) GetConn(serverAddr string) (*grpc.ClientConn, bool) {
	// 1. Fast path: check if the connection already exists in the cache for this address.
	if val, ok := g.rpcConnSet.Load(serverAddr); ok {
		state := val.GetState()
		// If the connection is definitely dead or in a persistent failure state, remove and fall through.
		// Ready: OK; Idle: Lazy; Connecting: In progress.
		// TransientFailure: Potential broken/restarting server.
		// Shutdown: Definitely closed.
		if state != connectivity.Shutdown && state != connectivity.TransientFailure {
			return val, true
		}
		// Connection issues detected, remove it from the cache.
		g.rpcConnSet.Delete(serverAddr)
	}

	// 2. Slow path: coordinate multiple goroutines trying to connect to the same server simultaneously.
	// 'g.group.Do' will ensure that for the same 'serverAddr' key, only one callback runs at a time.
	actual, err, _ := g.group.Do(serverAddr, func() (interface{}, error) {
		// Double check after acquiring the singleflight lock to avoid race conditions.
		if checkVal, innerOk := g.rpcConnSet.Load(serverAddr); innerOk {
			state := checkVal.GetState()
			if state != connectivity.Shutdown && state != connectivity.TransientFailure {
				return checkVal, nil
			}
			g.rpcConnSet.Delete(serverAddr)
		}

		// Perform the actual gRPC dial.
		newConn, err := NewRpcConn(serverAddr)
		if err != nil {
			return nil, err
		}

		// Cache the successful connection for future use.
		g.rpcConnSet.Store(serverAddr, newConn)
		return newConn, nil
	})
	if err != nil {
		return nil, false
	}

	return actual.(*grpc.ClientConn), true
}

// Close closes all cached connections and clears the pool.
func (g *RpcConnWrapper) Close() {
	g.rpcConnSet.Range(func(key string, value *grpc.ClientConn) bool {
		_ = value.Close()
		return true
	})
	// Re-initialize the map to release references.
	g.rpcConnSet = xsync.NewMapOf[string, *grpc.ClientConn]()
}

// DeleteConn removes and returns a specific connection from the cache.
func (g *RpcConnWrapper) DeleteConn(serverAddr string) *grpc.ClientConn {
	val, ok := g.rpcConnSet.LoadAndDelete(serverAddr)
	if ok {
		return val
	}
	return nil
}
