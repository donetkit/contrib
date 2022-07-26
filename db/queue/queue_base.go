package queue

type QueueBase struct {

	// 追踪名。默认Key，主要用于解决动态Topic导致产生大量埋点的问题
	TraceName string

	//是否在消息报文中自动注入TraceId。TraceId用于跨应用在生产者和消费者之间建立调用链，默认true
	AttachTraceId string

	// 失败时抛出异常。默认false
	ThrowOnFailure bool

	//发送消息失败时的重试次数。默认3次
	RetryTimesWhenSendFailed int

	// 重试间隔。默认1000ms
	RetryIntervalWhenSendFailed int

	//消息队列主题
	Topic string
}
