package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/arkbriar/ss-mgr/master/orm"
)

var (
	configPath = flag.String("c", "config.json", "Path of config file")
	verbose    = flag.Bool("v", false, "Verbose mode")
	webroot    = flag.String("w", "./frontend", "Path of web UI files")
)

type SlaveConfig struct {
	ID      string `json:"id"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Token   string `json:"token"`
	Name    string `json:"name"`
	PortMax int    `json:"portMax"`
	PortMin int    `json:"portMin"`
}

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Limit    struct {
		Flow int64 `json:"flow"`
		Time int64 `json:"time"`
	} `json:"limit"`
	Slaves []*SlaveConfig `json:"slaves"`
	Email  struct {
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
	Interval int `json:"interval"`
}

var db *gorm.DB

var config *Config

func main() {
	flag.Parse()
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var err error
	config, err = parseConfig(*configPath)
	if err != nil {
		logrus.Fatal(config)
	}

	db = orm.New(config.Database.Dialect, config.Database.Args)

	InitSlaves()

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
