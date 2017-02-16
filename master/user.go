package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"

	"github.com/arkbriar/ssmgr/master/orm"
	rpc "github.com/arkbriar/ssmgr/protocol"
)

func CreateUser(email string) *orm.User {
	now := time.Now()
	userID := hex.EncodeToString(uuid.NewV4().Bytes())
	user := orm.User{
		ID:        userID,
		Email:     email,
		QuotaFlow: defaultGroup.Config.Limit.Flow * 1024 * 1024,
		Time:      now.Unix(),
		Expired:   now.Add(time.Duration(defaultGroup.Config.Limit.Time) * time.Hour).Unix(),
		Disabled:  false,
		Group:     "default",
	}
	db.Save(&user)

	logrus.Infof("New user: %s, email: %s", user.ID, user.Email)

	// Allocating ports is slow, do it in another thread
	go allocateForUser(userID, "default")

	return &user
}

func ChangeUserGroup(userID, groupID string) error {
	var user orm.User
	db.First(&user)
	if user.ID == "" {
		return fmt.Errorf("User not found: %s", userID)
	}
	group := groups[groupID]
	if group == nil {
		return fmt.Errorf("Group not found: %s", groupID)
	}
	user.Group = groupID
	user.Expired = time.Unix(user.Time, 0).Add(time.Duration(defaultGroup.Config.Limit.Time) * time.Hour).Unix()
	user.QuotaFlow = group.Config.Limit.Flow * 1024 * 1024
	// Let the daemon routine check whether to remove user (set disable = 1)

	go func() {
		removeUserAllocation(userID)
		allocateForUser(userID, groupID)
	}()

	err := db.Debug().Save(&user).Error
	return err
}

func RemoveUser(userIDs ...string) {
	if len(userIDs) == 0 {
		return
	}

	db.Table("users").Where("id IN (?)", userIDs).Updates(&orm.User{Disabled: true})

	go removeUserAllocation(userIDs...)
}

func removeUserAllocation(userIDs ...string) {
	var allocs []orm.Allocation
	db.Where("user_id IN (?)", userIDs).Scan(&allocs)

	// gorm may print "no such table", just ignore it
	db.Where("user_id IN (?)", userIDs).Delete(&orm.Allocation{})

	for _, alloc := range allocs {
		err := FreeAllocation(alloc.ServerID, alloc.Port)
		if err != nil {
			logrus.Errorf("Failed to free ports for %s: %s", alloc.UserID, err.Error())
		}
	}
}

func AllocateAllUsers() {
	var users []*orm.User
	db.Where("disabled = 0").Find(&users)

	for _, user := range users {
		for _, serverID := range groups[user.Group].Config.SlaveIDs {
			port, password, err := findOrInitAllocation(user.ID, serverID)
			if err != nil {
				logrus.Error(err.Error())
				continue
			}
			logrus.Debugf("Allocate for user %s on server %s: Port %d, Password: %s",
				user.ID, serverID, port, password)
		}
	}
}

func allocateForUser(userID, groupID string) {
	for _, serverID := range groups[groupID].Config.SlaveIDs {
		err := allocateServerToUser(userID, serverID)
		if err != nil {
			logrus.Errorf("Failed to allocate ports for %s: %s", userID, err.Error())
		}
	}
}

func allocateServerToUser(userID, serverID string) error {
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
	serverConfig := slaves[serverID].Config

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
