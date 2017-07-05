package slave

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	proto "github.com/arkbriar/ssmgr/protocol"
	ss "github.com/arkbriar/ssmgr/slave/shadowsocks"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type server struct {
	proto.SSMgrSlaveServer

	token string
	mgr   ss.Manager
}

// NewSSMgrSlaveServer creates a SSMgrSlaveServer.
func NewSSMgrSlaveServer(token string, mgr ss.Manager) proto.SSMgrSlaveServer {
	return &server{
		token: token,
		mgr:   mgr,
	}
}

func authorize(ctx context.Context, token string) error {
	if md, ok := metadata.FromContext(ctx); ok {
		if len(md["token"]) > 0 && md["token"][0] == token {
			return nil
		}
		return errors.New("access denied")
	}
	return errors.New("empty metadata")
}

// StreamAuthInterceptor returns an interceptor to do authorization for grpc stream call.
func StreamAuthInterceptor(token string) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := authorize(stream.Context(), token); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}

// UnaryAuthInterceptor returns an interceptor to do authorization for grpc unary call.
func UnaryAuthInterceptor(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := authorize(ctx, token); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (s *server) Allocate(ctx context.Context, r *proto.AllocateRequest) (*google_protobuf.Empty, error) {
	server := &ss.Server{
		Host:     "0.0.0.0",
		Port:     r.GetPort(),
		Password: r.GetPassword(),
		Method:   r.GetMethod(),
		Timeout:  60,
	}

	log.Debugf("Recv allocate request: %v", r)

	return &google_protobuf.Empty{}, s.mgr.Add(server)
}

func (s *server) Free(ctx context.Context, r *proto.FreeRequest) (*google_protobuf.Empty, error) {
	log.Debugf("Recv free request: %v", r)

	return &google_protobuf.Empty{}, s.mgr.Remove(r.GetPort())
}

func (s *server) GetStats(ctx context.Context, _ *google_protobuf.Empty) (*proto.Statistics, error) {
	log.Debugf("Recv get stat request")

	flow := make(map[int32]*proto.FlowUnit)
	for port, server := range s.mgr.ListServers() {
		flow[port] = &proto.FlowUnit{
			Traffic:   server.GetStat().Traffic,
			StartTime: server.Extra.StartTime.UnixNano(),
		}
	}

	log.Debugf("Stats now: %v", flow)

	return &proto.Statistics{
		Flow: flow,
	}, nil
}
