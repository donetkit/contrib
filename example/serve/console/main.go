package main

import (
	"github.com/donetkit/contrib/server/consoleserve"
)

func main() {
	consoleserve.New().PrintHostInfo()
	consoleserve.New().Run()
}
