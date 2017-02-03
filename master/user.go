package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/kataras/iris"
	"github.com/satori/go.uuid"

	"github.com/arkbriar/ss-mgr/master/orm"
	rpc "github.com/arkbriar/ss-mgr/protocol"
)

func CreateUser(email string) *orm.User {
	userID := hex.EncodeToString(uuid.NewV4().Bytes())
	user := orm.User{
		ID:        userID,
		Email:     email,
		QuotaFlow: config.Limit.Flow * 1024 * 1024,
		Time:      time.Now().Unix(),
		Expired:   time.Now().Add(time.Duration(config.Limit.Time) * time.Hour).Unix(),
		Disabled:  false,
	}
	db.Save(&user)

	iris.Logger.Printf("New user: %s, email: %s", user.ID, user.Email)

	// Allocating ports is slow, do it in another thread
	go func() {
		for serverID, _ := range config.Slaves {
			err := AllocateForUser(user.ID, serverID)
			if err != nil {
				logrus.Errorf("Failed to allocate ports for %s: %s", userID, err.Error())
			}
		}
	}()

	return &user
}

func RemoveUser(userIDs ...string) {
	if len(userIDs) == 0 {
		return
	}

	db.Table("users").Where("id IN (?)", userIDs).Updates(&orm.User{Disabled: true})

	var allocs []orm.Allocation
	db.Where("user_id IN (?)", userIDs).Scan(&allocs)

	// gorm may print "no such table", just ignore it
	db.Where("user_id IN (?)", userIDs).Delete(&orm.Allocation{})

	// Freeing ports is slow, do it in another thread
	go func() {
		for _, alloc := range allocs {
			err := FreeAllocation(alloc.ServerID, alloc.Port)
			if err != nil {
				logrus.Errorf("Failed to free ports for %s: %s", alloc.UserID, err.Error())
			}
		}
	}()
}

func AllocateAllUsers() {
	var users []*orm.User
	db.Where("disabled = 0").Find(&users)

	for _, user := range users {
		for id, _ := range slaves {
			err := AllocateForUser(user.ID, id)
			if err != nil {
				logrus.Error(err.Error())
			}
		}
	}
}

func AllocateForUser(userID, serverID string) error {
	slave := slaves[serverID]
	if slave == nil {
		return fmt.Errorf("Server '%s' not found", serverID)
	}
	port, password, err := findOrInitAllocation(userID, serverID)
	if err != nil {
		return fmt.Errorf("Failed to get port for user %s: %s", userID, err.Error())
	}

	logrus.Debugf("Allocate for user %s on server %s: Port %d, Password: %s",
		userID, serverID, port, password)
	_, err = slave.stub.Allocate(slave.ctx, &rpc.AllocateRequest{
		Port:     int32(port),
		Password: password,
		Method:   "aes-256-cfb", // const
	})
	if err != nil {
		return fmt.Errorf("Failed to allocate port: %s", err.Error())
	}
	return nil
}

func findOrInitAllocation(userID, serverID string) (int, string, error) {
	serverConfig := config.Slaves[serverID]

	var allocation orm.Allocation
	db.Where(&orm.Allocation{
		UserID:   userID,
		ServerID: serverID,
	}).FirstOrInit(&allocation)

	if allocation.Port == 0 {
		// Not record is found in allocation table.
		// Search for an empty port and write into allocation table.
		var allocated []orm.Allocation
		db.Where(&orm.Allocation{
			ServerID: serverID,
		}).Find(&allocated)

		ports := make([]bool, 65536)
		for _, alloc := range allocated {
			ports[alloc.Port] = true
		}

		var empty int
		for i := serverConfig.PortMin; i <= serverConfig.PortMax; i++ {
			if ports[i] == false {
				empty = i
				break
			}
		}

		if empty == 0 {
			// TODO: this error should be told to user or manager
			return 0, "", fmt.Errorf("no port is available in %s", serverID)
		}

		allocation.Port = empty
		allocation.Password = RandomPassword()
		db.Save(&allocation)
	}

	return allocation.Port, allocation.Password, nil
}

func FreeAllocation(serverID string, port int) error {
	slave := slaves[serverID]
	if slave == nil {
		return fmt.Errorf("Server '%s' not found", serverID)
	}

	_, err := slave.stub.Free(slave.ctx, &rpc.FreeRequest{
		Port: int32(port),
	})
	if err != nil {
		return err
	}
	return nil
}
