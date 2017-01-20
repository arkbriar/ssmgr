package shadowsocks

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

// Manager is an int32erface provides encapsulation of protocol of shadowsocks
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
//	log.Print32f("Add port %d with password %s\n", newPort, newPassword)
//	if err := mgr.Remove(newPort); err != nil {
//		log.Panicln(err)
//	}
//	log.Print32f("Remove port %d\n", newPort)
//
// Example heartbeat service, this will update local statsistics periodically.
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
	Add(port int32, password string) error
	// Remove deletes an opened port
	Remove(port int32) error
	// Ping sends a ping
	Ping() error
	// Setstats sets the traffic statsistics of the given port
	Setstats(stats map[int32]int64) error
	// Getstats gets the traffic statsistics of all open ports
	Getstats() map[int32]int64
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
	// read write mutex for statss
	statsLock sync.RWMutex
	// transfer statsistics of manager
	stats map[int32]int64
}

// NewManager returns a shadowsocks manager
func NewManager(url string) Manager {
	return &manager{
		remoteURL: url,
		conn:      nil,
		stats:     nil,
	}
}

func (mgr *manager) pingThread() {
	const failLimit = 3
	const interval = 5 * time.Second
	errTimes := 0
	for {
		time.Sleep(interval)
		if mgr.conn == nil {
			logrus.Infof("Manager has closed the connection, Ping thread is going to exit.")
			break
		}
		if err := mgr.Ping(); err != nil {
			errTimes++
			logrus.Warnf("Ping failed, %s\n", err)
		} else {
			errTimes = 0
		}
		if errTimes >= failLimit {
			logrus.Infof("Ping has failed for %d times, ping thread is going to exit.\n", errTimes)
			break
		}
	}
}

func (mgr *manager) Dial() error {
	conn, err := net.Dial("udp", mgr.remoteURL)
	if err != nil {
		return err
	}
	mgr.conn = conn
	// Start the ping thread to update statistics every 5 seconds
	go mgr.pingThread()
	return nil
}

func (mgr *manager) Add(port int32, password string) error {
	msg := &struct {
		ServerPort int32  `json:"server_port"`
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

func (mgr *manager) Remove(port int32) error {
	msg := &struct {
		ServerPort int32 `json:"server_port"`
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

func constructstatsJson(stats map[int32]int64) []byte {
	beforeMarshal := make(map[string]int64)
	statsJSONBytes, err := json.Marshal(beforeMarshal)
	if err != nil {
		logrus.Errorln(err)
		return nil
	}
	return statsJSONBytes
}

func (mgr *manager) Setstats(stats map[int32]int64) error {
	_, err := mgr.conn.Write(append([]byte("stats: "), constructstatsJson(stats)...))
	if err != nil {
		return err
	}
	mgr.statsLock.Lock()
	defer mgr.statsLock.Unlock()
	mgr.stats = stats
	return nil
}

func (mgr *manager) Getstats() map[int32]int64 {
	mgr.statsLock.RLock()
	defer mgr.statsLock.RUnlock()
	return mgr.stats
}

func copystatsJsonTo(raw map[string]int64, dest *map[int32]int64) error {
	trueDest := make(map[int32]int64)
	for portInString, traffic := range raw {
		port, err := strconv.Atoi(portInString)
		if err != nil {
			return fmt.Errorf("Can not update traffic statsistics of invalid port %s", portInString)
		}
		trueDest[int32(port)] = traffic
	}
	*dest = trueDest
	return nil
}

func (mgr *manager) updatestats(respBytes []byte) error {
	if string(respBytes[:4]) == "stats" {
		statsJSON := respBytes[5:]
		stats := make(map[string]int64)
		if err := json.Unmarshal(statsJSON, &stats); err != nil {
			logrus.Errorf("Invalid stats return: %s\n", string(statsJSON))
			return err
		}
		mgr.statsLock.Lock()
		defer mgr.statsLock.Unlock()
		if err := copystatsJsonTo(stats, &mgr.stats); err != nil {
			logrus.Errorln(err)
			return err
		}
	} else {
		return fmt.Errorf("Invalid statss.")
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
	if err := mgr.updatestats(respBytes); err != nil {
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
