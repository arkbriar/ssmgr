package manager

type shadowsocksService struct {
	int    port     `json:"server_port"`
	string password `json:"password"`
}

// slaveMeta represents the meta information required by a local slave object.
type slaveMeta struct {
}

// Slave provides interfaces for managing the remote slave.
type Slave interface {
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
	// Meta returns the local meta object of slave.
	Meta() slaveMeta
}

// slave is the true object of remote slave process. It implements the
// `Slave` interface.
type slave struct {
	remoteURL string    // remote slave's grpc service url
	token     string    // token used to communicate with remote slave
	meta      slaveMeta // meta store meta information such as services, etc.
}
