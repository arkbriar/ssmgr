package main

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	ss "github.com/arkbriar/ss-mgr/slave/shadowsocks"
	"github.com/arkbriar/ss-mgr/slave/shadowsocks/process"
)

func addServers(mgr ss.Manager, ports ...int32) {
	s := &ss.Server{
		/* Port:     8001, */
		Host:     "0.0.0.0",
		Password: "SomePass",
		Method:   "aes-256-cfb",
		Timeout:  60,
	}
	for _, port := range ports {
		s.Port = port
		err := mgr.Add(s)
		if err != nil && err != ss.ErrServerExists {
			log.Panicln("Can not create a new ss server, ", err)
		}
		log.Infof("Adding ss server: %s", s)
	}
}

func main() {
	log.SetLevel(log.DebugLevel)

	mgr := ss.NewManager(6001)
	if err := mgr.Listen(context.Background()); err != nil {
		log.Panicln("Can not listen udp address 127.0.0.1:6001, ", err)
	}
	go mgr.WatchDaemon(context.Background())

	ports := []int32{8001, 8002, 8003, 8004, 8005}
	addServers(mgr, ports...)

	s, err := mgr.GetServer(8001)
	if err != nil {
		log.Panicln(err)
	}
	pid := s.Process().Pid
	log.Infoln("Running ss servers: ", mgr.ListServers())
	log.Infoln("Waiting for 10s ...")
	time.Sleep(10 * time.Second)
	log.Infoln("Removing ss server on port 8001")
	if err := mgr.Remove(8001); err != nil {
		log.Panicln("Can not remove server, ", err)
	}
	if process.Alive(pid) {
		log.Panicf("Server process %d is not supposed to be alive after remove action.", pid)
	}
	ch := make(chan struct{})
	<-ch
	log.Infoln("Graceful shutdone!")
}
