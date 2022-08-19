package gfunc

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestAsyncCallValue(t *testing.T) {
	ctx := context.Background()
	val, err := AsyncCallValue(ctx, nil, funcTestName, time.Millisecond*900)
	if err != nil {
		t.Error(err)
	}
	var index = val.(int64)
	t.Log(index)

}

func TestAsyncCall(t *testing.T) {
	ctx := context.Background()
	err := AsyncCall(ctx, nil, funcTest, time.Millisecond*1000)
	if err != nil {
		t.Error(err)
	}
}

func funcTestName(value interface{}) interface{} {
	fmt.Println("funcTestName", value)
	time.Sleep(time.Millisecond * 100)
	return time.Now().UnixMilli()
}

func funcTest(value interface{}) {
	fmt.Println("funcTest", value)
	time.Sleep(time.Millisecond * 100)
}
