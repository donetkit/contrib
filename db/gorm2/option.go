package gorm2

//
//import (
//	"github.com/donetkit/contrib-log/glog"
//	"github.com/donetkit/contrib/tracer"
//	"go.opentelemetry.io/otel/attribute"
//	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
//	"time"
//)
//
//type Option func(p *config)
//
//// WithDNS  dsn
//func WithDNS(dsn map[string]string) Option {
//	return func(p *config) {
//		p.dsn = dsn
//	}
//}
//
//// WithSlowThreshold  slowThreshold
//func WithSlowThreshold(slowThreshold time.Duration) Option {
//	return func(p *config) {
//		p.sqlConfig.slowThreshold = slowThreshold
//	}
//}
//
//// WithIgnoreRecordNotFoundError  ignoreRecordNotFoundError
//func WithIgnoreRecordNotFoundError(ignoreRecordNotFoundError bool) Option {
//	return func(p *config) {
//		p.sqlConfig.ignoreRecordNotFoundError = ignoreRecordNotFoundError
//	}
//}
//
//// WithTracer  tracerServer tracer.Server
//func WithTracer(tracerServer *tracer.Server) Option {
//	return func(p *config) {
//		p.sqlConfig.tracerServer = tracerServer
//	}
//}
//
//// WithAttributes configures attributes that are used to create a span.
//func WithAttributes(attrs ...attribute.KeyValue) Option {
//	return func(p *config) {
//		p.sqlConfig.attrs = append(p.sqlConfig.attrs, attrs...)
//	}
//}
//
//// WithDBName configures a db.name attribute.
//func WithDBName(name string) Option {
//	return func(p *config) {
//		p.sqlConfig.attrs = append(p.sqlConfig.attrs, semconv.DBNameKey.String(name))
//	}
//}
//
//// WithoutQueryVariables configures the db.statement attribute to exclude query variables
//func WithoutQueryVariables() Option {
//	return func(p *config) {
//		p.sqlConfig.excludeQueryVars = true
//	}
//}
//
//// WithQueryFormatter configures a query formatter
//func WithQueryFormatter(queryFormatter func(query string) string) Option {
//	return func(p *config) {
//		p.sqlConfig.queryFormatter = queryFormatter
//	}
//}
//
//// WithoutMetrics prevents DBStats metrics from being reported.
//func WithoutMetrics() Option {
//	return func(p *config) {
//		p.sqlConfig.excludeMetrics = true
//	}
//}
//
//// WithLogger prevents logger.
//func WithLogger(logger glog.ILoggerEntry) Option {
//	return func(p *config) {
//		p.sqlConfig.logger = logger
//	}
//}
//
//// WithDefaultStringSize  defaultStringSize
//func WithDefaultStringSize(defaultStringSize uint) Option {
//	return func(p *config) {
//		p.sqlConfig.defaultStringSize = defaultStringSize
//	}
//}
//
//// WithDisableDatetimePrecision  disableDatetimePrecision
//func WithDisableDatetimePrecision(disableDatetimePrecision bool) Option {
//	return func(p *config) {
//		p.sqlConfig.disableDatetimePrecision = disableDatetimePrecision
//	}
//}
//
//// WithDontSupportRenameIndex  dontSupportRenameIndex
//func WithDontSupportRenameIndex(dontSupportRenameIndex bool) Option {
//	return func(p *config) {
//		p.sqlConfig.dontSupportRenameIndex = dontSupportRenameIndex
//	}
//}
//
//// WithDontSupportRenameColumn  dontSupportRenameColumn
//func WithDontSupportRenameColumn(dontSupportRenameColumn bool) Option {
//	return func(p *config) {
//		p.sqlConfig.dontSupportRenameColumn = dontSupportRenameColumn
//	}
//}
//
//// WithSkipInitializeWithVersion  skipInitializeWithVersion
//func WithSkipInitializeWithVersion(skipInitializeWithVersion bool) Option {
//	return func(p *config) {
//		p.sqlConfig.skipInitializeWithVersion = skipInitializeWithVersion
//	}
//}
//
//// WithConnMaxIdleTime  connMaxIdleTime
//func WithConnMaxIdleTime(connMaxIdleTime time.Duration) Option {
//	return func(p *config) {
//		p.sqlConfig.connMaxIdleTime = connMaxIdleTime
//	}
//}
//
//// WithMaxOpenCons  maxOpenCons
//func WithMaxOpenCons(maxOpenCons int) Option {
//	return func(p *config) {
//		p.sqlConfig.maxOpenCons = maxOpenCons
//	}
//}
//
//// WithMaxIdleCons  maxIdleCons
//func WithMaxIdleCons(maxIdleCons int) Option {
//	return func(p *config) {
//		p.sqlConfig.maxIdleCons = maxIdleCons
//	}
//}
