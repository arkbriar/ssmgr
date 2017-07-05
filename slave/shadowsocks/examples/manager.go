package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	log "github.com/Sirupsen/logrus"
	ss "github.com/arkbriar/ssmgr/slave/shadowsocks"
)

func addServers(mgr ss.Manager, ports ...int32) {
	s := &ss.Server{
		Host:     "0.0.0.0",
		Password: "SomePass",
		Method:   "aes-256-cfb",
		Timeout:  60,
	}
	for _, port := range ports {
		s.Port = port
		err := mgr.Add(s)
		if err != nil {
			if err == ss.ErrServerExists {
				log.Warnf("Server(%d) already exists", port)
			} else {
				log.Panicf("Can not create a new ss server, %s", err)
			}
		}
		log.Infof("Adding ss server: %s", s)
	}
}

func contextMain(ctx context.Context) {
	defer log.Infof("Graceful shutdown!")
	select {
	case <-ctx.Done():
		return
	default:
	}
	mgr := ss.NewManager(6001)
	defer mgr.CleanUp()
	err := mgr.Restore()
	if err != nil {
		log.Warn(err)
	}

	if err := mgr.Listen(context.Background()); err != nil {
		log.Panicln("Can not listen udp address 127.0.0.1:6001, ", err)
	}

	ports := []int32{8001, 8002, 8003, 8004, 8005}
	addServers(mgr, ports...)

	s, err := mgr.GetServer(8001)
	if err != nil {
		log.Panicln(err)
	}
	log.Info("Running ss servers: ", mgr.ListServers())
	log.Info("Waiting for 2s")
	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		log.Info("Removing ss server(8001)")
		if err := mgr.Remove(8001); err != nil {
			log.Panicln("Can not remove server, ", err)
		}
		time.Sleep(100 * time.Millisecond)
		if s.Alive() {
			log.Debugf("Server is not supposed to be alive after remove action")

		}
	}
	<-ctx.Done()
}

func main() {
	log.SetLevel(log.DebugLevel)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill)
		<-ch
		cancel()
	}()
	contextMain(ctx)
}
