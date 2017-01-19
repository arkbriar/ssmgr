package shadowsocks

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/Sirupsen/logrus"
)

// Manager is an interface provides encapsulation of protocol of shadowsocks
// manager. One can add a port alone with corresponding password by calling `Add()`
// and remove a specified port with `Remove()`.
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
// Example heartbeat service, this will update local statistics periodically.
//	alive := make(chan struct{}, 1)
//	go func(alive chan struct{}, mgr shadowsocks.Manager) {
//		// You must ensure that channel `alive` is not nil.
//		for {
//			if err := mgr.Ping(); err != nil {
//				logrus.Errorln(err)
//			}
//			// Here we use some strategy to quit heartbeat when remote is offline.
//			// For example, 5 continous ping errors result in the break.
//			time.Sleep(5 * time.Second)
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
	// Ping sends a ping
	Ping() error
	// SetStat sets the traffic statistics of the given port
	SetStat(stat map[int32]int64) error
	// GetStat gets the traffic statistics of all open ports
	GetStat() map[int32]int64
	// Close closes the connection
	Close() error
}

// An implementation of manager
type manager struct {
	// remote shadowsocks manager url
	remoteURL string
	// udp connection of remote manager, opened on Dial() and closed on
	// Close()
	conn net.Conn
	// read write mutex for stats
	statLock sync.RWMutex
	// transfer statistics of manager
	stat map[int32]int64
}

// NewManager returns a shadowsocks manager
func NewManager(url string) Manager {
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

func (mgr *manager) Add(port int, password string) error {
	msg := &struct {
		ServerPort int    `json:"server_port"`
		Password   string `json:"password"`
	}{
		ServerPort: port,
		Password:   password,
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

func (mgr *manager) Remove(port int) error {
	msg := &struct {
		ServerPort int `json:"server_port"`
	}{
		ServerPort: port,
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

func constructStatJson(stat map[int32]int64) []byte {
	beforeMarshal := make(map[string]int64)
	statJSONBytes, err := json.Marshal(beforeMarshal)
	if err != nil {
		logrus.Errorln(err)
		return nil
	}
	return statJSONBytes
}

func (mgr *manager) SetStat(stat map[int32]int64) error {
	_, err := mgr.conn.Write(append([]byte("stat: "), constructStatJson(stat)...))
	if err != nil {
		return err
	}
	mgr.statLock.Lock()
	defer mgr.statLock.Unlock()
	mgr.stat = stat
	return nil
}

func (mgr *manager) GetStat() map[int32]int64 {
	mgr.statLock.RLock()
	defer mgr.statLock.RUnlock()
	return mgr.stat
}

func copyStatJsonTo(raw map[string]int64, dest *map[int32]int64) error {
	trueDest := make(map[int32]int64)
	for portInString, traffic := range raw {
		port, err := strconv.Atoi(portInString)
		if err != nil {
			return fmt.Errorf("Can not update traffic statistics of invalid port %s", portInString)
		}
		trueDest[port] = traffic
	}
	*dest = trueDest
	return nil
}

func (mgr *manager) updateStat(respBytes []byte) error {
	if string(respBytes[:4]) == "stat" {
		statJSON := respBytes[5:]
		stat := make(map[string]int64)
		if err := json.Unmarshal(statJSON, &stat); err != nil {
			logrus.Errorf("Invalid stat return: %s\n", string(statJSON))
			return err
		}
		mgr.statLock.Lock()
		defer mgr.statLock.Unlock()
		if err := copyStatJsonTo(stat, &mgr.stat); err != nil {
			logrus.Errorln(err)
			return err
		}
	} else {
		return fmt.Errorf("Invalid stats.")
	}
	return nil
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
	if err := mgr.updateStat(respBytes); err != nil {
		logrus.Errorln(err)
	}
	return nil
}

func (mgr *manager) Close() error {
	if mgr.conn != nil {
		return mgr.conn.Close()
	}
	return fmt.Errorf("Close on empty connection.")
}
