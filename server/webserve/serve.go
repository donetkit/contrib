package webserve

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
	"github.com/shirou/gopsutil/v3/host"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	exit            chan struct{}
	Ctx             context.Context
	Tracer          *tracer.Server
	Logger          glog.ILoggerEntry
	ServiceName     string
	Host            string
	Port            int
	handler         http.Handler
	httpServer      http.Server
	clientDiscovery discovery.Discovery
	readTimeout     time.Duration
	writerTimeout   time.Duration
	maxHeaderBytes  int
	Version         string
	protocol        string
	pId             int
	environment     string
	runMode         string
}

func New(opts ...Option) *Server {
	var server = &Server{
		exit:           make(chan struct{}),
		Ctx:            context.Background(),
		ServiceName:    "demo",
		Host:           chost.GetOutBoundIp(),
		Port:           80,
		Version:        "V0.1",
		protocol:       "HTTP API",
		pId:            os.Getpid(),
		environment:    server2.EnvName,
		writerTimeout:  time.Second * 120,
		readTimeout:    time.Second * 120,
		maxHeaderBytes: 1 << 20,
	}
	for _, opt := range opts {
		opt(server)
	}
	if server.Logger == nil {
		server.Logger = glog.New().WithField("WebServe", "WebServe")
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

func (s *Server) AddHandler(handler http.Handler) *Server {
	s.handler = handler
	return s
}

func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	s.httpServer = http.Server{
		Addr:           addr,
		Handler:        s.handler,
		ReadTimeout:    s.readTimeout,
		WriteTimeout:   s.writerTimeout,
		MaxHeaderBytes: s.maxHeaderBytes,
	}
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // 平滑关闭,等待5秒钟处理
	defer cancel()
	if err := s.deregister(); err != nil {
		s.Logger.Error("deregister http webserve error", err.Error())
	}
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.Logger.Error("shutdown http webserve error", err.Error())
	}
	s.Logger.Info("Server is stopped.")
}

func (s *Server) printLog() {
	fmt.Println("======================================================================")
	host, err := host.Info()
	if err == nil {
		fmt.Println(console_colors.Green("Loading System Info ..."))
		fmt.Println(fmt.Sprintf("hostName                 :  %s", host.Hostname))
		fmt.Println(fmt.Sprintf("upTime                   :  %s", gtime.ResolveTimeSecond(int(host.Uptime))))
		fmt.Println(fmt.Sprintf("bootTime                 :  %s", time.Unix(int64(host.BootTime), 0).Format("2006/01/02 15:04:05")))
		fmt.Println(fmt.Sprintf("procs                    :  %d", host.Procs))
		fmt.Println(fmt.Sprintf("os                       :  %s", host.OS))
		fmt.Println(fmt.Sprintf("platform                 :  %s", host.Platform))
		fmt.Println(fmt.Sprintf("platformFamily           :  %s", host.PlatformFamily))
		fmt.Println(fmt.Sprintf("platformVersion          :  %s", host.PlatformVersion))
		fmt.Println(fmt.Sprintf("kernelVersion            :  %s", host.KernelVersion))
		fmt.Println(fmt.Sprintf("kernelArch               :  %s", host.KernelArch))
		var cpuType = "Virtual"
		if host.VirtualizationSystem == "" {
			cpuType = "Physical"
		}
		fmt.Println(fmt.Sprintf("virtualizationSystem     :  %s", cpuType))
		if host.VirtualizationRole != "" {
			fmt.Println(fmt.Sprintf("virtualizationRole       :  %s", host.VirtualizationRole))
		}
		fmt.Println(fmt.Sprintf("hostId                   :  %s", host.HostID))
	}

	fmt.Println(console_colors.Green(fmt.Sprintf("Welcome to %s, starting application ...", s.ServiceName)))
	fmt.Println(fmt.Sprintf("framework version        :  %s", console_colors.Blue(s.Version)))
	fmt.Println(fmt.Sprintf("serve & protocol         :  %s", console_colors.Green(s.protocol)))
	fmt.Println(fmt.Sprintf("machine host ip          :  %s", console_colors.Blue(s.Host)))
	fmt.Println(fmt.Sprintf("listening on port        :  %s", console_colors.Blue(fmt.Sprintf("%d", s.Port))))
	fmt.Println(fmt.Sprintf("application running pid  :  %s", console_colors.Blue(strconv.Itoa(s.pId))))
	fmt.Println(fmt.Sprintf("application name         :  %s", console_colors.Blue(s.ServiceName)))
	fmt.Println(fmt.Sprintf("application exec path    :  %s", console_colors.Yellow(files.GetCurrentDirectory())))
	fmt.Println(fmt.Sprintf("application environment  :  %s", console_colors.Yellow(console_colors.Blue(s.environment))))
	fmt.Println(fmt.Sprintf("running in %s mode change ( Dev | Test | Prod ) mode by Environment .", console_colors.Red(s.environment)))
	fmt.Println(console_colors.Green("Server is Started."))
	fmt.Println("======================================================================")
}

func (s *Server) PrintHostInfo() {
	s.printLog()
}

func (s *Server) HostInfo() *host.InfoStat {
	host, _ := host.Info()
	return host
}
