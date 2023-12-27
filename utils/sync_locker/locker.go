package sync_locker

import (
	"github.com/donetkit/contrib-log/glog"
	"github.com/sirupsen/logrus"
	"runtime/debug"
	"sync"
	"time"
)

// Locker 写锁对象
type Locker struct {
	write     int // 使用int而不是bool值的原因，是为了与RWLocker中的read保持类型的一致；
	prevStack []byte
	mutex     sync.Mutex
	logger    *logrus.Entry
}

// 内部锁
// 返回值：
// 加锁是否成功
func (l *Locker) lock() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 如果已经被锁定，则返回失败
	if l.write == 1 {
		return false
	}

	// 否则，将写锁数量设置为１，并返回成功
	l.write = 1

	// 记录Stack信息
	l.prevStack = debug.Stack()

	return true
}

// Lock 尝试加锁，如果在指定的时间内失败，则会返回失败；否则返回成功
// timeout:指定的毫秒数,timeout<=0则将会死等
// 返回值：
// 成功或失败
// 如果失败，返回上一次成功加锁时的堆栈信息
// 如果失败，返回当前的堆栈信息
func (l *Locker) Lock(timeout int) (successful bool, prevStack string, currStack string) {
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

// WaitLock 锁定（死等方式）
func (l *Locker) WaitLock() {
	successful, prevStack, currStack := l.Lock(0)
	if successful == false {
		l.logger.Debugf("Locker.WaitLock(): {PrevStack:%s, currStack:%s}", prevStack, currStack)
	}
}

// Unlock 解锁
func (l *Locker) Unlock() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.write = 0
}

// NewLocker 创建新的锁对象
func NewLocker(logger glog.ILogger) *Locker {
	return &Locker{
		logger: logger.WithField("Locker", "Locker"),
	}
}
