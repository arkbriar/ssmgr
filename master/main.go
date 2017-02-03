package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"

	"github.com/arkbriar/ss-mgr/master/orm"
)

var configPath = flag.String("c", "config.json", "path of config file")

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Limit    struct {
		Flow int64 `json:"flow"`
		Time int64 `json:"time"`
	} `json:"limit"`
	Slaves map[string]*struct {
		Host    string `json:"host"`
		Port    int    `json:"port"`
		Token   string `json:"token"`
		Name    string `json:"name"`
		PortMax int    `json:"portMax"`
		PortMin int    `json:"portMin"`
	} `json:"slaves"`
	Email struct {
		Host      string `json:"host"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromAddr  string `json:"fromAddr"`
		FromAlias string `json:"fromAddr"`
	} `json:"email"`
}

var db   *gorm.DB

var config *Config

func main() {
	
	logrus.SetLevel(logrus.DebugLevel)
	
	var err error
	config, err = parseConfig(*configPath)
	if err != nil {
		logrus.Fatal(config)
	}
	
	db = orm.New()
	
	InitSlaves()

	CleanInvalidAllocation()

	AllocateAllUsers()
	
	go Monitoring()
	
	webServer := NewApp()
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
