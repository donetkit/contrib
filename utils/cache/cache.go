package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

//type ICache interface {
//	WithDB(db int) ICache
//	WithContext(ctx context.Context) ICache
//	Get(string) interface{}
//	GetString(string) (string, error)
//	Set(string, interface{}, time.Duration) error
//	IsExist(string) bool
//	Delete(string) (int64, error)
//	Increment(string, int64) (int64, error)
//	IncrementFloat(string, float64) (float64, error)
//	Decrement(string, int64) (int64, error)
//	Flush()
//}

type IMemoryCache interface {
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
	XRead(key string, startId string, count int64, block int64) []redis.XMessage
	XAdd(key, msgId string, trim bool, maxLength int64, value interface{}) string
	XDel(key string, id ...string) int64
	GetLock(string, time.Duration, time.Duration) (string, error)
	ReleaseLock(string, string) bool

	Increment(string, int64) (int64, error)
	IncrementFloat(string, float64) (float64, error)
	Decrement(string, int64) (int64, error)

	Flush()

	ZAdd(string, float64, interface{}) interface{}
	ZRangeByScore(string, int64, int64) ([]string, error)
	ZRem(string, ...interface{}) error

	XLen(string) int64
	Exists(...string) int64
	XInfoGroups(string) []redis.XInfoGroup
	XGroupCreateMkStream(key string, group string, start string) string
	XGroupDestroy(key string, group string) int64
	XPendingExt(key string, group string, startId string, endId string, count int64, consumer ...string) []redis.XPendingExt
	XPending(key string, group string) *redis.XPending
	XGroupDelConsumer(key string, group string, consumer string) int64
	XGroupSetID(key string, group string, start string) string
	XReadGroup(key string, group string, consumer string, count int64, block int64, id ...string) []redis.XMessage
	XInfoStream(key string) *redis.XInfoStream
	XInfoConsumers(key string, group string) []redis.XInfoConsumer
	Pipeline() redis.Pipeliner

	XClaim(key string, group string, consumer string, id string, msIdle int64) []redis.XMessage
	XAck(key string, group string, ids ...string) int64
	XTrimMaxLen(key string, maxLen int64) int64
	XRangeN(key string, start string, stop string, count int64) []redis.XMessage
	XRange(key string, start string, stop string) []redis.XMessage
}
