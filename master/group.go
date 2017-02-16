package main

import "github.com/Sirupsen/logrus"

type Group struct {
	Config *GroupConfig
}

var groups map[string]*Group
var defaultGroup *Group

func InitGroups() {
	groups = make(map[string]*Group)

	for _, config := range config.Groups {
		groups[config.ID] = &Group{
			Config: config,
		}
	}

	defaultGroup = groups["default"]
	if defaultGroup == nil {
		logrus.Fatal("Group 'default' is required")
	}
}

// GetGroupIDs returns all groups' ids.
func GetGroupIDs() []string {
	ids := make([]string, 0)
	for id, _ := range groups {
		ids = append(ids, id)
	}
	return ids
}

// HasGroup returns if group exists
func HasGroup(id string) bool {
	_, ok := groups[id]
	return ok
}
