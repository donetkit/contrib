package systemsignal

import (
	"github.com/donetkit/gin-contrib/server"
	"os"
	"os/signal"
	"syscall"
)

func HookSignals(host server.IService) {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	select {
	case quit := <-sig:
		host.StopNotify(quit)
		host.Shutdown()
	}
}
