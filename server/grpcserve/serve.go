package grpcserve

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/discovery"
	server2 "github.com/donetkit/contrib/server"
	"github.com/donetkit/contrib/server/ctime"
	"github.com/donetkit/contrib/server/systemsignal"
	"github.com/donetkit/contrib/tracer"
	"github.com/donetkit/contrib/utils/console_colors"
	"github.com/donetkit/contrib/utils/files"
	chost "github.com/donetkit/contrib/utils/host"
	"github.com/shirou/gopsutil/v3/host"
	"google.golang.org/grpc"
	"math"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	defaultServerMaxReceiveMessageSize = 1024 * 1024 * 4
	defaultServerMaxSendMessageSize    = math.MaxInt32
	// http2IOBufSize specifies the buffer size for sending frames.
	defaultWriteBufSize = 32 * 1024
	defaultReadBufSize  = 32 * 1024
)

type config struct {
	exit            chan struct{}
	Ctx             context.Context
	Tracer          *tracer.Server
	Logger          glog.ILoggerEntry
	ServiceName     string
	Host            string
	Port            int
	clientDiscovery discovery.Discovery
	Version         string
	protocol        string
	pId             int
	environment     string
	runMode         string
	GServer         *grpc.Server

	maxReceiveMessageSize int
	maxSendMessageSize    int
	connectionTimeout     time.Duration
	writeBufferSize       int
	readBufferSize        int

	grpcOpts []grpc.ServerOption
}

type Server struct {
	Options *config
}

func New(opts ...Option) *Server {
	var cfg = &config{
		exit:        make(chan struct{}),
		Ctx:         context.Background(),
		ServiceName: "demo",
		Host:        chost.GetOutBoundIp(),
		Port:        80,
		Version:     "V0.1",
		protocol:    "GRPC",
		pId:         os.Getpid(),
		environment: server2.EnvName,

		maxReceiveMessageSize: defaultServerMaxReceiveMessageSize,
		maxSendMessageSize:    defaultServerMaxSendMessageSize,

		connectionTimeout: 120 * time.Second,

		writeBufferSize: defaultWriteBufSize,
		readBufferSize:  defaultReadBufSize,
		grpcOpts:        []grpc.ServerOption{},
	}
	for _, opt := range opts {
		opt(cfg)
	}
	server := &Server{
		Options: cfg,
	}
	if cfg.Logger == nil {
		cfg.Logger = glog.New().WithField("GrpcServe", "GrpcServe")
	}

	gOpts := []grpc.ServerOption{
		grpc.WriteBufferSize(cfg.writeBufferSize),
		grpc.ReadBufferSize(cfg.readBufferSize),
		grpc.ConnectionTimeout(cfg.connectionTimeout),
		grpc.MaxRecvMsgSize(cfg.maxReceiveMessageSize),
		grpc.MaxSendMsgSize(cfg.maxSendMessageSize)}

	for _, grpcOpt := range cfg.grpcOpts {
		gOpts = append(gOpts, grpcOpt)
	}
	cfg.grpcOpts = gOpts

	//cfg.GServer = grpc.NewServer(gOpts...)

	return server
}

func (s *Server) IsDevelopment() bool {
	return s.Options.environment == server2.Dev
}

func (s *Server) IsTest() bool {
	return s.Options.environment == server2.Test
}

func (s *Server) IsProduction() bool {
	return s.Options.environment == server2.Prod
}

func (s *Server) registerDiscovery() *Server {
	if s.Options.clientDiscovery == nil {
		return nil
	}
	err := s.Options.clientDiscovery.Register()
	if err != nil {
		s.Options.Logger.Error(err.Error())
	}
	return s
}

func (s *Server) deregister() error {
	if s.Options.clientDiscovery == nil {
		return nil
	}
	err := s.Options.clientDiscovery.Deregister()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) AddDiscovery(client discovery.Discovery) *Server {
	if client == nil {
		return nil
	}
	s.Options.clientDiscovery = client
	return s
}

func (s *Server) AddTrace(tracer *tracer.Server) *Server {
	s.Options.Tracer = tracer
	return s
}

func (s *Server) NewServer() *Server {
	s.Options.GServer = grpc.NewServer(s.Options.grpcOpts...)
	return s
}

func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%d", s.Options.Host, s.Options.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.Options.Logger.Error(err)
		os.Exit(0)
	}
	go func() {
		if err := s.Options.GServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			return
		}
	}()
	s.registerDiscovery()
	s.printLog()
	systemsignal.HookSignals(s)
	//s.awaitSignal()
	select {
	case _ = <-s.Options.exit:
		os.Exit(0)
	}
}

func (s *Server) SetRunMode(mode string) {
	s.Options.runMode = mode
}

func (s *Server) StopNotify(sig os.Signal) {
	s.Options.Logger.Info("receive a signal, " + "signal: " + sig.String())
	s.stop()
	if s.Options.Tracer != nil {
		s.Options.Tracer.Stop(s.Options.Ctx)
	}
}

func (s *Server) Shutdown() {
	close(s.Options.exit)
}

func (s *Server) stop() {
	s.Options.Logger.Info("Server is stopping")
	if err := s.deregister(); err != nil {
		s.Options.Logger.Error("deregister http webserve error", err.Error())
	}
	s.Options.GServer.GracefulStop()
	//time.Sleep(time.Second * 3)
	s.Options.Logger.Info("Server is stopped.")
}

func (s *Server) printLog() {
	s.Options.Logger.Info("======================================================================")
	host, err := host.Info()
	if err == nil {
		s.Options.Logger.Info(console_colors.Green("Loading System Info ..."))
		s.Options.Logger.Info(fmt.Sprintf("hostname                 :  %s", host.Hostname))
		s.Options.Logger.Info(fmt.Sprintf("uptime                   :  %s", ctime.ResolveTimeSecond(int(host.Uptime))))
		s.Options.Logger.Info(fmt.Sprintf("bootTime                 :  %s", time.Unix(int64(host.BootTime), 0).Format("2006/01/02 15:04:05")))
		s.Options.Logger.Info(fmt.Sprintf("procs                    :  %d", host.Procs))
		s.Options.Logger.Info(fmt.Sprintf("os                       :  %s", host.OS))
		s.Options.Logger.Info(fmt.Sprintf("platform                 :  %s", host.Platform))
		s.Options.Logger.Info(fmt.Sprintf("platformFamily           :  %s", host.PlatformFamily))
		s.Options.Logger.Info(fmt.Sprintf("platformVersion          :  %s", host.PlatformVersion))
		s.Options.Logger.Info(fmt.Sprintf("kernelVersion            :  %s", host.KernelVersion))
		s.Options.Logger.Info(fmt.Sprintf("kernelArch               :  %s", host.KernelArch))
		var cpuType = "Virtual"
		if host.VirtualizationSystem == "" {
			cpuType = "Physical"
		}
		s.Options.Logger.Info(fmt.Sprintf("virtualizationSystem     :  %s", cpuType))
		if host.VirtualizationRole != "" {
			s.Options.Logger.Info(fmt.Sprintf("virtualizationRole       :  %s", host.VirtualizationRole))
		}
		s.Options.Logger.Info(fmt.Sprintf("hostId                   :  %s", host.HostID))
	}
	s.Options.Logger.Info(console_colors.Green(fmt.Sprintf("Welcome to %s, starting application ...", s.Options.ServiceName)))
	s.Options.Logger.Info(fmt.Sprintf("framework version        :  %s", console_colors.Blue(s.Options.Version)))
	s.Options.Logger.Info(fmt.Sprintf("serve & protocol         :  %s", console_colors.Green(s.Options.protocol)))
	s.Options.Logger.Info(fmt.Sprintf("machine host ip          :  %s", console_colors.Blue(s.Options.Host)))
	s.Options.Logger.Info(fmt.Sprintf("listening on port        :  %s", console_colors.Blue(fmt.Sprintf("%d", s.Options.Port))))
	s.Options.Logger.Info(fmt.Sprintf("application running pid  :  %s", console_colors.Blue(strconv.Itoa(s.Options.pId))))
	s.Options.Logger.Info(fmt.Sprintf("application name         :  %s", console_colors.Blue(s.Options.ServiceName)))
	s.Options.Logger.Info(fmt.Sprintf("application exec path    :  %s", console_colors.Yellow(files.GetCurrentDirectory())))
	s.Options.Logger.Info(fmt.Sprintf("application environment  :  %s", console_colors.Yellow(console_colors.Blue(s.Options.environment))))
	s.Options.Logger.Info(fmt.Sprintf("running in %s mode , change (Dev,Test,Prod) mode by Environment .", console_colors.Red(s.Options.environment)))
	s.Options.Logger.Info(console_colors.Green("Server is Started."))
	s.Options.Logger.Info("======================================================================")
}

// AddGrpcServerOptions set grpc.ServerOption function
func (s *Server) AddGrpcServerOptions(grpcOpts ...grpc.ServerOption) *Server {
	for _, grpcOpt := range grpcOpts {
		s.Options.grpcOpts = append(s.Options.grpcOpts, grpcOpt)
	}
	return s
}

// AddGrpcServerOption set grpc.ServerOption function
func (s *Server) AddGrpcServerOption(grpcOpt grpc.ServerOption) *Server {
	s.Options.grpcOpts = append(s.Options.grpcOpts, grpcOpt)
	return s
}
