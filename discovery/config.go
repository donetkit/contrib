package discovery

type Config struct {
	Id                 string
	Name               string
	RegisterAddr       string
	RegisterPort       int
	CheckAddr          string
	CheckPort          int
	Tags               []string
	IntervalTime       int // 健康检查间隔
	DeregisterTime     int //check失败后30秒删除本服务，注销时间，相当于过期时间
	TimeOut            int
	HttpRouter         HttpRouter
	CheckHealthyStatus bool
	CheckResponse      *CheckResponse
	CheckType          string // 检查类型 HTTP TCP GRPC
	CheckPath          string
}
