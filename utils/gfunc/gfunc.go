package gfunc

import (
	"context"
	"fmt"
	"time"
)

func AsyncCallValue(ctx context.Context, value interface{}, f func(value interface{}) interface{}, timeOut ...time.Duration) (interface{}, error) {
	timeout := time.Second * 1
	if len(timeOut) > 0 {
		timeout = timeOut[0]
	}
	done := make(chan interface{}, 1)
	var result interface{}
	go func(ctx context.Context, value interface{}) {
		done <- f(value)
	}(ctx, value)
	select {
	case result = <-done:
		return result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("func time out")
	}
}

func AsyncCall(ctx context.Context, value interface{}, f func(value interface{}), timeOut ...time.Duration) error {
	timeout := time.Second * 1
	if len(timeOut) > 0 {
		timeout = timeOut[0]
	}
	done := make(chan struct{}, 1)
	go func(ctx context.Context, value interface{}) {
		f(value)
		done <- struct{}{}
	}(ctx, value)
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("func time out")
	}
}
