package gorm2

//
//import (
//	"github.com/donetkit/contrib-log/glog"
//	"github.com/donetkit/contrib/tracer"
//	"go.opentelemetry.io/otel/attribute"
//	"gorm.io/driver/mysql"
//	"gorm.io/gorm"
//	"time"
//)
//
//type Client struct {
//	Client map[string]*gorm.DB
//	config *config
//}
//
//func (c *Client) DB(dbName ...string) (conn *gorm.DB) {
//	if c.Client == nil {
//		return nil
//	}
//	if len(dbName) <= 0 {
//		return c.Client["default"]
//	}
//	return c.Client[dbName[0]]
//}
//
//type config struct {
//	dsn       map[string]string
//	sqlConfig *sqlConfig
//}
//
//type sqlConfig struct {
//	ignoreRecordNotFoundError bool
//	slowThreshold             time.Duration
//	logger                    glog.ILoggerEntry
//	tracerServer              *tracer.Server
//	attrs                     []attribute.KeyValue
//	excludeQueryVars          bool
//	excludeMetrics            bool
//	queryFormatter            func(query string) string
//	defaultStringSize         uint          // string 类型字段的默认长度
//	disableDatetimePrecision  bool          // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
//	dontSupportRenameIndex    bool          // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
//	dontSupportRenameColumn   bool          // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
//	skipInitializeWithVersion bool          // 根据当前 MySQL 版本自动配置
//	connMaxIdleTime           time.Duration // 设置了连接可复用的最大时间
//	maxOpenCons               int           // 设置打开数据库连接的最大数量
//	maxIdleCons               int           // 设置空闲连接池中连接的最大数量
//}
//
//func NewDb(opts ...Option) *Client {
//	c := &config{}
//	c.sqlConfig = &sqlConfig{
//		defaultStringSize:         256,
//		disableDatetimePrecision:  true,
//		dontSupportRenameIndex:    true,
//		dontSupportRenameColumn:   true,
//		skipInitializeWithVersion: false,
//		connMaxIdleTime:           time.Second * 1800,
//		maxOpenCons:               100,
//		maxIdleCons:               20,
//	}
//	for _, opt := range opts {
//		opt(c)
//	}
//	sqlConfig := &gorm.Config{}
//	if c.sqlConfig.logger != nil {
//		sqlConfig.Logger = &LogSql{Logger: c.sqlConfig.logger, config: c.sqlConfig}
//	}
//	var dbs = map[string]*gorm.DB{}
//	for key, val := range c.dsn {
//		db, err := gorm.Open(mysql.New(mysql.Config{
//			DSN:                       val,
//			DefaultStringSize:         256,   // string 类型字段的默认长度
//			DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
//			DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
//			DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
//			SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
//		}), sqlConfig)
//		if err != nil && c.sqlConfig.logger != nil {
//			c.sqlConfig.logger.Error(err)
//		}
//		if c.sqlConfig.tracerServer != nil {
//			if err := db.Use(NewPlugin(c.sqlConfig)); err != nil && c.sqlConfig.logger != nil {
//				c.sqlConfig.logger.Error(err)
//			}
//		}
//		sdb, err := db.DB()
//		if err == nil {
//			sdb.SetConnMaxIdleTime(c.sqlConfig.connMaxIdleTime) //最大生存时间(s) 30 SetConnMaxLifetime 设置了连接可复用的最大时间。
//			sdb.SetMaxOpenConns(c.sqlConfig.maxOpenCons)        // SetMaxOpenConns 设置打开数据库连接的最大数量。
//			sdb.SetMaxIdleConns(c.sqlConfig.maxIdleCons)        //最大连接数 1000 SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
//		}
//
//		dbs[key] = db
//	}
//	return &Client{Client: dbs, config: c}
//}
