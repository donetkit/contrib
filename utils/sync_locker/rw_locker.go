package sync_locker

import (
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"sync"
	"time"
)

// RWLocker 读写锁对象
type RWLocker struct {
	read      int
	write     int // 使用int而不是bool值的原因，是为了与read保持类型的一致；
	prevStack []byte
	mutex     sync.Mutex
	logger    *logrus.Entry
}

// 尝试加写锁
// 返回值：加写锁是否成功
func (l *RWLocker) lock() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 如果已经被锁定，则返回失败
	if l.write == 1 || l.read > 0 {
		return false
	}

	// 否则，将写锁数量设置为１，并返回成功
	l.write = 1

	// 记录Stack信息
	l.prevStack = debug.Stack()

	return true
}

// Lock 写锁定
// timeout:超时毫秒数,timeout<=0则将会死等
// 返回值：
// 成功或失败
// 如果失败，返回上一次成功加锁时的堆栈信息
// 如果失败，返回当前的堆栈信息
func (l *RWLocker) Lock(timeout int) (successful bool, prevStack string, currStack string) {
	timeout = getTimeout(timeout)

	// 遍历指定的次数（即指定的超时时间）
	for i := 0; i < timeout; i = i + conLockSleepMillisecond {
		// 如果锁定成功，则返回成功
		if l.lock() {
			successful = true
			break
		}

		// 如果锁定失败，则休眠con_Lock_Sleep_Millisecond ms，然后再重试
		time.Sleep(conLockSleepMillisecond * time.Millisecond)
	}

	// 如果时间结束仍然是失败，则返回上次成功的堆栈信息，以及当前的堆栈信息
	if successful == false {
		prevStack = string(l.prevStack)
		currStack = string(debug.Stack())
	}

	return
}

// WaitLock 写锁定(死等)
func (l *RWLocker) WaitLock() {
	successful, prevStack, currStack := l.Lock(0)
	if successful == false {
		fmt.Printf("RWLocker:WaitLock():{PrevStack:%s, currStack:%s}\n", prevStack, currStack)
	}
}

// Unlock 释放写锁
func (l *RWLocker) Unlock() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.write = 0
}

// 尝试加读锁
// 返回值：加读锁是否成功
func (l *RWLocker) getLock() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 如果已经被锁定，则返回失败
	if l.write == 1 {
		return false
	}

	// 否则，将读锁数量加１，并返回成功
	l.read += 1

	// 记录Stack信息
	l.prevStack = debug.Stack()

	return true
}

// RLock 读锁定
// timeout:超时毫秒数,timeout<=0则将会死等
// 返回值：
// 成功或失败
// 如果失败，返回上一次成功加锁时的堆栈信息
// 如果失败，返回当前的堆栈信息
func (l *RWLocker) RLock(timeout int) (successful bool, prevStack string, currStack string) {
	timeout = getTimeout(timeout)

	// 遍历指定的次数（即指定的超时时间）
	// 读锁比写锁优先级更低，所以每次休眠2ms，所以尝试的次数就是时间/2
	for i := 0; i < timeout; i = i + conBlockSleepMillisecond {
		// 如果锁定成功，则返回成功
		if l.lock() {
			successful = true
			break
		}

		// 如果锁定失败，则休眠2ms，然后再重试
		time.Sleep(conBlockSleepMillisecond * time.Millisecond)
	}

	// 如果时间结束仍然是失败，则返回上次成功的堆栈信息，以及当前的堆栈信息
	if successful == false {
		prevStack = string(l.prevStack)
		currStack = string(debug.Stack())
	}

	return
}

// WaitRLock 读锁定(死等)
func (l *RWLocker) WaitRLock() {
	successful, prevStack, currStack := l.RLock(0)
	if successful == false {
		l.logger.Debugf("RWLocker:WaitRLock(): {PrevStack:%s, currStack:%s}", prevStack, currStack)
	}
}

// RUnlock 释放读锁
func (l *RWLocker) RUnlock() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.read > 0 {
		l.read -= 1
	}
}

// NewRWLocker 创建新的读写锁对象
func NewRWLocker(logger glog.Logger) *RWLocker {
	return &RWLocker{
		logger: logger.WithField("RWLocker", "RWLocker"),
	}
}
