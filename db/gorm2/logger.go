package gorm2

import (
	"context"
	"errors"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

var (
	infoStr      = "%s\n[info] "
	warnStr      = "%s\n[warn] "
	errStr       = "%s\n[error] "
	traceStr     = "%s\n[%.3fms] [rows:%v] %s"
	traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
	traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
)

type LogSql struct {
	Logger glog.ILogger
	config *sqlConfig
}

func (l *LogSql) LogMode(level logger.LogLevel) logger.Interface {
	logger := *l
	return &logger
}

func (l *LogSql) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Infof(fmt.Sprintf(infoStr, utils.FileWithLineNum())+msg, data...)
}

func (l *LogSql) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Warningf(fmt.Sprintf(warnStr, utils.FileWithLineNum())+msg, data...)
}

func (l *LogSql) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.Logger == nil {
		return
	}
	l.Logger.Errorf(fmt.Sprintf(errStr, utils.FileWithLineNum())+msg, data...)
}

func (l *LogSql) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.Logger == nil {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && (!errors.Is(err, logger.ErrRecordNotFound) || !l.config.ignoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Logger.Info(fmt.Sprintf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql))
		} else {
			l.Logger.Info(fmt.Sprintf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql))
		}
	case elapsed > l.config.slowThreshold && l.config.slowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.config.slowThreshold)
		if rows == -1 {
			l.Logger.Info(fmt.Sprintf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql))
		} else {
			l.Logger.Info(fmt.Sprintf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql))
		}
	case elapsed > l.config.slowThreshold && l.config.slowThreshold == 0:
		sql, rows := fc()
		if rows == -1 {
			l.Logger.Info(fmt.Sprintf(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql))
		} else {
			l.Logger.Info(fmt.Sprintf(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql))
		}
	}
}
