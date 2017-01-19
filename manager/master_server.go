package manager

import (
	"github.com/arkbriar/ss-mgr/manager/protocol"
)

// masterServer represents the master grpc service.
type masterServer struct {
	protocol.ShadowsocksManagerMasterServer
}
