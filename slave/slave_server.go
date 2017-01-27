package manager

import (
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/arkbriar/ss-mgr/protocol"
	"github.com/arkbriar/ss-mgr/slave/shadowsocks"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type slaveServer struct {
	srvsLock sync.RWMutex
	srvs     map[int32]*ShadowsocksService
	manager  shadowsocks.Manager
}

// newSlaveGRPCServer creates a new server instance and tries to connect to local
// ss-manager.
// Errors are returned when there's somthing wrong with your ss-manager service.
func newSlaveGRPCServer(managerURL string) (protocol.ShadowsocksManagerSlaveServer, error) {
	mgr := shadowsocks.NewManager(managerURL)
	err := mgr.Dial()
	if err != nil {
		return nil, err
	}
	return newSlaveGRPCServerWithActiveManager(mgr), nil
}

func newSlaveGRPCServerWithActiveManager(manager shadowsocks.Manager) protocol.ShadowsocksManagerSlaveServer {
	s := &slaveServer{
		srvs:    make(map[int32]*ShadowsocksService),
		manager: manager,
	}
	return s
}

func (s *slaveServer) Allocate(ctx context.Context, r *protocol.AllocateRequest) (*protocol.AllocateResponse, error) {
	reqServices := r.GetServiceList().GetServices()
	if len(reqServices) == 0 {
		return nil, nil
	}
	s.srvsLock.Lock()
	defer s.srvsLock.Unlock()
	allocatedServices := make([]*ShadowsocksService, 0, len(reqServices))
	for _, srv := range reqServices {
		err := s.manager.Add(srv.GetPort(), srv.GetPassword())
		if err != nil {
			logrus.Errorf("Allocating service %d:%s failed, %s\n", srv.GetPort(), srv.GetPassword(), err)
			continue
		}
		allocatedServices = append(allocatedServices, constructShadowsocksServiceFromProtocolService(srv))
	}
	for _, alsrv := range allocatedServices {
		s.srvs[alsrv.Port] = alsrv
	}
	return &protocol.AllocateResponse{
		ServiceList: constructProtocolServiceList(allocatedServices...),
	}, nil
}

func (s *slaveServer) Free(ctx context.Context, r *protocol.FreeRequest) (*protocol.FreeResponse, error) {
	reqServices := r.GetServiceList().GetServices()
	if len(reqServices) == 0 {
		return nil, nil
	}
	s.srvsLock.Lock()
	defer s.srvsLock.Unlock()
	freedServices := make([]*ShadowsocksService, 0, len(reqServices))
	for _, srv := range reqServices {
		if err := s.manager.Remove(srv.GetPort()); err != nil {
			logrus.Errorf("Removing service on port %d faild, %s\n", srv.GetPort(), err)
			continue
		}
		freedServices = append(freedServices, constructShadowsocksServiceFromProtocolService(srv))
	}
	for _, srv := range freedServices {
		delete(s.srvs, srv.Port)
	}
	return nil, nil
}

func (s *slaveServer) listServices() []*ShadowsocksService {
	s.srvsLock.RLock()
	defer s.srvsLock.RUnlock()
	srvSlice := make([]*ShadowsocksService, 0, len(s.srvs))
	for _, srv := range s.srvs {
		srvSlice = append(srvSlice, srv)
	}
	return srvSlice
}

func (s *slaveServer) ListServices(ctx context.Context, _ *google_protobuf.Empty) (*protocol.ServiceList, error) {
	s.srvsLock.RLock()
	defer s.srvsLock.RUnlock()
	return constructProtocolServiceList(s.listServices()...), nil
}

func (s *slaveServer) GetStats(ctx context.Context, _ *google_protobuf.Empty) (*protocol.Statistics, error) {
	return &protocol.Statistics{
		Traffics: copyStats(s.manager.GetStats()),
	}, nil
}

func (s *slaveServer) GetStatsStream(_ *google_protobuf.Empty, ss protocol.ShadowsocksManagerSlave_GetStatsStreamServer) error {
	close := make(chan struct{}, 1)
	go func() {
		ss.RecvMsg(nil)
		close <- struct{}{}
	}()
	next := make(chan struct{}, 1)
	next <- struct{}{}
	const limit = 5
	errTimes := 0
StreamLoop:
	for {
		select {
		case <-next:
			err := ss.Send(&protocol.Statistics{Traffics: copyStats(s.manager.GetStats())})
			if err != nil {
				errTimes++
				logrus.Errorf("Stats stream sent failed, %s\n", err)
			} else {
				errTimes = 0
			}
			if errTimes >= limit {
				logrus.Errorf("Stats stream has encountered 5 continous errors, closing")
				break StreamLoop
			}
		case <-close:
			logrus.Debugln("Stats stream is closing.")
			break StreamLoop
		}
		time.Sleep(5 * time.Second)
		next <- struct{}{}
	}
	return nil
}

func (s *slaveServer) SetStats(ctx context.Context, stats *protocol.Statistics) (*google_protobuf.Empty, error) {
	err := s.manager.SetStats(stats.GetTraffics())
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// RegisterSlaveServer creates a new slave grpc server and registers itself to grpc services.
// It will panic if there are something wrong when creating slave grpc server.
func RegisterSlaveServer(s *grpc.Server, managerURL string) {
	slaveServer, err := newSlaveGRPCServer(managerURL)
	if err != nil {
		logrus.Panicln(err)
	}
	protocol.RegisterShadowsocksManagerSlaveServer(s, slaveServer)
}

// RegisterSlaveServerWithActiveManager creates a new slave grpc server the same as
// `RegisterSlaveServer` except it needs an active(dialed) shadowsocks manager.
func RegisterSlaveWithActiveManager(s *grpc.Server, manager shadowsocks.Manager) {
	slaveServer := newSlaveGRPCServerWithActiveManager(manager)
	protocol.RegisterShadowsocksManagerSlaveServer(s, slaveServer)
}
