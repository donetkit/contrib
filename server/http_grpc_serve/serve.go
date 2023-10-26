package http_grpc_serve

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/pkg/discovery"
	server2 "github.com/donetkit/contrib/server"
	"github.com/donetkit/contrib/server/systemsignal"
	"github.com/donetkit/contrib/tracer"
	"github.com/donetkit/contrib/utils/console_colors"
	"github.com/donetkit/contrib/utils/files"
	"github.com/donetkit/contrib/utils/gtime"
	chost "github.com/donetkit/contrib/utils/host"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/shirou/gopsutil/host"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"math"
	"net"
	"net/http"
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

type Server struct {
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

	credentials credentials.TransportCredentials

	maxReceiveMessageSize int
	maxSendMessageSize    int
	connectionTimeout     time.Duration
	writeBufferSize       int
	readBufferSize        int

	grpcOpts []grpc.ServerOption

	HTTPListener net.Listener
	GRPCListener net.Listener
	httpServer   *http.Server
	Router       *http.ServeMux
	registerHTTP registerFunc
	registerGRPC registerFunc
	ServerMux    *runtime.ServeMux
	tcpMux       cmux.CMux

	readTimeout   time.Duration
	writerTimeout time.Duration
	idleTimeout   time.Duration
}

type registerFunc func(ctx context.Context, s *Server)

func New(opts ...Option) *Server {
	var server = &Server{
		exit:        make(chan struct{}),
		Ctx:         context.Background(),
		ServiceName: "demo",
		Host:        chost.GetOutBoundIp(),
		Port:        80,
		Version:     "V0.1",
		protocol:    "HTTP-GRPC",
		pId:         os.Getpid(),
		environment: server2.EnvName,

		maxReceiveMessageSize: defaultServerMaxReceiveMessageSize,
		maxSendMessageSize:    defaultServerMaxSendMessageSize,

		connectionTimeout: 120 * time.Second,

		writeBufferSize: defaultWriteBufSize,
		readBufferSize:  defaultReadBufSize,
		grpcOpts:        []grpc.ServerOption{},

		readTimeout:   time.Second * 120,
		writerTimeout: time.Second * 120,
	}
	for _, opt := range opts {
		opt(server)
	}
	if server.Logger == nil {
		server.Logger = glog.New().WithField("GrpcServe", "GrpcServe")
	}

	gOpts := []grpc.ServerOption{
		grpc.WriteBufferSize(server.writeBufferSize),
		grpc.ReadBufferSize(server.readBufferSize),
		grpc.ConnectionTimeout(server.connectionTimeout),
		grpc.MaxRecvMsgSize(server.maxReceiveMessageSize),
		grpc.MaxSendMsgSize(server.maxSendMessageSize)}

	if server.credentials != nil {
		gOpts = append(gOpts, grpc.Creds(server.credentials))
	}

	for _, grpcOpt := range server.grpcOpts {
		gOpts = append(gOpts, grpcOpt)
	}
	server.grpcOpts = gOpts

	//cfg.GServer = grpc.NewServer(gOpts...)

	server.Router = http.NewServeMux()
	server.ServerMux = runtime.NewServeMux()

	return server
}

func (s *Server) Stop() {
	s.tcpMux.Close()
}

func (s *Server) start() {
	s.Router.Handle("/", s.ServerMux)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Host, s.Port),
		Handler:      s.Router,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writerTimeout,
		IdleTimeout:  s.idleTimeout,
	}

	if err := s.httpServer.Serve(s.HTTPListener); err != nil {
		panic(err)
	}
}

func (s *Server) IsDevelopment() bool {
	return s.environment == server2.Dev
}

func (s *Server) IsTest() bool {
	return s.environment == server2.Test
}

func (s *Server) IsProduction() bool {
	return s.environment == server2.Prod
}

func (s *Server) registerDiscovery() *Server {
	if s.clientDiscovery == nil {
		return nil
	}
	err := s.clientDiscovery.Register()
	if err != nil {
		s.Logger.Error(err.Error())
	}
	return s
}

func (s *Server) deregister() error {
	if s.clientDiscovery == nil {
		return nil
	}
	err := s.clientDiscovery.Deregister()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) AddDiscovery(client discovery.Discovery) *Server {
	if client == nil {
		return nil
	}
	s.clientDiscovery = client
	return s
}

func (s *Server) AddTrace(tracer *tracer.Server) *Server {
	s.Tracer = tracer
	return s
}

func (s *Server) NewServer() *Server {
	s.GServer = grpc.NewServer(s.grpcOpts...)
	return s
}

func (s *Server) Run() {
	if s.GServer == nil {
		s.NewServer()
	}
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.Logger.Error(err)
		os.Exit(0)
	}
	s.tcpMux = cmux.New(listener)

	s.GRPCListener = s.tcpMux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	s.HTTPListener = s.tcpMux.Match(cmux.HTTP1Fast())

	go func() {
		s.registerGRPC(s.Ctx, s)
	}()

	go func() {
		s.registerHTTP(s.Ctx, s)
		s.start()
	}()

	go func() {
		if err := s.tcpMux.Serve(); err != nil && err != grpc.ErrServerStopped {
			return
		}
	}()
	s.registerDiscovery()
	s.printLog()
	systemsignal.HookSignals(s)
	//s.awaitSignal()
	select {
	case _ = <-s.exit:
		os.Exit(0)
	}
}

func (s *Server) SetRunMode(mode string) {
	s.runMode = mode
}

func (s *Server) StopNotify(sig os.Signal) {
	s.Logger.Info("receive a signal, " + "signal: " + sig.String())
	s.stop()
	if s.Tracer != nil {
		s.Tracer.Stop(s.Ctx)
	}
}

func (s *Server) Shutdown() {
	close(s.exit)
}

func (s *Server) stop() {
	s.Logger.Info("Server is stopping")
	if err := s.deregister(); err != nil {
		s.Logger.Error("deregister http webserve error", err.Error())
	}
	s.GServer.GracefulStop()
	//time.Sleep(time.Second * 3)
	s.Logger.Info("Server is stopped.")
}

func (s *Server) printLog() {
	s.Logger.Info("======================================================================")
	host, err := host.Info()
	if err == nil {
		s.Logger.Info(console_colors.Green("Loading System Info ..."))
		s.Logger.Info(fmt.Sprintf("hostName                 :  %s", host.Hostname))
		s.Logger.Info(fmt.Sprintf("upTime                   :  %s", gtime.ResolveTimeSecond(int(host.Uptime))))
		s.Logger.Info(fmt.Sprintf("bootTime                 :  %s", time.Unix(int64(host.BootTime), 0).Format("2006/01/02 15:04:05")))
		s.Logger.Info(fmt.Sprintf("procs                    :  %d", host.Procs))
		s.Logger.Info(fmt.Sprintf("os                       :  %s", host.OS))
		s.Logger.Info(fmt.Sprintf("platform                 :  %s", host.Platform))
		s.Logger.Info(fmt.Sprintf("platformFamily           :  %s", host.PlatformFamily))
		s.Logger.Info(fmt.Sprintf("platformVersion          :  %s", host.PlatformVersion))
		s.Logger.Info(fmt.Sprintf("kernelVersion            :  %s", host.KernelVersion))
		s.Logger.Info(fmt.Sprintf("kernelArch               :  %s", host.KernelArch))
		var cpuType = "Virtual"
		if host.VirtualizationSystem == "" {
			cpuType = "Physical"
		}
		s.Logger.Info(fmt.Sprintf("virtualizationSystem     :  %s", cpuType))
		if host.VirtualizationRole != "" {
			s.Logger.Info(fmt.Sprintf("virtualizationRole       :  %s", host.VirtualizationRole))
		}
		s.Logger.Info(fmt.Sprintf("hostId                   :  %s", host.HostID))
	}
	s.Logger.Info(console_colors.Green(fmt.Sprintf("Welcome to %s, starting application ...", s.ServiceName)))
	s.Logger.Info(fmt.Sprintf("framework version        :  %s", console_colors.Blue(s.Version)))
	s.Logger.Info(fmt.Sprintf("serve & protocol         :  %s", console_colors.Green(s.protocol)))
	s.Logger.Info(fmt.Sprintf("machine host ip          :  %s", console_colors.Blue(s.Host)))
	s.Logger.Info(fmt.Sprintf("listening on port        :  %s", console_colors.Blue(fmt.Sprintf("%d", s.Port))))
	s.Logger.Info(fmt.Sprintf("application running pid  :  %s", console_colors.Blue(strconv.Itoa(s.pId))))
	s.Logger.Info(fmt.Sprintf("application name         :  %s", console_colors.Blue(s.ServiceName)))
	s.Logger.Info(fmt.Sprintf("application exec path    :  %s", console_colors.Yellow(files.GetCurrentDirectory())))
	s.Logger.Info(fmt.Sprintf("application environment  :  %s", console_colors.Yellow(console_colors.Blue(s.environment))))
	s.Logger.Info(fmt.Sprintf("running in %s mode change ( Dev | Test | Prod ) mode by Environment .", console_colors.Red(s.environment)))
	s.Logger.Info(console_colors.Green("Server is Started."))
	s.Logger.Info("======================================================================")
}

// AddGrpcServerOptions set grpc.ServerOption function
func (s *Server) AddGrpcServerOptions(grpcOpts ...grpc.ServerOption) *Server {
	for _, grpcOpt := range grpcOpts {
		s.grpcOpts = append(s.grpcOpts, grpcOpt)
	}
	return s
}

// AddGrpcServerOption set grpc.ServerOption function
func (s *Server) AddGrpcServerOption(grpcOpt grpc.ServerOption) *Server {
	s.grpcOpts = append(s.grpcOpts, grpcOpt)
	return s
}

// AddHandler set handler function
func (s *Server) AddHandler(pattern string, handler http.Handler) *Server {
	s.Router.Handle(pattern, handler)
	return s
}

func (s *Server) PrintHostInfo() {
	s.printLog()
}

func (s *Server) HostInfo() *host.InfoStat {
	host, _ := host.Info()
	return host
}
