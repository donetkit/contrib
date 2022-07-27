package queue

type QueueBase struct {

	// 失败时抛出异常。默认false
	ThrowOnFailure bool

	//发送消息失败时的重试次数。默认3次
	RetryTimesWhenSendFailed int

	// 重试间隔。默认1000ms
	RetryIntervalWhenSendFailed int

	//消息队列主题
	Topic string
}
