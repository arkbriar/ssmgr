package shadowsocks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/Sirupsen/logrus"
	proc "github.com/arkbriar/ss-mgr/slave/shadowsocks/process"
	"github.com/coreos/go-iptables/iptables"
)

var (
	ipt *iptables.IPTables
	usr *user.User
)

func init() {
	if _, err := exec.LookPath("ss-server"); err != nil {
		log.Warnf("Can not find ss-server in $PATH. Install it.")
		os.Exit(-1)
	}
	if runtime.GOOS != "linux" {
		log.Warnf("Connection limit and auto ban is not supported on non-linux system.")
	} else {
		usr, _ = user.Current()
		if usr.Name == "root" {
			ipt, _ = iptables.New()
		} else {
			log.Warnf("Connection limit and auto ban is only supported when running with root.")
		}
	}
}

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

func (o *serverOptions) args() []string {
	args := make([]string, 0)
	if o.UDPRelay {
		args = append(args, "-u")
	}
	if o.IPv6First {
		args = append(args, "-6")
	}
	if o.MPTCP {
		args = append(args, "--mptcp")
	}
	if o.TCPFastOpen {
		args = append(args, "--fast-open")
	}
	if o.Auth {
		args = append(args, "-A")
	}
	if len(o.NameServer) != 0 {
		args = append(args, "-d", o.NameServer)
	}
	if len(o.PidFile) != 0 {
		args = append(args, "-f", o.PidFile)
	}
	if len(o.ManagerAddress) != 0 {
		args = append(args, "--manager-address", o.ManagerAddress)
	}
	if o.FireWall {
		args = append(args, "--firewall")
	}
	if o.Verbose {
		args = append(args, "-v")
	}
	return args
}

func (o *serverOptions) reset() {
	*o = serverOptions{}
}

var methods = []string{
	"table", "rc4", "rc4-md5", "aes-128-cfb", "aes-192-cfb", "aes-256-cfb",
	"aes-128-ctr", "aes-192-ctr", "aes-256-ctr", "bf-cfb", "camellia-128-cfb",
	"camellia-192-cfb", "camellia-256-cfb", "cast5-cfb", "des-cfb", "idea-cfb",
	"rc2-cfb", "seed-cfb", "salsa20", "chacha20", "chacha20-ietf",
}

// validEncryptMethod checks if the encrypt method is supported.
func validEncryptMethod(m string) bool {
	for _, method := range methods {
		if m == method {
			return true
		}
	}
	return false
}

func validPort(p int32) bool {
	return p > 0 && p < (1<<16)
}

type serverRuntime struct {
	proc *os.Process
}

func (rt *serverRuntime) alive() bool {
	return rt.proc != nil && proc.Alive(rt.proc.Pid)
}

type serverExtra struct {
	StartTime time.Time `json:"start_time"`
}

// Server represents a ss-server instance.
type Server struct {
	Host        string       `json:"server"`
	Port        int32        `json:"server_port"`
	Password    string       `json:"password"`
	Method      string       `json:"method"`
	Timeout     int          `json:"timeout"`
	Extra       *serverExtra `json:"extra,omitempty"`
	opts        serverOptions
	connLimit   int
	watchDaemon struct {
		enable bool
		cancel context.CancelFunc
	}
	rtMu    sync.RWMutex
	runPath string
	runtime *serverRuntime
	stat    atomic.Value
}

// WithUDPRelay enables udp relay.
func (s *Server) WithUDPRelay() *Server {
	s.opts.UDPRelay = true
	return s
}

// WithIPv6First enables resolving to ipv6 first.
func (s *Server) WithIPv6First() *Server {
	s.opts.IPv6First = true
	return s
}

// WithMPTCP enables MPTCP.
func (s *Server) WithMPTCP() *Server {
	s.opts.MPTCP = true
	return s
}

// WithTCPFastOpen enables tcp fast open.
func (s *Server) WithTCPFastOpen() *Server {
	s.opts.TCPFastOpen = true
	return s
}

// WithOneTimeAuth enables one time auth.
func (s *Server) WithOneTimeAuth() *Server {
	s.opts.Auth = true
	return s
}

// WithNameServer sets the default nameserver to use.
func (s *Server) WithNameServer(ns string) *Server {
	s.opts.NameServer = ns
	return s
}

// WithPidFile sets the pid file.
func (s *Server) WithPidFile(f string) *Server {
	s.opts.PidFile = f
	return s
}

// WithManagerAddress sets the manager address.
func (s *Server) WithManagerAddress(addr string) *Server {
	s.opts.ManagerAddress = addr
	return s
}

// WithInterface sets the interface to listen.
func (s *Server) WithInterface(itf string) *Server {
	s.opts.Interface = itf
	return s
}

// WithFireWall enables firewall for auto ban.
func (s *Server) WithFireWall() *Server {
	if runtime.GOOS != "linux" {
		return s
	}
	if usr.Name == "root" {
		s.opts.FireWall = true
	}
	return s
}

// WithVerbose sets the verbose mode.
func (s *Server) WithVerbose() *Server {
	s.opts.Verbose = true
	return s
}

// WithWatchDaemon starts a watch daemon to check health of ss-server's process and
// restart the unexpectely dead process.
func (s *Server) WithWatchDaemon() *Server {
	s.watchDaemon.enable = true
	return s
}

// WithConnLimit sets the connection limit for shadowsock port.
func (s *Server) WithConnLimit(l int) *Server {
	if runtime.GOOS != "linux" {
		return s
	}
	s.connLimit = l
	return s
}

// WithRunPath sets the running path to store config of this server.
func (s *Server) WithRunPath(runPath string) *Server {
	s.runPath = runPath
	return s
}

// WithDefaults sets the default options for the server.
func (s *Server) WithDefaults() *Server {
	return s.WithConnLimit(32).
		WithWatchDaemon().
		WithUDPRelay().
		WithVerbose()
}

// ResetOptions sets all the server options to default.
func (s *Server) ResetOptions() *Server {
	s.opts.reset()
	return s
}

func (s *Server) args() []string {
	var args []string
	if len(s.runPath) != 0 {
		args = []string{"-c", path.Join(s.runPath, "ss_server.conf")}
	} else {
		args = []string{"-s", s.Host, "-p", fmt.Sprint(s.Port), "-m", s.Method, "-k", s.Password, "-d", fmt.Sprint(s.Timeout)}
	}
	return append(args, s.opts.args()...)
}

func (s *Server) valid() bool {
	return len(s.Host) != 0 && validPort(s.Port) && len(s.Password) >= 8 && validEncryptMethod(s.Method) && s.Timeout > 0
}

// command constructs a new shadowsock server command
func (s *Server) command() *exec.Cmd {
	return exec.Command("ss-server", s.args()...)
}

// Command returns the command string this server starts with.
func (s *Server) Command() string {
	return fmt.Sprintf("ss-server %s", strings.Join(s.args(), " "))
}

func (s *Server) clone() *Server {
	s.rtMu.RLock()
	defer s.rtMu.RUnlock()
	c := *s
	c.rtMu = sync.RWMutex{}
	return &c
}

// String implements the Stringer interface.
func (s *Server) String() string {
	data, _ := json.Marshal(s)
	return string(data)
}

var (
	errWatchDaemonNotEnabled     = errors.New("watch daemon is not enabled")
	errWatchDaemonAlreadyStarted = errors.New("watch daemon is already started")
	errWatchDaemonIsNotStarted   = errors.New("watch daemon is not started")
)

func (s *Server) startWatchDaemon() error {
	if !s.watchDaemon.enable {
		return errWatchDaemonNotEnabled
	}
	if s.watchDaemon.cancel != nil {
		return errWatchDaemonAlreadyStarted
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.watchDaemon.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				if !s.Alive() {
					log.Warnf("Server(%s) is detected dead.", s)
					if err := s.revive(); err != nil {
						if err != errServerAlive {
							log.Warnf("Can not restart server(%s), %s", s, err)
						}
					} else {
						log.Infof("Server(%s) is back to work.", s)
					}
				}
			}
		}
	}(ctx)
	return nil
}

func (s *Server) stopWatchDaemon() error {
	if !s.watchDaemon.enable {
		return errWatchDaemonNotEnabled
	}
	if s.watchDaemon.cancel == nil {
		return errWatchDaemonIsNotStarted
	}
	s.watchDaemon.cancel()
	s.watchDaemon.cancel = nil
	return nil
}

func (s *Server) connLimitIptablesRule() []string {
	if s.connLimit != 0 {
		return []string{"-p", "tcp", "--syn", "--dport", fmt.Sprint(s.Port), "-m", "connlimit", "--connlimit-above", fmt.Sprint(s.connLimit), "-j", "REJECT", "--reject-with", "tcp-reset"}
	}
	return nil
}

func (s *Server) createConnLimit() error {
	if ipt == nil {
		return nil
	}
	rule := s.connLimitIptablesRule()
	err := ipt.AppendUnique("filter", "INPUT", rule...)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) deleteConnLimit() error {
	if ipt == nil {
		return nil
	}
	rule := s.connLimitIptablesRule()
	err := ipt.Delete("filter", "INPUT", rule...)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) checkConnLimit() (bool, error) {
	return ipt.Exists("filter", "INPUT", s.connLimitIptablesRule()...)
}

func readPidFile(filename string) (int, error) {
	pidname, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(pidname))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func findProcFromPidFile(filename string) (*os.Process, error) {
	pid, err := readPidFile(filename)
	if err != nil {
		return nil, err
	}
	return os.FindProcess(pid)
}

func (s *Server) save(filename string) error {
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

func (s *Server) load(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}

var (
	errServerAlive          = errors.New("Server is alive.")
	errServerAlreadyStarted = errors.New("server already started")
	errServerNotStarted     = errors.New("server not started")
)

func (s *Server) unsafeExec() error {
	cmd := s.command()
	if len(s.runPath) != 0 {
		logw, err := os.Create(path.Join(s.runPath, "ss_server.log"))
		if err != nil {
			log.Warnf("Can not open log file, %s", err)
		} else {
			cmd.Stdout, cmd.Stderr = logw, logw
		}
	}
	if len(s.opts.PidFile) != 0 {
		if err := cmd.Run(); err != nil {
			return err
		}
		proc, err := findProcFromPidFile(s.opts.PidFile)
		if err != nil {
			log.Warn(err)
			return errors.New("can not get process from pid file")
		}
		s.runtime = &serverRuntime{
			proc: proc,
		}
	} else {
		if err := cmd.Start(); err != nil {
			return err
		}
		s.runtime = &serverRuntime{
			proc: cmd.Process,
		}
	}
	return nil
}

func (s *Server) exec() error {
	s.rtMu.Lock()
	defer s.rtMu.Unlock()
	if s.runtime != nil {
		return errServerAlreadyStarted
	}
	return s.unsafeExec()
}

func (s *Server) afterStart() {
	if s.watchDaemon.enable {
		err := s.startWatchDaemon()
		if err != nil {
			log.Warn(err)
		}
	}
	if s.connLimit > 0 {
		err := s.createConnLimit()
		if err != nil {
			log.Warn(err)
		}
	}
}

// Start starts the server.
func (s *Server) Start() error {
	if !s.valid() {
		return errors.New("invalid server configuration")
	}
	if len(s.runPath) == 0 {
		return errors.New("start server without run path is not supported")
	}
	s.Extra = &serverExtra{
		StartTime: time.Now(),
	}
	err := s.save(path.Join(s.runPath, "ss_server.conf"))
	if err != nil {
		return err
	}
	err = s.exec()
	if err == nil {
		s.afterStart()
	}
	return err
}

func (s *Server) kill() error {
	s.rtMu.Lock()
	if s.runtime == nil {
		return errServerNotStarted
	}
	proc := s.runtime.proc
	s.runtime = nil
	s.Extra = nil
	/* go func(proc *os.Process, defers ...func()) {
	 *     for _, d := range defers {
	 *         defer d()
	 *     }
	 *     proc.Kill()
	 *     proc.Wait()
	 * }(proc, s.rtMu.Unlock) */
	proc.Kill()
	proc.Wait()
	s.rtMu.Unlock()
	return nil
}

func (s *Server) forceKill() {
	s.rtMu.Lock()
	if s.runtime == nil {
		return
	}
	proc := s.runtime.proc
	s.runtime = nil
	s.Extra = nil
	/* go func(proc *os.Process, defers ...func()) {
	 *     for _, d := range defers {
	 *         defer d()
	 *     }
	 *     proc.Kill()
	 *     proc.Wait()
	 * }(proc, s.rtMu.Unlock) */
	proc.Kill()
	proc.Wait()
	s.rtMu.Unlock()
}

func (s *Server) beforeStop() {
	if s.connLimit > 0 {
		err := s.deleteConnLimit()
		if err != nil {
			log.Warn(err)
		}
	}
	if s.watchDaemon.enable {
		err := s.stopWatchDaemon()
		if err != nil {
			log.Warn(err)
		}
	}
}

// Stop stops the server.
func (s *Server) Stop() error {
	s.beforeStop()
	return s.kill()
}

func (s *Server) restart() error {
	s.forceKill()
	return s.exec()
}

// Alive returns if the server is alive
func (s *Server) Alive() bool {
	s.rtMu.RLock()
	defer s.rtMu.RUnlock()
	return s.runtime != nil && s.runtime.alive()
}

// revive starts the process when it is dead.
func (s *Server) revive() error {
	s.rtMu.Lock()
	defer s.rtMu.Unlock()
	if s.runtime == nil || !s.runtime.alive() {
		s.runtime = nil
		if err := s.unsafeExec(); err != nil {
			return err
		}
		s.Extra = &serverExtra{
			StartTime: time.Now(),
		}
	} else {
		return errServerAlive
	}
	return nil
}

func (s *Server) restoreRuntime(runPath string) error {
	s.rtMu.Lock()
	defer s.rtMu.Unlock()
	proc, err := findProcFromPidFile(path.Join(runPath, "ss_server.pid"))
	if err != nil {
		return err
	}
	s.runtime = &serverRuntime{
		proc: proc,
	}
	if !s.runtime.alive() {
		s.runtime = nil
		log.Debugf("Recovered process is not alive, reset runtime.")
	}
	return nil
}

func (s *Server) restoreConf(runPath string) error {
	return s.load(path.Join(runPath, "ss_server.conf"))
}

// Restore from files leaved.
func (s *Server) Restore(runPath string) error {
	err := s.restoreConf(runPath)
	if err != nil {
		return err
	}
	err = s.restoreRuntime(runPath)
	if err != nil {
		log.Warnf("Can not restore runtime of server (%s)", s)
	} else {
		if s.Alive() {
			s.afterStart()
		}
	}
	s.runPath = runPath
	return nil
}

// Stat represents the statistics collected from a shadowsocks server
type Stat struct {
	Traffic int64 `json:"traffic"` // Transfered traffic in bytes
	/* Rx      int64 `json:"rx"`      // Receive in bytes
	 * Tx      int64 `json:"tx"`      // Transmit in bytes */
}

func (s *Server) updateStat(stat Stat) {
	s.stat.Store(stat)
}

// GetStat returns the stats of the server.
func (s *Server) GetStat() Stat {
	stat := s.stat.Load()
	if stat != nil {
		return stat.(Stat)
	}
	return Stat{}
}
