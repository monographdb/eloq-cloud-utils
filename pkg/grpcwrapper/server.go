package grpcwrapper

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	// DefaultServerMaxMsgSize sets the limit for gRPC message transfer to 128MB for incoming/outgoing server messages.
	DefaultServerMaxMsgSize = 1024 * 1024 * 128
)

// NewGrpcServer creates a new gRPC server with unified production settings.
// Features:
//   - Max message sizes: Standardized at 128MB across all services.
//   - Keepalive Enforcement: Prevents client connections from becoming stale.
//   - Interceptors: Allows for easy injection of logging, metrics, etc.
func NewGrpcServer(unaryInterceptors []grpc.UnaryServerInterceptor, streamInterceptors []grpc.StreamServerInterceptor) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(DefaultServerMaxMsgSize),
		grpc.MaxSendMsgSize(DefaultServerMaxMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: 10 * time.Second,
		}),
	}

	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	return grpc.NewServer(opts...)
}
