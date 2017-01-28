package shadowsocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"sync"
	"sync/atomic"

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

// Errors of `Manager`
var (
	ErrServerNotFound = errors.New("Server not found.")
	ErrInvalidServer  = errors.New("Invalid server.")
	ErrServerExists   = errors.New("Server already exists.")
)

// Server is a struct describes a shadowsocks server.
type Server struct {
	Host     string `json:"server"`
	Port     int32  `json:"server_port"`
	Password string `json:"password"`
	Method   string `json:"method"`
	Timeout  int    `json:"timeout"`
	Stat     atomic.Value
	options  serverOptions
	runtime  struct {
		path   string
		cmd    *exec.Cmd
		logw   io.WriteCloser
		config string
		cancel context.CancelFunc
	}
}

// Valid checks if it is a valid server configuration.
func (s *Server) Valid() bool {
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
	if len(s.runtime.config) != 0 {
		opts = []string{"-c " + s.runtime.config}
	} else {
		opts = []string{"-s " + s.Host, fmt.Sprintf("-p %d", s.Port), "-m " + s.Method, "-k " + s.Password, fmt.Sprintf("-d %d", s.Timeout)}
	}
	opts = append(opts, s.options.BuildArgs()...)
	return exec.CommandContext(ctx, "ss-server", opts...)
}

func (s *Server) clone() *Server {
	copy := *s
	copy.Stat.Store(s.GetStat())
	copy.runtime.config = ""
	copy.runtime.path = ""
	copy.runtime.cancel = nil
	return &copy
}

// GetStat returns the statistics of this server
func (s *Server) GetStat() Stat {
	return s.Stat.Load().(Stat)
}

// Process returns the running process / nil of server
func (s *Server) Process() *os.Process {
	return s.runtime.cmd.Process
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
	// Listen listens udp connection on 127.0.0.1:{udpPort} and handles the stats update
	// sent from ss-server.
	Listen() error
	// Add adds a ss-server with given arguments.
	Add(s *Server) error
	// Remove kills the ss-server if found.
	Remove(port int32) error
	// ListServers list the active ss-servers.
	ListServers() map[int32]*Server
	// GetServer gets a clone of `Server` struct of given port.
	GetServer(port int32) (*Server, error)
}

// Implementation of `Manager` interface.
type manager struct {
	serverLock sync.RWMutex
	servers    map[int32]*Server
	path       string
	udpPort    int
}

// NewManager returns a new manager.
func NewManager(udpPort int) Manager {
	return &manager{
		servers: make(map[int32]*Server),
		path:    path.Join(os.Getenv("HOME"), ".shadowsocks_manager"),
		udpPort: udpPort,
	}
}

func (mgr *manager) StatRecvHandler(data []byte) {
	cmd := string(data[:4])
	if string(data[:4]) != "stat:" {
		log.Warnf("Unrecognized command %s, dropped\n", cmd)
		return
	}
	body := bytes.TrimSpace(data[5:])
	var stat map[string]int64
	err := json.Unmarshal(body, stat)
	if err != nil {
		log.Warnln(err)
		return
	}
	port, traffic := -1, int64(-1)
	for portS, trafficS := range stat {
		port, _ = strconv.Atoi(portS)
		traffic = trafficS
		break
	}
	if port < 0 || traffic < 0 {
		log.Warnf("Invalid stat!\n")
		return
	}
	// Update statistic
	mgr.serverLock.RLock()
	defer mgr.serverLock.RUnlock()
	s, ok := mgr.servers[int32(port)]
	if !ok {
		log.Warnf("Server on port %d not found!\n", port)
		return
	}
	s.Stat.Store(Stat{Traffic: traffic})
}

func (mgr *manager) Listen() error {
	port := mgr.udpPort
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	go func() {
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			n, from, err := conn.ReadFromUDP(buf)
			log.Debugf("Receving packet from %s: %s\n", from, buf[:n])
			if err != nil {
				log.Warnln(err)
				continue
			}
			mgr.StatRecvHandler(buf[:n])
		}
	}()
	log.Infof("Listening on 127.0.0.1:%d ...\n", port)
	return nil
}

func (mgr *manager) prepareExec(s *Server) error {
	pathPrefix := path.Join(mgr.path, fmt.Sprint(s.Port))
	s.runtime.path = pathPrefix

	s.options.PidFile = path.Join(pathPrefix, "ss_server.pid")
	s.options.ManagerAddress = fmt.Sprintf("127.0.0.1:%d", mgr.udpPort)
	s.options.Verbose = true

	err := os.MkdirAll(pathPrefix, 0744)
	if err != nil {
		return err
	}
	configFile := path.Join(pathPrefix, "ss_server.json")
	err = s.Save(configFile)
	if err != nil {
		return err
	}
	s.runtime.config = configFile
	return nil
}

func (mgr *manager) deleteResidue(s *Server) error {
	err := os.RemoveAll(s.runtime.path)
	if err != nil {
		log.Warnf("Can not delete managed server path %s\n", s.runtime.path)
	}
	return err
}

func (mgr *manager) exec(s *Server) error {
	err := mgr.prepareExec(s)
	if err != nil {
		return err
	}
	logw, err := os.Open(path.Join(s.runtime.path, "ss_server.log"))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	cmd := s.Command(ctx)
	cmd.Stdout, cmd.Stderr = logw, logw
	if err := cmd.Start(); err != nil {
		return err
	}
	s.runtime.cancel = cancel
	s.runtime.logw = logw
	s.runtime.cmd = cmd
	log.Infof("ss-server running at process %d\n", cmd.Process.Pid)
	return nil
}

func (mgr *manager) kill(s *Server) {
	s.runtime.cancel()
	s.runtime.logw.Close()
	mgr.deleteResidue(s)
}

func (mgr *manager) Add(s *Server) error {
	mgr.serverLock.Lock()
	defer mgr.serverLock.Unlock()
	if _, ok := mgr.servers[s.Port]; ok {
		return ErrServerExists
	}
	if !s.Valid() {
		return ErrInvalidServer
	}
	err := mgr.exec(s)
	if err != nil {
		return err
	}
	mgr.servers[s.Port] = s
	return nil
}

func (mgr *manager) Remove(port int32) error {
	mgr.serverLock.Lock()
	defer mgr.serverLock.Unlock()
	s, ok := mgr.servers[port]
	if !ok {
		return ErrServerNotFound
	}
	delete(mgr.servers, port)
	mgr.kill(s)
	return nil
}

func (mgr *manager) ListServers() map[int32]*Server {
	mgr.serverLock.RLock()
	defer mgr.serverLock.RUnlock()
	currentServers := make(map[int32]*Server)
	for port, s := range mgr.servers {
		currentServers[port] = s.clone()
	}
	return currentServers
}

func (mgr *manager) GetServer(port int32) (*Server, error) {
	mgr.serverLock.RLock()
	defer mgr.serverLock.RUnlock()
	s, ok := mgr.servers[port]
	if !ok {
		return nil, ErrServerNotFound
	}
	return s.clone(), nil
}
