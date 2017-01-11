package shadowsocks

import (
	"encoding/json"
	"fmt"
	"net"
)

type Server struct {
}

// Manager is an interface provides encapsulation of protocol of shadowsocks
// manager. One can add a port alone with corresponding password by calling `Add()`
// and remove a specified port with `Remove()`.
// Method `Ping()` is used to send a heartbeat and expect a pong back.
// Method `Stat()` returns the tranfer statistics of opened ports.
//
// Example:
//	mgr := shadowsocks.NewManager("localhost:6001")
//	if err := mgr.Dial(); err != nil {
//		log.Panicln(err)
//	}
//	defer mgr.Close()
//	newPort := 8001
//	newPassword = "7cd308cc059"
//	err := mgr.Add(newPort, newPassword)
//	if err != nil {
//		log.Panicln(err)
//	}
//	log.Printf("Add port %d with password %s\n", newPort, newPassword)
//	if err := mgr.Remove(newPort); err != nil {
//		log.Panicln(err)
//	}
//	log.Printf("Remove port %d\n", newPort)
//
// Example heartbeat service:
//	alive := make(chan struct{}, 1)
//	go func(alive chan struct{}, mgr shadowsocks.Manager) {
//		// You must ensure that channel `alive` is not nil.
//		for {
//			if err := mgr.Ping(); err != nil {
//				break
//			}
//			time.Sleep(500 * time.Millisecond)
//		}
//		close(alive)
//	} (alive, mgr)
//	select {
//	case <- alive:
//		// hang here
//	}
type Manager interface {
	// Dial connects to remote manager
	Dial() error
	// Add adds a port along with password
	Add(port int, password string) error
	// Remove deletes an opened port
	Remove(port int) error
	// Ping send a ping and expect a pong
	Ping() error
	// Get the current transfer stat
	Stat() map[string]int64
	// Close closes the connection
	Close() error
}

// An implementation of manager, may work well on the same host of
// shadowsocks manager.
type manager struct {
	// remote shadowsocks manager url
	remoteURL string
	// udp connection of remote manager, opened on Dial() and closed on
	// Close()
	conn net.Conn
	// transfer statistics of manager
	stat map[string]int64
}

// NewLocalManager returns a manager which should only be used locally
func NewLocalManager(url string) Manager {
	return &manager{
		remoteURL: url,
		conn:      nil,
		stat:      nil,
	}
}

func (mgr *manager) Dial() error {
	conn, err := net.Dial("udp", mgr.remoteURL)
	if err != nil {
		return err
	}
	mgr.conn = conn
	return nil
}

type addMsg struct {
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
}

func (mgr *manager) Add(port int, password string) error {
	msg := &addMsg{
		serverPort: port,
		password:   password,
	}
	bytes, _ := json.Marshal(msg)
	_, err := mgr.conn.Write(append([]byte("add: "), bytes...))
	if err != nil {
		return err
	}
	var respBytes []byte
	_, err = mgr.conn.Read(respBytes)
	if err != nil {
		return err
	}
	if string(respBytes) != "ok" {
		return fmt.Errorf("Invalid response.")
	}
	return nil
}

type removeMsg struct {
	ServerPort int `json:"server_port"`
}

func (mgr *manager) Remove(port int) error {
	msg := &removeMsg{
		serverPort: port,
	}
	bytes, _ := json.Marshal(msg)
	_, err := mgr.conn.Write(append([]byte("remove: "), bytes...))
	if err != nil {
		return err
	}
	var respBytes []byte
	_, err = mgr.conn.Read(respBytes)
	if err != nil {
		return err
	}
	if string(respBytes) != "ok" {
		return fmt.Errorf("Invalid response.")
	}
	return nil
}

func (mgr *manager) Stat() map[string]int64 {
	return mgr.stat
}

func (mgr *manager) Ping() error {
	_, err := mgr.conn.Write([]byte("ping"))
	if err != nil {
		return err
	}
	var respBytes []byte
	_, err = mgr.conn.Read(respBytes)
	if err != nil {
		return err
	}
	if string(respBytes) != "pong" {
		return fmt.Errorf("Invalid response.")
	}
	return nil
}

func (mgr *manager) Close() error {
	if mgr.conn != nil {
		return mgr.conn.Close()
	}
	return fmt.Errorf("Close on empty connection.")
}
