package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	proto "github.com/arkbriar/ssmgr/protocol"
	slave "github.com/arkbriar/ssmgr/slave"
	ss "github.com/arkbriar/ssmgr/slave/shadowsocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	config  = flag.String("c", "config.json", "Config file")
	verbose = flag.Bool("v", false, "Verbose mode")
)

func validPort(p int) bool {
	return p > 0 && p < 65536
}

type slaveConfig struct {
	Port    int    `json:"port,omitemtpy"`
	MgrPort int    `json:"manager_port,omitempty"`
	Token   string `json:"token"`
	TLS     *struct {
		CertFile string `json:"cert_file"`
		KeyFile  string `json:"key_file"`
	} `json:"tls,omitempty"`
}

// Global configuration object
var conf *slaveConfig

func parseConfig(filename string) (*slaveConfig, error) {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	c := &slaveConfig{Port: 8001, MgrPort: 6001}
	if err := json.Unmarshal(d, c); err != nil {
		return nil, err
	}
	return c, nil
}

func checkConfig(c *slaveConfig) error {
	if len(c.Token) == 0 {
		return errors.New("invalid token")
	}
	return nil
}

func init() {
	flag.Parse()
	if c, err := parseConfig(*config); err != nil {
		log.Error(err)
		os.Exit(-1)
	} else {
		conf = c
	}
	if err := checkConfig(conf); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("early cancel")
	default:
	}

	mgr := ss.NewManager(conf.MgrPort)
	if err := mgr.Listen(context.Background()); err != nil {
		return fmt.Errorf("Can not listen udp address 127.0.0.1:%d, %s", conf.MgrPort, err)
	}

	token := conf.Token
	serverOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(slave.UnaryAuthInterceptor(token)),
		grpc.StreamInterceptor(slave.StreamAuthInterceptor(token)),
	}

	// enable grpc channel with credentials

	if conf.TLS != nil {
		log.Info("Encrypting grpc channel with TLS")

		cred, err := credentials.NewServerTLSFromFile(conf.TLS.CertFile, conf.TLS.KeyFile)
		if err != nil {
			return err
		}
		serverOpts = append(serverOpts, grpc.Creds(cred))
	}

	s := grpc.NewServer(serverOpts...)
	proto.RegisterSSMgrSlaveServer(s, slave.NewSSMgrSlaveServer(token, mgr))

	// listen and do the restoration

	conn, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		return err
	}
	err = mgr.Restore()
	if err != nil {
		log.Warn(err)
	}

	// start rpc server

	errc := make(chan error, 1)
	go func() {
		log.Infof("Starting server on 0.0.0.0:%d", conf.Port)

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
