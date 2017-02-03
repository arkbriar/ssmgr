package main

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/arkbriar/ss-mgr/master/orm"
	rpc "github.com/arkbriar/ss-mgr/protocol"
)

type Slave struct {
	stub rpc.SSMgrSlaveClient
	ctx  context.Context
}

var slaves map[string]*Slave

func InitSlaves() {
	slaves = make(map[string]*Slave)

	for id, info := range config.Slaves {
		address := fmt.Sprintf("%s:%d", info.Host, info.Port)
		md := metadata.Pairs("token", info.Token)
		ctx := metadata.NewContext(context.Background(), md)

		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			logrus.Warn("Failed to dial %s", address)
		}

		client := rpc.NewSSMgrSlaveClient(conn)

		slaves[id] = &Slave{
			stub: client,
			ctx:  ctx,
		}
	}
}

func CleanInvalidAllocation() {
	serverIDs := make([]string, 0)
	for serverID, _ := range slaves {
		serverIDs = append(serverIDs, serverID)
	}

	db.Where("server_id NOT IN (?)", serverIDs).Delete(&orm.Allocation{})
}

func Monitoring() {
	for {
		for id, slave := range slaves {
			if err := updateStats(id, slave); err != nil {
				logrus.Error("Update status error: ", err.Error())
			}
		}
		if err := checkUserLimit(); err != nil {
			logrus.Error("Check user limit error: ", err.Error())
		}
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

func updateStats(serverID string, slave *Slave) error {

	type portInfo struct {
		Password string
		UserID   string
	}
	portMap := make(map[int]portInfo)

	// Expected & actual ports allocation status
	var expected, actual []int

	var allocs []orm.Allocation
	db.Where("server_id = ?", serverID).Find(&allocs)
	for _, alloc := range allocs {
		expected = append(expected, alloc.Port)
		portMap[alloc.Port] = portInfo{
			Password: alloc.Password,
			UserID:   alloc.UserID,
		}
	}

	stats, err := slave.stub.GetStats(slave.ctx, &empty.Empty{})
	if err != nil {
		return err
	}
	for port, _ := range stats.Flow {
		actual = append(actual, int(port))
	}

	// In most cases expected ports should be same with actual ports.
	// If not, allocate the ports which should be allocated, and free ports which should not exist.

	shouldAlloc, shouldFree := diffPorts(expected, actual)

	for _, port := range shouldAlloc {
		_, err = slave.stub.Allocate(slave.ctx, &rpc.AllocateRequest{
			Port:     int32(port),
			Password: portMap[port].Password,
			Method:   "aes-256-cfb", // const
		})
		if err != nil {
			logrus.Errorf("Failed to allocate port: %s", err.Error())
		}
	}

	for _, port := range shouldFree {
		_, err = slave.stub.Free(slave.ctx, &rpc.FreeRequest{
			Port: int32(port),
		})
		if err != nil {
			logrus.Errorf("Failed to allocate port: %s", err.Error())
		}
	}

	// Update flow records according to statistics

	for port, stat := range stats.Flow {
		if _, ok := portMap[int(port)]; !ok {
			continue // skip shouldFree
		}
		var record orm.FlowRecord
		db.Where(&orm.FlowRecord{
			UserID:    portMap[int(port)].UserID,
			ServerID:  serverID,
			StartTime: stat.StartTime,
		}).FirstOrCreate(&record)

		// db.Save(&record) not works as expected due to gorm's bug

		db.Model(&orm.FlowRecord{}).Where(&orm.FlowRecord{
			UserID:    portMap[int(port)].UserID,
			ServerID:  serverID,
			StartTime: stat.StartTime,
		}).Update("flow", stat.Traffic)
	}

	return nil
}

func checkUserLimit() error {
	const SQL = `SELECT user_id, quota_flow, sum(flow) AS current_flow, expired
FROM users JOIN flow_record ON users.id == flow_record.user_id
WHERE disabled = 0
GROUP BY user_id`

	rows, err := db.Raw(SQL).Rows()
	if err != nil {
		return err
	}
	defer rows.Close()

	shouldDisable := make([]string, 0)
	for rows.Next() {
		var (
			userID      string
			quotaFlow   int64
			currentFlow int64
			expired     int64
		)
		rows.Scan(&userID, &quotaFlow, &currentFlow, &expired)

		if currentFlow >= quotaFlow || expired <= time.Now().Unix() {
			logrus.Infof("User expired or reached limit: %s", userID)
			shouldDisable = append(shouldDisable, userID)
		}
	}

	if len(shouldDisable) > 0 {
		RemoveUser(shouldDisable...)
	}
	return nil
}

func diffPorts(a []int, b []int) ([]int, []int) {
	ports := make([]int8, 65536)
	for _, p := range a {
		ports[p]--
	}
	for _, p := range b {
		ports[p]++
	}
	var diffA, diffB []int
	for p, v := range ports {
		switch v {
		case -1:
			diffA = append(diffA, p)
		case 1:
			diffB = append(diffB, p)
		}
	}
	return diffA, diffB
}
