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
	mgr := &manager{
		servers: make(map[int32]*Server),
		path:    path.Join(os.Getenv("HOME"), ".shadowsocks_manager"),
		udpPort: udpPort,
	}
	return mgr
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
	_, ok := mgr.servers[int32(port)]
	if !ok {
		log.Warnf("Server on port %d not found!", port)
		return
	}
	/* s.stat.Store(Stat{Traffic: traffic}) */
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

/* func (mgr *manager) deleteResidue(s *Server) error {
 *     err := os.RemoveAll(s.path)
 *     if err != nil {
 *         log.Warnf("Can not delete managed server path %s", s.path)
 *     }
 *     return err
 * } */

/* func (mgr *manager) exec(s *Server) error {
 *     mgr.lockExec()
 *     defer mgr.unlockExec()
 *     s.Extra.StartTime = time.Now()
 *     err := mgr.prepareExec(s)
 *     if err != nil {
 *         return err
 *     }
 *     logw, err := os.Create(path.Join(s.path, "ss_server.log"))
 *     if err != nil {
 *         return err
 *     }
 *     cmd := s.Command()
 *     cmd.Stdout, cmd.Stderr = logw, logw
 *     if err := cmd.Start(); err != nil {
 *         return err
 *     }
 *     s.runtime.proc = cmd.Process
 *     if err := s.SavePidFile(); err != nil {
 *         log.Warnf("Can not save pid file, %s", err)
 *     }
 *     log.Infof("ss-server running at process %d", cmd.Process.Pid)
 *     mgr.applyOptionsOnExec(s, mgr.opts)
 *     return nil
 * }
 *
 * func (mgr *manager) kill(s *Server) {
 *     mgr.lockExec()
 *     defer mgr.unlockExec()
 *     if s.Alive() {
 *         if err := s.Process().Kill(); err != nil {
 *             log.Warnln(err)
 *         }
 *         // release process's resource
 *         s.runtime.proc.Wait()
 *     }
 *     mgr.deleteResidue(s)
 *     mgr.clearOptionsOnKill(s, mgr.opts)
 * } */

/* func (mgr *manager) add(s *Server) error {
 *     mgr.serverMu.Lock()
 *     defer mgr.serverMu.Unlock()
 *     if _, ok := mgr.servers[s.Port]; ok {
 *         return ErrServerExists
 *     }
 *     mgr.servers[s.Port] = s
 *     return nil
 * } */

func (mgr *manager) Add(s *Server) error {
	return nil
	/* mgr.serverMu.Lock()
	 * defer mgr.serverMu.Unlock()
	 * if _, ok := mgr.servers[s.Port]; ok {
	 *     return ErrServerExists
	 * }
	 * if !s.Valid() {
	 *     return ErrInvalidServer
	 * }
	 * s = s.clone()
	 * err := mgr.exec(s)
	 * if err != nil {
	 *     return err
	 * }
	 * mgr.servers[s.Port] = s
	 * log.Debugf("Adding server: %s", s)
	 * return nil */
}

/* func (mgr *manager) remove(port int32) {
 *     mgr.serverMu.Lock()
 *     defer mgr.serverMu.Unlock()
 *     delete(mgr.servers, port)
 * } */

func (mgr *manager) Remove(port int32) error {
	/* mgr.serverMu.Lock()
	 * defer mgr.serverMu.Unlock()
	 * s, ok := mgr.servers[port]
	 * if !ok {
	 *     return ErrServerNotFound
	 * }
	 * delete(mgr.servers, port)
	 * mgr.kill(s)
	 * log.Debugf("Removing server(%d)", s.Port) */
	return nil
}

func (mgr *manager) ListServers() map[int32]*Server {
	mgr.serverMu.RLock()
	defer mgr.serverMu.RUnlock()
	currentServers := make(map[int32]*Server)
	/* for port, s := range mgr.servers {
	 *     currentServers[port] = s.clone()
	 * } */
	return currentServers
}

func (mgr *manager) GetServer(port int32) (*Server, error) {
	mgr.serverMu.RLock()
	defer mgr.serverMu.RUnlock()
	s, _ := mgr.servers[port]
	/* if !ok {
	 *     return nil, ErrServerNotFound
	 * } */
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
	/* if !validPort(s.Port) {
	 *     log.Warn("Invalid port.")
	 * }
	 * s.path = serverPath
	 * err := s.Restore(serverPath)
	 * if err != nil {
	 *     return err
	 * }
	 * if s.Alive() {
	 *     mgr.fillServerOptions(s)
	 *     if err := mgr.add(s); err != nil {
	 *         return err
	 *     }
	 *     return nil
	 * }
	 * err = mgr.Add(s)
	 * if err != nil {
	 *     return err
	 * } */
	return nil
}

// Restore starts all ss-servers that leaves their dirs in manager.path
func (mgr *manager) Restore() error {
	/* if _, err := os.Stat(mgr.path); err != nil {
	 *     return err
	 * }
	 * if !isDir(mgr.path) {
	 *     return errors.New(mgr.path + " is not a directory.")
	 * }
	 * names, err := readDirNames(mgr.path)
	 * if err != nil {
	 *     return err
	 * }
	 * for _, name := range names {
	 *     serverPath := path.Join(mgr.path, name)
	 *     if isDir(serverPath) {
	 *         if port, ok := getPort(name); ok {
	 *             log.Infof("Restoring server(%d) ...", port)
	 *             err := mgr.restore(&Server{Port: port}, serverPath)
	 *             if err != nil {
	 *                 log.Warnf("Can not restore server(%d), remove it. %s", port, err)
	 *                 os.RemoveAll(serverPath)
	 *             }
	 *         } else {
	 *             log.Warnf("Ignore unrecognized port %s.", name)
	 *         }
	 *     } else {
	 *         log.Warnf("Ignore normal file %s.", serverPath)
	 *     }
	 * } */
	return nil
}

func (mgr *manager) CleanUp() {
	os.RemoveAll(mgr.path)
	for p := range mgr.ListServers() {
		mgr.Remove(p)
	}
}
