package manager

import (
	"github.com/arkbriar/ss-mgr/manager/protocol"
	"google.golang.org/grpc"
)

type shadowsocksService struct {
	port     int    `json:"server_port"`
	password string `json:"password"`
}

// slaveMeta represents the meta information required by a local slave object.
type slaveMeta struct {
}

// Slave provides interfaces for managing the remote slave.
type Slave interface {
	// Dial opens the connection to the remote slave node.
	Dial() error
	// Close closes the connection.
	Close() error
	// Allocate adds services on the remote slave node.
	// The services added is returned.
	Allocate(srvs ...shadowsocksService) ([]shadowsocksService, error)
	// Free removes services on the remote slave node.
	// The services removed is returned
	Free(srvs ...shadowsocksService) ([]shadowsocksService, error)
	// ListServices gets all alive services.
	ListServices() ([]shadowsocksService, error)
	// GetStats gets the traffic statistics of all alive services.
	GetStats() (map[int]int64, error)
	// SetStats sets the traffic statistics of all alive services.
	SetStats(traffics map[int]int64) error
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
}

func NewSlave(url string) Slave {
	return &slave{
		remoteURL: url,
		conn:      nil,
		stub:      nil,
		token:     "",
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

func (s *slave) Allocate(srvs ...shadowsocksService) ([]shadowsocksService, error) {
	return nil, nil
}

func (s *slave) Free(srvs ...shadowsocksService) ([]shadowsocksService, error) {

	return nil, nil
}

func (s *slave) ListServices() ([]shadowsocksService, error) {

	return nil, nil
}

func (s *slave) GetStats() (map[int]int64, error) {

	return nil, nil
}

func (s *slave) SetStats(traffics map[int]int64) error {

	return nil
}

func (s *slave) Meta() slaveMeta {
	return s.meta
}
