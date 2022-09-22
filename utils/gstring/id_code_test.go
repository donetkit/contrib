package gstring

import (
	"testing"
	"time"
)

func TestIDCode(t *testing.T) {
	id := 1000009

	for i := 0; i <= 100; i += 10 {
		id += i
		code := GenCodeBase34(uint64(id))

		value, err := CodeToIDBase34(code)
		if err == nil && int(value) == id {
			t.Logf("id:%v code:%v pass", id, string(code))
		} else {
			t.Errorf("id:%v code:%v failed", id, string(code))
		}
	}
}

func TestIDCodeTimespan(t *testing.T) {
	id := time.Now().Unix()
	code := GenCodeReverse(id)
	value, err := CodeToIDReverse(code)
	if err == nil && value == id {
		t.Logf("id:%v code:%v pass", id, code)
	} else {
		t.Errorf("id:%v code:%v failed", id, code)
	}

}
