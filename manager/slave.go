package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/arkbriar/ss-mgr/manager/protocol"
	"google.golang.org/grpc"
)

type shadowsocksService struct {
	UserId   string `json:"user_id"`
	Port     int32  `json:"server_port"`
	Password string `json:"password"`
}

// slaveMeta represents the meta information required by a local slave object.
type slaveMeta struct {
	openedPorts map[int32]*shadowsocksService
	stats       map[int32]int64
}

func (m *slaveMeta) addPorts(srvs ...*shadowsocksService) {
	for _, srv := range srvs {
		m.openedPorts[srv.Port] = srv
		m.stats[srv.Port] = 0
	}
}

func (m *slaveMeta) removePorts(srvs ...*shadowsocksService) {
	for _, srv := range srvs {
		delete(m.openedPorts, srv.Port)
		delete(m.stats, srv.Port)
	}
}

func (m *slaveMeta) setStats(stats map[int32]int64) {
	for port, traffic := range stats {
		m.stats[port] = traffic
	}
}

func (m *slaveMeta) ListServices() []*shadowsocksService {
	srvs := make([]*shadowsocksService, 0, len(m.openedPorts))
	for _, srv := range m.openedPorts {
		srvs = append(srvs, srv)
	}
	return srvs
}

func (m *slaveMeta) GetStats() map[int32]int64 {
	return m.stats
}

// Slave provides interfaces for managing the remote slave.
type Slave interface {
	// Dial opens the connection to the remote slave node.
	Dial() error
	// Close closes the connection.
	Close() error
	// Allocate adds services on the remote slave node.
	// The services added is returned.
	Allocate(srvs ...*shadowsocksService) ([]*shadowsocksService, error)
	// Free removes services on the remote slave node.
	// The services removed is returned.
	Free(srvs ...*shadowsocksService) ([]*shadowsocksService, error)
	// ListServices gets all alive services.
	ListServices() ([]*shadowsocksService, error)
	// GetStats gets the traffic statistics of all alive services.
	GetStats() (map[int32]int64, error)
	// SetStats sets the traffic statistics of all alive services.
	SetStats(traffics map[int32]int64) error
	// Meta returns a copy of local meta object of slave.
	Meta() slaveMeta
}

// slave is the true object of remote slave process. It implements the
// `Slave` interface.
type slave struct {
	remoteURL string                                 // remote slave's grpc service url
	conn      *grpc.ClientConn                       // grpc client connection
	stub      protocol.ShadowsocksManagerSlaveClient // remote slave's grpc service client
	token     string                                 // token used to communicate with remote slave
	ctx       context.Context                        // context for grpc communication
	meta      slaveMeta                              // meta store meta information such as services, etc.
	Slave
}

func NewSlave(url, token string) Slave {
	return &slave{
		remoteURL: url,
		conn:      nil,
		stub:      nil,
		token:     token,
		ctx:       context.WithValue(context.Background(), "TOKEN", token),
		meta: slaveMeta{
			openedPorts: make(map[int32]*shadowsocksService),
		},
	}
}

func (s *slave) isTokenInvalid() bool {
	return len(s.token) == 0
}

func (s *slave) Dial() error {
	// FIXME(arkbriar@gmail.com) Here I initialize the connection using `grpc.WithInsecure`.
	conn, err := grpc.Dial(s.remoteURL, grpc.WithInsecure())
	if err != nil {
		return err
	}
	s.conn = conn
	s.stub = protocol.NewShadowsocksManagerSlaveClient(conn)
	return nil
}

func (s *slave) Close() error {
	conn := s.conn
	s.conn, s.stub = nil, nil
	return conn.Close()
}

func constructProtocolServiceList(srvs ...*shadowsocksService) *protocol.ServiceList {
	services := make([]*protocol.ShadowsocksService, 0, len(srvs))
	for _, srv := range srvs {
		services = append(services, &protocol.ShadowsocksService{
			Port:     srv.Port,
			Password: srv.Password,
		})
	}
	return &protocol.ServiceList{
		Services: services,
	}
}

func constructServiceList(srvList *protocol.ServiceList) []*shadowsocksService {
	srvs := make([]*shadowsocksService, 0, len(srvList.GetServices()))
	for _, pbsrv := range srvList.GetServices() {
		srvs = append(srvs, &shadowsocksService{
			UserId:   pbsrv.GetUserId(),
			Port:     pbsrv.GetPort(),
			Password: pbsrv.GetPassword(),
		})
	}
	return srvs
}

func compareLists(required, current *protocol.ServiceList) (diff []*shadowsocksService) {
	diff = make([]*shadowsocksService, 0, 1)
	for _, a := range required.GetServices() {
		for _, b := range current.GetServices() {
			if a.GetUserId() == b.GetUserId() && a.GetPassword() == b.GetPassword() {
				break
			}
		}
		diff = append(diff, &shadowsocksService{
			UserId:   a.GetUserId(),
			Port:     -1,
			Password: a.GetPassword(),
		})
	}
	return diff
}

func constructErrorFromDifferenceServiceList(diff []*shadowsocksService) error {
	if diff == nil || len(diff) == 0 {
		return nil
	}
	var errMsg string
	if len(diff) == 1 {
		errMsg = fmt.Sprintf("There is 1 service not allocated (user: password):")
	} else {
		errMsg = fmt.Sprintf("There are %d services not allocated (user: password):")
	}
	for _, srv := range diff {
		errMsg += fmt.Sprintf("\n  %s: %s", srv.UserId, srv.Password)
	}
	return errors.New(errMsg)
}

func (s *slave) Allocate(srvs ...*shadowsocksService) ([]*shadowsocksService, error) {
	serviceList := constructProtocolServiceList(srvs...)
	resp, err := s.stub.Allocate(s.ctx, &protocol.AllocateRequest{
		ServiceList: serviceList,
	})
	if err != nil {
		return nil, err
	}
	diff := compareLists(serviceList, resp.GetServiceList())
	allocatedList := constructServiceList(resp.GetServiceList())
	s.meta.addPorts(allocatedList...)
	if len(diff) != 0 {
		return allocatedList, constructErrorFromDifferenceServiceList(diff)
	}
	return allocatedList, nil
}

func (s *slave) Free(srvs ...*shadowsocksService) ([]*shadowsocksService, error) {
	serviceList := constructProtocolServiceList(srvs...)
	resp, err := s.stub.Free(s.ctx, &protocol.FreeRequest{
		ServiceList: serviceList,
	})
	if err != nil {
		return nil, err
	}
	diff := compareLists(serviceList, resp.GetServiceList())
	freedList := constructServiceList(resp.GetServiceList())
	s.meta.removePorts(freedList...)
	if len(diff) != 0 {
		return freedList, constructErrorFromDifferenceServiceList(diff)
	}
	return freedList, nil
}

func (s *slave) ListServices() ([]*shadowsocksService, error) {
	resp, err := s.stub.ListServices(s.ctx, nil)
	if err != nil {
		return nil, err
	}
	// Compare the returned list with those recorded.
	diff := compareLists(constructProtocolServiceList(s.meta.ListServices()...), resp)
	if len(diff) != 0 {
		logrus.Warnln(constructErrorFromDifferenceServiceList(diff))
	}
	return constructServiceList(resp), nil
}

func (s *slave) GetStats() (map[int32]int64, error) {
	resp, err := s.stub.GetStats(s.ctx, nil)
	if err != nil {
		return nil, err
	}
	s.meta.setStats(resp.GetTraffics())
	return resp.GetTraffics(), nil
}

func (s *slave) SetStats(traffics map[int32]int64) error {
	_, err := s.stub.SetStats(s.ctx, &protocol.Statistics{
		Traffics: traffics,
	})
	if err != nil {
		return err
	}
	s.meta.setStats(traffics)
	return nil
}

func (s *slave) Meta() slaveMeta {
	return s.meta
}
