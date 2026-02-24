package grpcwrapper

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	// DefaultServiceMaxMsgSize sets the limit for gRPC message transfer to 64MB.
	// This is crucial for components like the Scraper that pull large amounts of traffic data.
	DefaultServiceMaxMsgSize = 1024 * 1024 * 64

	// DefaultClientRetryPolicy defines the standard retry behavior using gRPC's ServiceConfig.
	// It implements exponential backoff and retries on UNAVAILABLE/INTERNAL errors,
	// providing "transparent retries" for the business layer.
	DefaultClientRetryPolicy = `
{
	"methodConfig": [
		{
			"name": [],
			"waitForReady": true,
			"retryPolicy": {
				"MaxAttempts": 5,
				"InitialBackoff": "0.1s",
				"MaxBackoff": "0.5s",
				"BackoffMultiplier": 2.0,
				"RetryableStatusCodes": ["UNAVAILABLE", "INTERNAL"]
			}
		}
	],
	"loadBalancingConfig": [
		{"round_robin": {}},
		{"weighted_round_robin": {}},
		{"pick_first": {"shuffleAddressList": true}}
	]
}`
)

// NewRpcConnWithOptions creates a new gRPC connection with production-ready settings.
// Features:
//   - Keepalive: Sends pings to keep the connection alive through idle firewalls/LBs.
//   - Retry: Built-in transparent retry via ServiceConfig.
//   - Message Size: Increased limits for large data synchronization.
//   - Interceptors: Extensible for logging, metrics, or recovery via parameters.
func NewRpcConnWithOptions(addr string, retryPolicy string, maxMsgSize int, unaryInterceptors []grpc.UnaryClientInterceptor, streamInterceptors []grpc.StreamClientInterceptor) (*grpc.ClientConn, error) {
	if maxMsgSize <= 0 {
		maxMsgSize = DefaultServiceMaxMsgSize
	}
	if retryPolicy == "" {
		retryPolicy = DefaultClientRetryPolicy
	}

	callOption := grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxMsgSize),
		grpc.MaxCallSendMsgSize(maxMsgSize),
	)

	connectParams := grpc.WithConnectParams(grpc.ConnectParams{
		MinConnectTimeout: time.Second * 3,
	})

	keepaliveParams := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                time.Second * 10,
		Timeout:             time.Second * 3,
		PermitWithoutStream: true,
	})

	dialOptions := []grpc.DialOption{
		callOption,
		connectParams,
		keepaliveParams,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(retryPolicy),
	}

	for _, interceptor := range unaryInterceptors {
		dialOptions = append(dialOptions, grpc.WithUnaryInterceptor(interceptor))
	}
	for _, interceptor := range streamInterceptors {
		dialOptions = append(dialOptions, grpc.WithStreamInterceptor(interceptor))
	}

	return grpc.NewClient(addr, dialOptions...)
}

// NewRpcConn creates a new gRPC connection with default settings.
func NewRpcConn(addr string) (*grpc.ClientConn, error) {
	return NewRpcConnWithOptions(addr, "", 0, nil, nil)
}
