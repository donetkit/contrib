package gorm

import (
	"context"
	"errors"
	"fmt"
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
	case err != nil && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Logger.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Errorf(traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Logger.Warnf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Warnf(traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold == 0:
		sql, rows := fc()
		if rows == -1 {
			l.Logger.Warnf(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Logger.Warnf(traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
