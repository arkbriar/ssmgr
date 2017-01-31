package slave

import (
	"errors"

	proto "github.com/arkbriar/ss-mgr/protocol"
	ss "github.com/arkbriar/ss-mgr/slave/shadowsocks"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type slaveServer struct {
	proto.SSMgrSlaveServer

	token string
	mgr   ss.Manager
}

// NewSSMgrSlaveServer creates a SSMgrSlaveServer along with two auth interceptors.
func NewSSMgrSlaveServer(token string, mgr ss.Manager) (proto.SSMgrSlaveServer, grpc.StreamServerInterceptor, grpc.UnaryServerInterceptor) {
	return &slaveServer{
		token: token,
		mgr:   mgr,
	}, streamAuthInterceptor(token), unaryAuthInterceptor(token)
}

func authorize(ctx context.Context, token string) error {
	if md, ok := metadata.FromContext(ctx); ok {
		if len(md["token"]) > 0 && md["token"][0] == token {
			return nil
		}
		return errors.New("Access denied.")
	}
	return errors.New("Empty metadata.")
}

func streamAuthInterceptor(token string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authorize(stream.Context(), token); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

func unaryAuthInterceptor(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := authorize(ctx, token); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (s *slaveServer) Allocate(ctx context.Context, r *proto.AllocateRequest) (*google_protobuf.Empty, error) {
	return nil, nil
}

func (s *slaveServer) Free(ctx context.Context, r *proto.FreeRequest) (*google_protobuf.Empty, error) {
	return nil, nil
}

func (s *slaveServer) GetStats(ctx context.Context, _ *google_protobuf.Empty) (*proto.Statistics, error) {
	return nil, nil
}
