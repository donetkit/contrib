package systemsignal

import (
	"fmt"
	"github.com/donetkit/contrib/server"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HookSignals(host server.IService) {
	quitSig := make(chan os.Signal)
	signal.Notify(
		quitSig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGSTOP,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGKILL,
	)

	go func() {
		var sig os.Signal
		for {
			sig = <-quitSig
			fmt.Println()
			switch sig {
			// graceful stop
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT:
				host.StopNotify()
				host.Shutdown()
				// terminate now
			}
			time.Sleep(time.Second * 15)
		}
	}()
}
