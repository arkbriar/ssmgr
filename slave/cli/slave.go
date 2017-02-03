package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	proto "github.com/arkbriar/ss-mgr/protocol"
	slave "github.com/arkbriar/ss-mgr/slave"
	ss "github.com/arkbriar/ss-mgr/slave/shadowsocks"
	"google.golang.org/grpc"
)

var (
	port        = flag.Int("port", 6001, "Port of slave service.")
	managerPort = flag.Int("manager-port", 6001, "UDP port (of origin shadowsocks manager) to listen.")
	token       = flag.String("token", "", "Token shared between master and slave.")
	debug       = flag.Bool("debug", false, "Debug mode.'")
)

func validPort(p int) bool {
	return p > 0 && p < 65536
}

func init() {
	flag.Parse()
	if *token == "" {
		log.Error("Token must not be empty.")
		os.Exit(-1)
	}
	if !validPort(*port) {
		log.Error("Invalid port.")
		os.Exit(-1)
	}
	if !validPort(*managerPort) {
		log.Error("Invalid manager port.")
		os.Exit(-1)
	}
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
}

func run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("Early cancel.")
	default:
	}
	mgr := ss.NewManager(*managerPort)
	if err := mgr.Listen(context.Background()); err != nil {
		return fmt.Errorf("Can not listen udp address 127.0.0.1:%d, %s", *managerPort, err)
	}
	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", *managerPort))
	if err != nil {
		return err
	}
	err = mgr.Restore()
	if err != nil {
		log.Warn(err)
	}
	token := *token
	s := grpc.NewServer(grpc.UnaryInterceptor(slave.UnaryAuthInterceptor(token)),
		grpc.StreamInterceptor(slave.StreamAuthInterceptor(token)))
	proto.RegisterSSMgrSlaveServer(s, slave.NewSSMgrSlaveServer(token, mgr))
	errc := make(chan error, 1)
	go func() {
		log.Infof("Starting server on 0.0.0.0:%d", *port)
		errc <- s.Serve(conn)
	}()
	select {
	case <-ctx.Done():
		s.GracefulStop()
		mgr.CleanUp()
		log.Info("Graceful shutdown.")
	case err := <-errc:
		return err
	}
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, os.Kill)
		<-ch
		cancel()
	}()
	if err := run(ctx); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}
