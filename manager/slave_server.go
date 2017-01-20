package manager

import (
	"errors"
	"sync"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/arkbriar/ss-mgr/manager/protocol"
	"github.com/arkbriar/ss-mgr/shadowsocks"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

var (
	poolFullErr = errors.New("Pool is full.")
)

type portPool struct {
	lock      sync.Mutex
	lower     int32
	size      int32
	last      int32
	allocated int32
	exist     func(int32) bool
	// blacklist map[int32]struct{}
}

func newPortPool(lower, size int32, exist func(int32) bool) *portPool {
	if size <= 0 {
		panic("Size must be positive.")
	}
	return &portPool{
		lower:     lower,
		size:      size,
		last:      -1,
		allocated: 0,
		exist:     exist,
	}
}

func (pool *portPool) Allocate() int32 {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if pool.allocated == pool.size {
		return -1
	}
	this := pool.last + 1
	if pool.last == -1 {
		this = pool.lower
	}
	for {
		if !pool.exist(this) {
			pool.allocated++
			pool.last = this
		}
		this++
		if this >= pool.lower+pool.size {
			this -= pool.size
		}
	}
}

func (pool *portPool) Free(port int32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()
	if pool.allocated == 0 {
		return
	}
	pool.allocated--
}

type slaveServer struct {
	srvsLock  sync.RWMutex
	srvs      map[int32]*shadowsocksService
	statsLock sync.RWMutex
	stats     map[int32]int64
	manager   shadowsocks.Manager
	portPool  *portPool
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
		srvs:    make(map[int32]*shadowsocksService),
		stats:   make(map[int32]int64),
		manager: manager,
	}
	s.portPool = newPortPool(20000, 3000, func(port int32) bool {
		_, ok := s.srvs[port]
		return ok
	})
	return s
}

func (s *slaveServer) initializeServicesStatistics(srvs ...*shadowsocksService) {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()
	for _, srv := range srvs {
		s.stats[srv.Port] = 0
	}
}

func (s *slaveServer) allocatePort() (int32, error) {
	port := s.portPool.Allocate()
	if port == -1 {
		return -1, poolFullErr
	}
	return port, nil
}

func (s *slaveServer) Allocate(ctx context.Context, r *protocol.AllocateRequest) (*protocol.AllocateResponse, error) {
	reqServices := r.GetServiceList().GetServices()
	if len(reqServices) == 0 {
		return nil, nil
	}
	s.srvsLock.Lock()
	defer s.srvsLock.Unlock()
	allocatedServices := make([]*shadowsocksService, 0, len(reqServices))
	for _, srv := range reqServices {
		newPort, err := s.allocatePort()
		if err != nil {
			if err == poolFullErr {
				break
			}
		}
		err = s.manager.Add(newPort, srv.GetPassword())
		if err != nil {
			logrus.Errorf("Allocating service %d:%s failed, %s\n", srv.GetPort(), srv.GetPassword(), err)
			continue
		}
		allocatedServices = append(allocatedServices, &shadowsocksService{
			UserId:   srv.GetUserId(),
			Port:     newPort,
			Password: srv.GetPassword(),
		})
	}
	for _, alsrv := range allocatedServices {
		s.srvs[alsrv.Port] = alsrv
	}
	s.initializeServicesStatistics(allocatedServices...)
	return &protocol.AllocateResponse{
		ServiceList: constructProtocolServiceList(allocatedServices...),
	}, nil
}

func (s *slaveServer) freePort(port int32) {
	s.portPool.Free(port)
}

func (s *slaveServer) freeServiceStatistics(srvs ...*shadowsocksService) {
	s.statsLock.Lock()
	defer s.statsLock.Unlock()
	for _, srv := range srvs {
		delete(s.stats, srv.Port)
	}
}

func (s *slaveServer) Free(ctx context.Context, r *protocol.FreeRequest) (*protocol.FreeResponse, error) {
	reqServices := r.GetServiceList().GetServices()
	if len(reqServices) == 0 {
		return nil, nil
	}
	s.srvsLock.Lock()
	defer s.srvsLock.Unlock()
	freedServices := make([]*shadowsocksService, 0, len(reqServices))
	for _, srv := range reqServices {
		if err := s.manager.Remove(srv.GetPort()); err != nil {
			logrus.Errorf("Removing service on port %d faild, %s\n", srv.GetPort(), err)
			continue
		}
		freedServices = append(freedServices, constructShadowsocksServiceFromProtocolService(srv))
	}
	for _, srv := range freedServices {
		s.freePort(srv.Port)
		delete(s.srvs, srv.Port)
	}
	s.freeServiceStatistics(freedServices...)
	return nil, nil
}

func (s *slaveServer) ListServices(ctx context.Context, _ *google_protobuf.Empty) (*protocol.ServiceList, error) {
	return nil, nil
}

func (s *slaveServer) GetStats(ctx context.Context, _ *google_protobuf.Empty) (*protocol.Statistics, error) {
	return nil, nil
}

func (s *slaveServer) GetStatsStream(_ *google_protobuf.Empty, ss protocol.ShadowsocksManagerSlave_GetStatsStreamServer) error {
	return nil
}

func (s *slaveServer) SetStats(ctx context.Context, stats *protocol.Statistics) (*google_protobuf.Empty, error) {
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
