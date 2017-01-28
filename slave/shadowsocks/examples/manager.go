package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	ss "github.com/arkbriar/ss-mgr/slave/shadowsocks"
	"github.com/arkbriar/ss-mgr/slave/shadowsocks/process"
)

func main() {
	log.SetLevel(log.DebugLevel)
	mgr := ss.NewManager(6001)
	if err := mgr.Listen(); err != nil {
		log.Panicln("Can not listen udp address 127.0.0.1:6001, ", err)
	}
	newS := &ss.Server{
		Host:     "0.0.0.0",
		Port:     8001,
		Password: "SomePass",
		Method:   "aes-256-cfb",
		Timeout:  60,
	}
	log.Infof("Adding ss server: %s", newS)
	err := mgr.Add(newS)
	if err != nil {
		log.Panicln("Can not create a new ss server, ", err)
	}
	s, err := mgr.GetServer(8001)
	if err != nil {
		log.Panicln(err)
	}
	pid := s.Process().Pid
	log.Infoln("Running ss servers: ", mgr.ListServers())
	log.Infoln("Waiting for 5s ...")
	time.Sleep(5 * time.Second)
	log.Infoln("Removing ss server on port 8001")
	if err := mgr.Remove(8001); err != nil {
		log.Panicln("Can not remove server, ", err)
	}
	log.Infoln("Waiting for 5s ...")
	time.Sleep(5 * time.Second)
	if process.Alive(pid) {
		log.Panicf("Server process %d is not supposed to be alive after remove action.", pid)
	}
	log.Infoln("PASS")
}
