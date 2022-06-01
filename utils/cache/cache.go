package cache

import (
	"context"
	"time"
)

//Save(io.Writer) error
//SaveFile(string) error
//Load(io.Reader) error
//LoadFile(string) error

type ICache interface {
	WithDB(db int) ICache
	WithContext(ctx context.Context) ICache
	Get(string) interface{}
	GetString(string) (string, error)
	Set(string, interface{}, time.Duration) error
	IsExist(string) bool
	Delete(string) (int64, error)
	LPush(string, interface{}) (int64, error)
	RPop(string) interface{}
	XRead(string, int64) (interface{}, error) // default type []redis.XStream
	XAdd(string, string, []string) (string, error)
	XDel(string, string) (int64, error)
	GetLock(string, time.Duration, time.Duration) (string, error)
	ReleaseLock(string, string) bool

	Increment(string, int64) (int64, error)
	IncrementFloat(string, float64) (float64, error)
	Decrement(string, int64) (int64, error)

	Flush()

	ZAdd(string, float64, interface{}) interface{}
	ZRangeByScore(string, int64, int64) ([]string, error)
	ZRem(string, ...interface{}) error
}

type IShortCache interface {
	WithDB(db int) ICache
	WithContext(ctx context.Context) ICache
	Get(string) interface{}
	GetString(string) (string, error)
	Set(string, interface{}, time.Duration) error
	IsExist(string) bool
	Delete(string) (int64, error)
	Increment(string, int64) (int64, error)
	IncrementFloat(string, float64) (float64, error)
	Decrement(string, int64) (int64, error)
	Flush()
}
