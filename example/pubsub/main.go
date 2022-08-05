package main

import (
	"fmt"
	"github.com/donetkit/contrib/pkg/pubsub"
	"strings"
	"time"
)

func main() {
	p := pubsub.NewPublisher(100*time.Microsecond, 10)
	golang := p.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, "golang:") {
				return true
			}
		}
		return false
	})

	docker := p.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, "docker:") {
				return true
			}
		}
		return false
	})

	go p.Publish("wang")
	go p.Publish("golang: https://golang.org")
	go p.Publish("docker: https://www.docker.com")

	time.Sleep(time.Second * 2)
	go func() {
		fmt.Println("golang topic:", <-golang)
	}()

	go func() {
		fmt.Println("docker topic:", <-docker)
	}()
	time.Sleep(time.Second * 3)
	fmt.Println("end")
}
