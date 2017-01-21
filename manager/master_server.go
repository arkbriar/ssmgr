package manager

import (
	"golang.org/x/net/context"

	"github.com/arkbriar/ss-mgr/manager/protocol"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

// masterServer represents the master grpc service.
type masterServer struct {
	slaveList map[uuid.UUID]Slave
	protocol.ShadowsocksManagerMasterServer
}

// newMasterServer creates a new master server
func newMasterServer() protocol.ShadowsocksManagerMasterServer {
	return &masterServer{
		slaveList: make(map[uuid.UUID]Slave),
	}
}

func (s *masterServer) Register(ctx context.Context, r *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	slaveId := uuid.NewV4()
	s.slaveList[slaveId] = NewSlave(r.GetSelfUrl(), slaveId.String())
	return &protocol.RegisterResponse{
		Token: slaveId.String(),
	}, nil
}

// RegisterMasterServer creates a new master server and registers it to grpc server
func RegisterMasterServer(s *grpc.Server) {
	protocol.RegisterShadowsocksManagerMasterServer(s, newMasterServer())
}
