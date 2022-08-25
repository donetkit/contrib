package cmuxserve

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/donetkit/contrib/pkg/discovery"
	server2 "github.com/donetkit/contrib/server"
	"github.com/donetkit/contrib/server/ctime"
	"github.com/donetkit/contrib/server/systemsignal"
	"github.com/donetkit/contrib/tracer"
	"github.com/donetkit/contrib/utils/console_colors"
	"github.com/donetkit/contrib/utils/files"
	chost "github.com/donetkit/contrib/utils/host"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/soheilhy/cmux"
	"net"
	"os"
	"strconv"
	"time"
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

	readTimeout   time.Duration
	writerTimeout time.Duration

	CMux cmux.CMux
}

type Server struct {
	Options *config
}

func New(opts ...Option) *Server {
	var cfg = &config{
		exit:          make(chan struct{}),
		Ctx:           context.Background(),
		ServiceName:   "demo",
		Host:          chost.GetOutBoundIp(),
		Port:          80,
		Version:       "V0.1",
		protocol:      "MuxServe",
		pId:           os.Getpid(),
		environment:   server2.EnvName,
		writerTimeout: time.Second * 120,
		readTimeout:   time.Second * 120,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	server := &Server{
		Options: cfg,
	}
	if cfg.Logger == nil {
		cfg.Logger = glog.New().WithField("MuxServe", "MuxServe")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		cfg.Logger.Error(err)
		os.Exit(0)
	}
	cfg.CMux = cmux.New(lis)
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

func (s *Server) Run() {
	go func() {
		s.Options.CMux.SetReadTimeout(s.Options.readTimeout)
		if err := s.Options.CMux.Serve(); err != nil {
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
	s.Options.CMux.Close()
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
