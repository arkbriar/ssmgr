package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/arkbriar/ssmgr/master/orm"

	"github.com/arkbriar/ssmgr/master/slack"
)

var (
	configPath = flag.String("c", "config.json", "Path of config file")
	verbose    = flag.Bool("v", false, "Verbose mode")
	webroot    = flag.String("w", "./frontend", "Path of web UI files")
)

type SlaveConfig struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Token   string `json:"token"`
	PortMax int    `json:"portMax"`
	PortMin int    `json:"portMin"`
}

type GroupConfig struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	SlaveIDs []string `json:"slaves"`
	Limit    struct {
		Flow int64 `json:"flow"` // MB
		Time int64 `json:"time"` // hours
	} `json:"limit"`
}

type Config struct {
	Host     string         `json:"host"`
	Port     int            `json:"port"`
	Password string         `json:"password"`
	Interval int            `json:"interval"`
	Slaves   []*SlaveConfig `json:"slaves"`
	Groups   []*GroupConfig `json:"groups"`
	Email    struct {
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromAddr  string `json:"fromAddr"`
		FromAlias string `json:"fromAddr"`
	} `json:"email"`
	Database struct {
		Dialect string `json:"dialect"`
		Args    string `json:"args"`
	} `json:"database"`
	Slack *struct {
		Token   string   `json:"token"`
		Channel string   `json:"channel"`
		Levels  []string `json:"levels"`
	} `json:"slack,omitempty"`
}

var db *gorm.DB

var config *Config

func parseLogrusLevels(levels []string) ([]logrus.Level, error) {
	if levels == nil {
		return nil, nil
	}
	ret := make([]logrus.Level, len(levels))
	for _, level := range levels {
		l, err := logrus.ParseLevel(level)
		if err != nil {
			return nil, err
		}
		ret = append(ret, l)
	}
	return ret, nil
}

func main() {
	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var err error
	config, err = parseConfig(*configPath)
	if err != nil {
		logrus.Fatal(err)
	}

	// enable slack hook if slack is configured

	if config.Slack != nil {
		levels, err := parseLogrusLevels(config.Slack.Levels)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.AddHook((&slack.SlackrusHook{
			Channel:        config.Slack.Channel,
			Token:          config.Slack.Token,
			AcceptedLevels: levels,
		}).Connect())
	}

	db = orm.New(config.Database.Dialect, config.Database.Args)

	InitSlaves()
	InitGroups()

	// If servers config is changed, clear removed and allocate new
	CleanInvalidAllocation()
	AllocateAllUsers()

	go Monitoring()

	webServer := NewApp(*webroot)
	listenAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	webServer.Listen(listenAddr)
}

func parseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
