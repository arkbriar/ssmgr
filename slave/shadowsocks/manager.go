package shadowsocks

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type serverOptions struct {
	MPTCP          bool
	TCPFastOpen    bool
	Auth           bool
	NameServer     string
	PidFile        string
	ManagerAddress string
	Interface      string
	FireWall       bool
	Verbose        bool
	ConfigFile     string
}

func (o *serverOptions) BuildArgs() []string {
	opts := make([]string, 0)
	if o.MPTCP {
		opts = append(opts, "--mptcp")
	}
	if o.TCPFastOpen {
		opts = append(opts, "--fast-open")
	}
	if o.Auth {
		opts = append(opts, "-A")
	}
	if len(o.NameServer) != 0 {
		opts = append(opts, "-d "+o.NameServer)
	}
	if len(o.PidFile) != 0 {
		opts = append(opts, "-f "+o.PidFile)
	}
	if len(o.ManagerAddress) != 0 {
		opts = append(opts, "--manager-address "+o.ManagerAddress)
	}
	if o.FireWall {
		opts = append(opts, "--firewall")
	}
	if o.Verbose {
		opts = append(opts, "-v")
	}
	return opts
}

var (
	methods = []string{
		"table", "rc4", "rc4-md5", "aes-128-cfb", "aes-192-cfb", "aes-256-cfb",
		"aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "bf-cfb", "camellia-128-cfb",
		"camellia-192-cfb", "camellia-256-cfb", "cast5-cfb", "des-cfb", "idea-cfb",
		"rc2-cfb", "seed-cfb", "salsa20", "chacha20", "chacha20-ietf",
	}
)

// ValidateEncryptMethod validates if the encrypt method is supported.
func ValidateEncryptMethod(m string) bool {
	for _, method := range methods {
		if m == method {
			return true
		}
	}
	return false
}

// Server is a struct describes a shadowsocks server.
type Server struct {
	Host     string `json:"server"`
	Port     int32  `json:"server_port"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Timeout  int    `json:"timeout"`
	Stat     *Stat
	options  *serverOptions
	cancel   context.CancelFunc
	path     string
}

// Validate checks if it is a valid server configuration.
func (s *Server) Validate() bool {
	return len(s.Host) != 0 && s.Port > 0 && s.Port < 65536 && len(s.Password) >= 8 && ValidateEncryptMethod(s.Method) && s.Timeout > 0
}

// Save saves this server's configuration to file in JSON.
func (s *Server) Save(filename string) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Command constructs a new shadowsock server command
func (s *Server) Command(ctx context.Context) *exec.Cmd {
	var opts []string
	if len(s.options.ConfigFile) != 0 {
		opts = []string{"-c " + s.options.ConfigFile}
	} else {
		opts = []string{"-s " + s.Host, fmt.Sprintf("-p %d", s.Port), "-m " + s.Method, "-k " + s.Password, fmt.Sprintf("-d %d", s.Timeout)}
	}
	if s.options != nil {
		opts = append(opts, s.options.BuildArgs()...)
	}
	return exec.CommandContext(ctx, "ss-server", opts...)
}

// Stat represents the statistics collected from a shadowsocks server
type Stat struct {
	Traffic int64 `json:"traffic"` // Transfered traffic in bytes
	/* Rx      int64 `json:"rx"`      // Receive in bytes
	 * Tx      int64 `json:"tx"`      // Transmit in bytes */
}

// Manager is an interface provides a few methods to manager shadowsocks
// servers.
type Manager interface {
	Add(s *Server) error
	Remove(port int32) error
	ListServers() ([]Server, error)
	GetServer(port int32) (Server, error)
}

var (
	home = os.Getenv("HOME")
)

// Implementation of `Manager` interface.
type manager struct {
	serverLock sync.RWMutex
	servers    map[int32]*Server
	path       string
}

// NewManager returns a new manager.
func NewManager() Manager {
	return &manager{
		servers: make(map[int32]*Server),
		path:    path.Join(home, ".shadowsocks_manager"),
	}
}

func (mgr *manager) updateStat(port int32, stat *Stat) {
	mgr.serverLock.Lock()
	defer mgr.serverLock.Unlock()
	s, ok := mgr.servers[port]
	if !ok {
		log.Warnf("There's no server listening on port %d\n", port)
		return
	}
	*s.Stat = *stat
}

func (mgr *manager) prepareExec(s *Server) error {
	s.path = path.Join(mgr.path, fmt.Sprint(s.Port))

	s.options.PidFile = path.Join(s.path, "ss_server.pid")
	s.options.ManagerAddress = "localhost"
	s.options.Verbose = true

	err := os.MkdirAll(s.path, 0644)
	if err != nil {
		return err
	}
	configFile := path.Join(s.path, "ss_server.json")
	err = s.Save(configFile)
	if err != nil {
		return err
	}
	s.options.ConfigFile = configFile
	return nil
}

func (mgr *manager) deleteResidue(s *Server) error {
	err := os.RemoveAll(s.path)
	return err
}

func (mgr *manager) exec(s *Server) error {
	logWriter, err := os.Open(path.Join(s.path, "ss_server.log"))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	cmd := s.Command(ctx)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	return cmd.Start()
}

func (mgr *manager) kill(s *Server) {
	err := mgr.deleteResidue(s)
	if err != nil {
		log.Warnf("Can not delete managed server path %s\n", s.path)
	}
	s.cancel()
}

func (mgr *manager) Add(s *Server) error {
	return nil
}

func (mgr *manager) Remove(port int32) error {
	return nil
}

func (mgr *manager) ListServers() ([]Server, error) {
	return nil, nil
}

func (mgr *manager) GetServer(port int32) (Server, error) {
	return Server{}, nil
}
