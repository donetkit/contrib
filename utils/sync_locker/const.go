package sync_locker

const (
	// 默认超时的毫秒数(1小时)
	conDefaultTimeoutMilliseconds = 60 * 60 * 1000

	// 写锁每次休眠的时间比读锁的更短，这样是因为写锁有更高的优先级，所以尝试的频率更大
	// 写锁每次休眠的毫秒数
	conLockSleepMillisecond = 1

	// 读锁每次休眠的毫秒数
	conBlockSleepMillisecond = 2
)

// 获取超时时间
func getTimeout(timeout int) int {
	if timeout > 0 {
		return timeout
	} else {
		return conDefaultTimeoutMilliseconds
	}
}
