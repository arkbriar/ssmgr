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

// NewSSMgrSlaveServer creates a SSMgrSlaveServer.
func NewSSMgrSlaveServer(token string, mgr ss.Manager) proto.SSMgrSlaveServer {
	return &slaveServer{
		token: token,
		mgr:   mgr,
	}
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

func StreamAuthInterceptor(token string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authorize(stream.Context(), token); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

func UnaryAuthInterceptor(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := authorize(ctx, token); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (s *slaveServer) Allocate(ctx context.Context, r *proto.AllocateRequest) (*google_protobuf.Empty, error) {
	ss := &ss.Server{
		Host:     "0.0.0.0",
		Port:     r.GetPort(),
		Password: r.GetPassword(),
		Method:   r.GetMethod(),
		Timeout:  60,
	}
	return &google_protobuf.Empty{}, s.mgr.Add(ss)
}

func (s *slaveServer) Free(ctx context.Context, r *proto.FreeRequest) (*google_protobuf.Empty, error) {
	return &google_protobuf.Empty{}, s.mgr.Remove(r.GetPort())
}

func (s *slaveServer) GetStats(ctx context.Context, _ *google_protobuf.Empty) (*proto.Statistics, error) {
	flow := make(map[int32]*proto.FlowUnit)
	for port, ss := range s.mgr.ListServers() {
		flow[port] = &proto.FlowUnit{
			Traffic:   ss.GetStat().Traffic,
			StartTime: ss.Extra.StartTime.UnixNano(),
		}
	}
	return &proto.Statistics{
		Flow: flow,
	}, nil
}
