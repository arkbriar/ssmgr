package manager

import (
	"context"
	"errors"
	"fmt"

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
}

func (m *slaveMeta) addPorts(srvs ...*shadowsocksService) {
	for _, srv := range srvs {
		m.openedPorts[srv.Port] = srv
	}
}

func (m *slaveMeta) removePorts(srvs ...*shadowsocksService) {
	for _, srv := range srvs {
		delete(m.openedPorts, srv.Port)
	}
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
	// The services removed is returned
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
	meta      slaveMeta                              // meta store meta information such as services, etc.
	Slave
}

func NewSlave(url string) Slave {
	return &slave{
		remoteURL: url,
		conn:      nil,
		stub:      nil,
		token:     "",
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
	resp, err := s.stub.Allocate(context.Background(), &protocol.AllocateRequest{
		ServiceList: serviceList,
	})
	if err != nil {
		return nil, err
	}
	diff := compareLists(serviceList, resp.GetServiceList())
	allocatedList := constructServiceList(resp.GetServiceList())
	if len(diff) != 0 {
		return allocatedList, constructErrorFromDifferenceServiceList(diff)
	}
	return allocatedList, nil
}

func (s *slave) Free(srvs ...*shadowsocksService) ([]*shadowsocksService, error) {
	serviceList := constructProtocolServiceList(srvs...)
	resp, err := s.stub.Free(context.Background(), &protocol.FreeRequest{
		ServiceList: serviceList,
	})
	if err != nil {
		return nil, err
	}
	diff := compareLists(serviceList, resp.GetServiceList())
	freedList := constructServiceList(resp.GetServiceList())
	if len(diff) != 0 {
		return freedList, constructErrorFromDifferenceServiceList(diff)
	}
	return freedList, nil
}

func (s *slave) ListServices() ([]*shadowsocksService, error) {
	resp, err := s.stub.ListServices(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return constructServiceList(resp), nil
}

func (s *slave) GetStats() (map[int32]int64, error) {
	resp, err := s.stub.GetStats(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	return resp.GetTraffics(), nil
}

func (s *slave) SetStats(traffics map[int32]int64) error {
	_, err := s.stub.SetStats(context.Background(), &protocol.Statistics{
		Traffics: traffics,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *slave) Meta() slaveMeta {
	return s.meta
}
