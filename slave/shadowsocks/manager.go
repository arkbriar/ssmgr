package shadowsocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/arkbriar/ss-mgr/slave/shadowsocks/process"
)

type serverOptions struct {
	UDPRelay       bool
	IPv6First      bool
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
	if o.UDPRelay {
		opts = append(opts, "-u")
	}
	if o.IPv6First {
		opts = append(opts, "-6")
	}
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
		opts = append(opts, "-d", o.NameServer)
	}
	// DO NOT USE THIS OPTION
	// When use pid file, ss-server will create a child process and we can
	// not operate on it directly.
	/* if len(o.PidFile) != 0 {
	 *     opts = append(opts, "-f", o.PidFile)
	 * } */
	if len(o.ManagerAddress) != 0 {
		opts = append(opts, "--manager-address", o.ManagerAddress)
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
	stat     atomic.Value
	options  serverOptions
	path     string
	runtime  struct {
		proc   *os.Process
		config string
	}
}

func validPort(p int32) bool {
	return p > 0 && p < 65536
}

// Valid checks if it is a valid server configuration.
func (s *Server) Valid() bool {
	return len(s.Host) != 0 && validPort(s.Port) && len(s.Password) >= 8 && ValidateEncryptMethod(s.Method) && s.Timeout > 0
}

// Save saves this server's configuration to file in JSON.
func (s *Server) Save(filename string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) restoreProc() error {
	pidname, err := ioutil.ReadFile(path.Join(s.path, "ss_server.pid"))
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(string(pidname))
	if err != nil {
		log.Warnf("Invalid pid file, content: %s", pidname)
		return err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Warnf("Can not find process(%d)", pid)
		return err
	}
	s.runtime.proc = proc
	return nil
}

// Restore restores the configuration and runtime infos.
func (s *Server) Restore(serverPath string) error {
	s.path = serverPath
	s.runtime.config = path.Join(s.path, "ss_server.conf")

	data, err := ioutil.ReadFile(s.runtime.config)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, s)
	if err != nil {
		return err
	}
	err = s.restoreProc()
	if err != nil {
		log.Warn(err)
	}
	return nil
}

// SavePidFile saves current pid to a file (s.options.PidFile). This method
// is to replace ss-server's '-f' option.
func (s *Server) SavePidFile() error {
	proc := s.Process()
	if len(s.options.PidFile) != 0 && proc != nil {
		return ioutil.WriteFile(s.options.PidFile, []byte(fmt.Sprint(proc.Pid)), 0644)
	}
	return nil
}

func (s *Server) opts() []string {
	var opts []string
	if len(s.runtime.config) != 0 {
		opts = []string{"-c", s.runtime.config}
	} else {
		opts = []string{"-s", s.Host, "-p", fmt.Sprint(s.Port), "-m", s.Method, "-k", s.Password, "-d", fmt.Sprint(s.Timeout)}
	}
	opts = append(opts, s.options.BuildArgs()...)
	return opts
}

// Command constructs a new shadowsock server command
func (s *Server) Command() *exec.Cmd {
	return exec.Command("ss-server", s.opts()...)
}

// String returns the command line string
func (s *Server) String() string {
	return fmt.Sprintf("ss-server %s", strings.Join(s.opts(), " "))
}

func (s *Server) clone() *Server {
	copy := *s
	return &copy
}

// GetStat returns the statistics of this server
func (s *Server) GetStat() Stat {
	stat := s.stat.Load()
	if stat == nil {
		return Stat{}
	}
	return stat.(Stat)
}

// Process returns the running process / nil of server
func (s *Server) Process() *os.Process {
	return s.runtime.proc
}

// Alive returns if the server is alive
func (s *Server) Alive() bool {
	proc := s.Process()
	return proc != nil && process.Alive(proc.Pid)
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
	Listen(ctx context.Context) error
	// WatchDaemon is a daemon that watches all server processes and rises dead servers.
	WatchDaemon(ctx context.Context, notifyFail func(*Server))
	// Add adds a ss-server with given arguments.
	Add(s *Server) error
	// Remove kills the ss-server if found.
	Remove(port int32) error
	// ListServers list the active ss-servers.
	ListServers() map[int32]*Server
	// GetServer gets a clone of `Server` struct of given port.
	GetServer(port int32) (*Server, error)
	// Restore all stopped servers, this must be called before any other actions.
	Restore() error
	// CleanUp removes all servers and files.
	CleanUp()
}

// Implementation of `Manager` interface.
type manager struct {
	serverMu sync.RWMutex
	servers  map[int32]*Server
	path     string
	udpPort  int
	execLock sync.Mutex
}

// NewManager returns a new manager.
func NewManager(udpPort int) Manager {
	return &manager{
		servers: make(map[int32]*Server),
		path:    path.Join(os.Getenv("HOME"), ".shadowsocks_manager"),
		udpPort: udpPort,
	}
}

// lockExec gets the command executing lock.
func (mgr *manager) lockExec() {
	mgr.execLock.Lock()
}

func (mgr *manager) unlockExec() {
	mgr.execLock.Unlock()
}

func (mgr *manager) handleStat(data []byte) {
	cmd := string(data[:4])
	if string(data[:4]) != "stat" {
		log.Warnf("Unrecognized command %s, dropped", cmd)
		return
	}
	body := bytes.TrimSpace(data[5:])
	var stat map[string]int64
	err := json.Unmarshal(body, &stat)
	if err != nil {
		log.Warnln("Unmarshal error:", err)
		return
	}
	port, traffic := -1, int64(-1)
	for portS, trafficS := range stat {
		port, _ = strconv.Atoi(portS)
		traffic = trafficS
		break
	}
	if port < 0 || traffic < 0 {
		log.Warnf("Invalid stat!")
		return
	}
	// Update statistic
	mgr.serverMu.RLock()
	defer mgr.serverMu.RUnlock()
	s, ok := mgr.servers[int32(port)]
	if !ok {
		log.Warnf("Server on port %d not found!", port)
		return
	}
	s.stat.Store(Stat{Traffic: traffic})
}

func (mgr *manager) Listen(ctx context.Context) error {
	port := mgr.udpPort
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return errors.New("Canceled.")
	default:
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	go func() {
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				n, from, err := conn.ReadFromUDP(buf)
				// the n-th is \x00 to indicate end
				log.Debugf("Receving packet from %s: %s", from, buf[:n-1])
				if err != nil {
					log.Warnln(err)
					continue
				}
				mgr.handleStat(buf[:n-1])
			}
		}
	}()
	log.Infof("Listening on 127.0.0.1:%d ...", port)
	return nil
}

func (mgr *manager) fillOptions(s *Server) {
	s.options.UDPRelay = true
	s.options.PidFile = path.Join(s.path, "ss_server.pid")
	s.options.ManagerAddress = fmt.Sprintf("127.0.0.1:%d", mgr.udpPort)
	s.options.Verbose = true
}

func (mgr *manager) prepareExec(s *Server) error {
	s.path = path.Join(mgr.path, fmt.Sprint(s.Port))
	mgr.fillOptions(s)

	err := os.MkdirAll(s.path, 0744)
	if err != nil {
		return err
	}
	config := path.Join(s.path, "ss_server.conf")
	err = s.Save(config)
	if err != nil {
		return err
	}
	s.runtime.config = config
	return nil
}

func (mgr *manager) deleteResidue(s *Server) error {
	err := os.RemoveAll(s.path)
	if err != nil {
		log.Warnf("Can not delete managed server path %s", s.path)
	}
	return err
}

func (mgr *manager) exec(s *Server) error {
	mgr.lockExec()
	defer mgr.unlockExec()
	err := mgr.prepareExec(s)
	if err != nil {
		return err
	}
	logw, err := os.Create(path.Join(s.path, "ss_server.log"))
	if err != nil {
		return err
	}
	cmd := s.Command()
	cmd.Stdout, cmd.Stderr = logw, logw
	if err := cmd.Start(); err != nil {
		return err
	}
	s.runtime.proc = cmd.Process
	if err := s.SavePidFile(); err != nil {
		log.Warnf("Can not save pid file, %s", err)
	}
	log.Infof("ss-server running at process %d", cmd.Process.Pid)
	return nil
}

func (mgr *manager) kill(s *Server) {
	mgr.lockExec()
	defer mgr.unlockExec()
	if err := s.Process().Kill(); err != nil {
		log.Warnln(err)
	}
	// release process's resource
	s.runtime.proc.Wait()
	mgr.deleteResidue(s)
}

func (mgr *manager) add(s *Server) error {
	mgr.serverMu.Lock()
	defer mgr.serverMu.Unlock()
	if _, ok := mgr.servers[s.Port]; ok {
		return ErrServerExists
	}
	mgr.servers[s.Port] = s
	return nil
}

func (mgr *manager) Add(s *Server) error {
	mgr.serverMu.Lock()
	defer mgr.serverMu.Unlock()
	if _, ok := mgr.servers[s.Port]; ok {
		return ErrServerExists
	}
	if !s.Valid() {
		return ErrInvalidServer
	}
	s = s.clone()
	err := mgr.exec(s)
	if err != nil {
		return err
	}
	mgr.servers[s.Port] = s
	log.Debugf("Adding server: %s", s)
	return nil
}

func (mgr *manager) remove(port int32) {
	mgr.serverMu.Lock()
	defer mgr.serverMu.Unlock()
	delete(mgr.servers, port)
}

func (mgr *manager) Remove(port int32) error {
	mgr.serverMu.Lock()
	defer mgr.serverMu.Unlock()
	s, ok := mgr.servers[port]
	if !ok {
		return ErrServerNotFound
	}
	delete(mgr.servers, port)
	mgr.kill(s)
	log.Debugf("Removing server: %s", s)
	return nil
}

func (mgr *manager) ListServers() map[int32]*Server {
	mgr.serverMu.RLock()
	defer mgr.serverMu.RUnlock()
	currentServers := make(map[int32]*Server)
	for port, s := range mgr.servers {
		currentServers[port] = s.clone()
	}
	return currentServers
}

func (mgr *manager) GetServer(port int32) (*Server, error) {
	mgr.serverMu.RLock()
	defer mgr.serverMu.RUnlock()
	s, ok := mgr.servers[port]
	if !ok {
		return nil, ErrServerNotFound
	}
	return s.clone(), nil
}

func (mgr *manager) riseDaemon(ctx context.Context, s *Server, notifyFail func(*Server)) {
	if s.Alive() {
		return
	}
	// If the server failes to restart for 10 times, manager will give up.
	const upperLimit = 10
	count := 0
	for count < upperLimit {
		select {
		case <-ctx.Done():
			log.Warn("Cancel rise.")
			return
		case <-time.After(100 * time.Millisecond):
			if err := mgr.exec(s); err != nil {
				log.Warnf("Can not restart server(%d), %s", s.Port, err)
				count++
			} else {
				log.Infof("Server(%d) back to work.", s.Port)
				return
			}
		}
	}
	// Fail to start the server
	log.Warnf("Deleting server(%d)", s.Port)
	mgr.deleteResidue(s)
	mgr.remove(s.Port)
	if notifyFail != nil {
		notifyFail(s.clone())
	}
}

// WatchDaemon provides a way to monitor all server processes
func (mgr *manager) WatchDaemon(ctx context.Context, notifyFail func(*Server)) {
	rising := make(map[int32]context.CancelFunc)
	for {
		select {
		case <-ctx.Done():
			go func() {
				for _, cancel := range rising {
					cancel()
				}
			}()
			return
		case <-time.After(5 * time.Second):
			for _, s := range mgr.ListServers() {
				if !s.Alive() {
					if _, ok := rising[s.Port]; ok {
						continue
					}
					ctx, cancel := context.WithCancel(context.Background())
					rising[s.Port] = cancel
					log.Warnf("Server on port %d should be alive, rising it", s.Port)
					go mgr.riseDaemon(ctx, s, notifyFail)
				} else {
					delete(rising, s.Port)
				}
			}
		}
	}
}

func isDir(dirname string) bool {
	fileInfo, err := os.Stat(dirname)
	return err == nil && fileInfo.IsDir()
}

func getPort(portname string) (int32, bool) {
	p, err := strconv.Atoi(portname)
	return int32(p), err == nil && validPort(int32(p))
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func (mgr *manager) restore(s *Server, serverPath string) error {
	if !validPort(s.Port) {
		log.Warn("Invalid port.")
	}
	s.path = serverPath
	err := s.Restore(serverPath)
	if err != nil {
		return err
	}
	if s.Alive() {
		mgr.fillOptions(s)
		if err := mgr.add(s); err != nil {
			return err
		}
		return nil
	}
	err = mgr.Add(s)
	if err != nil {
		return err
	}
	return nil
}

// Restore starts all ss-servers that leaves their dirs in manager.path
func (mgr *manager) Restore() error {
	if !isDir(mgr.path) {
		return errors.New(mgr.path + " is not a directory.")
	}
	names, err := readDirNames(mgr.path)
	if err != nil {
		return err
	}
	for _, name := range names {
		serverPath := path.Join(mgr.path, name)
		if isDir(serverPath) {
			if port, ok := getPort(name); ok {
				log.Infof("Restoring server(%d) ...", port)
				err := mgr.restore(&Server{Port: port}, serverPath)
				if err != nil {
					log.Warnf("Can not restore server(%d), remove it. %s", port, err)
					os.RemoveAll(serverPath)
				}
			} else {
				log.Warnf("Ignore unrecognized port %s.", name)
			}
		} else {
			log.Warnf("Ignore normal file %s.", serverPath)
		}
	}
	return nil
}

func (mgr *manager) CleanUp() {
	os.RemoveAll(mgr.path)
	for p := range mgr.ListServers() {
		mgr.Remove(p)
	}
}
