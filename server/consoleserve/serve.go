package consoleserve

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	server2 "github.com/donetkit/contrib/server"
	"github.com/donetkit/contrib/server/systemsignal"
	"github.com/donetkit/contrib/tracer"
	"github.com/donetkit/contrib/utils/console_colors"
	"github.com/donetkit/contrib/utils/files"
	"github.com/donetkit/contrib/utils/gtime"
	"github.com/shirou/gopsutil/v3/host"
	"os"
	"strconv"
	"time"
)

type Server struct {
	exit        chan struct{}
	Ctx         context.Context
	Tracer      *tracer.Server
	Logger      glog.ILoggerEntry
	ServiceName string
	Version     string
	protocol    string
	pId         int
	environment string
	runMode     string
}

func New(opts ...Option) *Server {
	server := &Server{
		exit:        make(chan struct{}),
		Ctx:         context.Background(),
		ServiceName: "demo",
		Version:     "V0.1",
		protocol:    "serve",
		pId:         os.Getpid(),
		environment: server2.EnvName,
	}
	for _, opt := range opts {
		opt(server)
	}
	if server.Logger == nil {
		server.Logger = glog.New().WithField("Serve", "Serve")
	}
	return server
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

func (s *Server) AddTrace(tracer *tracer.Server) *Server {
	s.Tracer = tracer
	return s
}

func (s *Server) Run() {
	s.printLog()
	systemsignal.HookSignals(s)
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
	if err := s.stop(); err != nil {
		s.Logger.Error("stop http webserve error %s", err.Error())
	}
	if s.Tracer != nil {
		s.Tracer.Stop(s.Ctx)
	}
}

func (s *Server) Shutdown() {
	close(s.exit)
}

func (s *Server) stop() error {
	s.Logger.Info("Server is stopping")
	_, cancel := context.WithTimeout(context.Background(), time.Second*5) // 平滑关闭,等待5秒钟处理
	defer cancel()
	s.Logger.Info("Server is stopped.")
	return nil
}

func (s *Server) printLog() {
	s.Logger.Info("======================================================================")
	host, err := host.Info()
	if err == nil {
		s.Logger.Info(console_colors.Green("Loading System Info ..."))
		s.Logger.Info(fmt.Sprintf("hostname                 :  %s", host.Hostname))
		s.Logger.Info(fmt.Sprintf("uptime                   :  %s", gtime.ResolveTimeSecond(int(host.Uptime))))
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
	s.Logger.Info(fmt.Sprintf("application running pid  :  %s", console_colors.Blue(strconv.Itoa(s.pId))))
	s.Logger.Info(fmt.Sprintf("application name         :  %s", console_colors.Blue(s.ServiceName)))
	s.Logger.Info(fmt.Sprintf("application exec path    :  %s", console_colors.Yellow(files.GetCurrentDirectory())))
	s.Logger.Info(fmt.Sprintf("application environment  :  %s", console_colors.Yellow(console_colors.Blue(s.environment))))
	s.Logger.Info(fmt.Sprintf("running in %s mode , change (Dev,Test,Prod) mode by Environment .", console_colors.Red(s.environment)))
	s.Logger.Info(console_colors.Green("Server is Started."))
	s.Logger.Info("======================================================================")
}
