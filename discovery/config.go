package discovery

type Config struct {
	Id                  string
	ServiceName         string
	ServiceRegisterAddr string
	ServiceRegisterPort int
	ServiceCheckAddr    string
	ServiceCheckPort    int
	Tags                []string
	IntervalTime        int // 健康检查间隔
	DeregisterTime      int //check失败后30秒删除本服务，注销时间，相当于过期时间
	TimeOut             int
	CheckHTTP           string
	Router              Router
}
