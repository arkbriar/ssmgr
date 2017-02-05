package shadowsocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"sort"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// Errors of `Manager`
var (
	ErrServerNotFound = errors.New("server not found")
	ErrInvalidServer  = errors.New("invalid server")
	ErrServerExists   = errors.New("server already exists")
)

// Manager is an interface provides a few methods to manager shadowsocks
// servers.
type Manager interface {
	// Listen listens udp connection on 127.0.0.1:{udpPort} and handles the stats update
	// sent from ss-server.
	Listen(ctx context.Context) error
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
}

// NewManager returns a new manager.
func NewManager(udpPort int) Manager {
	mgr := &manager{
		servers: make(map[int32]*Server),
		path:    path.Join(os.Getenv("HOME"), ".ssmgr"),
		udpPort: udpPort,
	}
	return mgr
}

func (mgr *manager) handleStat(data []byte) {
	cmd := string(data[:4])
	if string(data[:4]) != "stat" {
		log.Warnf("Unrecognized command %s, dropped", cmd)
		return
	}

	var stat map[string]int64

	body := bytes.TrimSpace(data[5:])
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

	// update statistic
	mgr.serverMu.RLock()
	defer mgr.serverMu.RUnlock()

	s, ok := mgr.servers[int32(port)]
	if !ok {
		log.Warnf("Server on port %d not found!", port)
		return
	}
	s.updateStat(Stat{Traffic: traffic})
}

func (mgr *manager) managerAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", mgr.udpPort)
}

func (mgr *manager) Listen(ctx context.Context) error {
	port := mgr.udpPort
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return errors.New("canceled")
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
				if err != nil {
					log.Warnln(err)
					continue
				}
				data := bytes.Trim(buf[:n-1], "\x00\r\n")

				log.Debugf("Receving packet from %s: %s", from, data)

				mgr.handleStat(data)
			}
		}
	}()

	log.Debugf("Listening on 127.0.0.1:%d", port)

	return nil
}

func (mgr *manager) addAlive(s *Server) error {
	mgr.serverMu.Lock()
	defer mgr.serverMu.Unlock()

	if _, ok := mgr.servers[s.Port]; ok {
		return ErrServerExists
	}
	mgr.servers[s.Port] = s
	return nil
}

func (mgr *manager) prepareServer(s *Server) *Server {
	runPath := path.Join(mgr.path, fmt.Sprint(s.Port))
	s = s.clone().WithDefaults().WithRunPath(runPath).WithPidFile(
		path.Join(runPath, "ss_server.pid"),
	).WithManagerAddress(mgr.managerAddress())
	return s
}

func (mgr *manager) Add(s *Server) error {
	if !s.valid() {
		return ErrInvalidServer
	}

	s = mgr.prepareServer(s)
	if err := os.MkdirAll(s.runPath, 0744); err != nil {
		return err
	}

	mgr.serverMu.Lock()
	defer mgr.serverMu.Unlock()

	if _, ok := mgr.servers[s.Port]; ok {
		return ErrServerExists
	}
	if err := s.Start(); err != nil {
		return err
	}
	mgr.servers[s.Port] = s

	log.Infof("Add server(%s)", s)

	return nil
}

func (mgr *manager) Remove(port int32) error {
	mgr.serverMu.Lock()
	defer mgr.serverMu.Unlock()

	s, ok := mgr.servers[port]
	if !ok {
		return ErrServerNotFound
	}

	delete(mgr.servers, port)
	if err := s.Stop(); err != nil {
		log.Warn(err)
	}
	os.RemoveAll(s.runPath)

	log.Info("Remove server(%s)", s)

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

	s = mgr.prepareServer(s)
	err := s.Restore(serverPath)
	if err != nil {
		return err
	}

	if s.Alive() {
		if err := mgr.addAlive(s); err != nil {
			return err
		}
		return nil
	}

	// when server process is dead
	err = mgr.Add(s)
	if err != nil {
		return err
	}

	log.Info("Server(%s) restored.", s)

	return nil
}

// Restore starts all ss-servers that leaves their dirs in manager.path
func (mgr *manager) Restore() error {
	if _, err := os.Stat(mgr.path); err != nil {
		return errors.New(mgr.path + " doesn't not exsits")
	}
	if !isDir(mgr.path) {
		return errors.New(mgr.path + " is not a directory")
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
					log.Warnf("Can not restore server(%d), %s, remove it.", port, err)
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
	names, err := readDirNames(mgr.path)
	if err != nil {
		log.Warn(err)
	}
	for _, name := range names {
		os.RemoveAll(path.Join(mgr.path, name))
	}
	for p := range mgr.ListServers() {
		mgr.Remove(p)
	}

	log.Infof("Clean up all managed servers.")
}
