package gorm

import (
	"database/sql"
	"fmt"
	"github.com/donetkit/contrib/db/db_sql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"strings"
)

var dbRowsAffected = attribute.Key("db.rows_affected")

func NewPlugin(plugin *Config) *Config {
	return plugin
}

func (p *Config) Name() string {
	return "gorm"
}

type gormHookFunc func(tx *gorm.DB)

type gormRegister interface {
	Register(name string, fn func(*gorm.DB)) error
}

func (p *Config) Initialize(db *gorm.DB) (err error) {
	if !p.ExcludeMetrics {
		if db, ok := db.ConnPool.(*sql.DB); ok {
			db_sql.ReportDBStatsMetrics(db)
		}
	}
	//cb := db.Callback()
	hooks := []struct {
		callback gormRegister
		hook     gormHookFunc
		name     string
	}{
		// before hooks
		{db.Callback().Create().Before("gorm:before_create"), p.before(opCreate), beforeName("create")},
		{db.Callback().Query().Before("gorm:query"), p.before(opQuery), beforeName("query")},
		{db.Callback().Delete().Before("gorm:before_delete"), p.before(opDelete), beforeName("delete")},
		{db.Callback().Update().Before("gorm:before_update"), p.before(opUpdate), beforeName("update")},
		{db.Callback().Row().Before("gorm:row"), p.before(opQuery), beforeName("row")},
		{db.Callback().Raw().Before("gorm:raw"), p.before(opQuery), beforeName("raw")},

		// after hooks
		{db.Callback().Create().After("gorm:after_create"), p.after(opCreate), afterName("create")},
		{db.Callback().Query().After("gorm:after_query"), p.after(opQuery), afterName("select")},
		{db.Callback().Delete().After("gorm:after_delete"), p.after(opDelete), afterName("delete")},
		{db.Callback().Update().After("gorm:after_update"), p.after(opUpdate), afterName("update")},
		{db.Callback().Row().After("gorm:row"), p.after(opQuery), afterName("row")},
		{db.Callback().Raw().After("gorm:raw"), p.after(opQuery), afterName("raw")},
	}

	var firstErr error

	for _, h := range hooks {
		if err := h.callback.Register("gorm:"+h.name, h.hook); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("callback register %s failed: %w", h.name, err)
		}
	}
	return firstErr
}

func (p *Config) before(operation string) gormHookFunc {
	return func(tx *gorm.DB) {
		if p.TracerServer == nil {
			return
		}
		spanName := p.spanName(tx, strings.ToLower(operation))
		spanName = fmt.Sprintf("db:gorm:%s", spanName)
		tx.Statement.Context, _ = p.TracerServer.Tracer.Start(tx.Statement.Context, spanName, trace.WithSpanKind(trace.SpanKindClient))
	}
}

func (p *Config) after(operation string) gormHookFunc {
	return func(tx *gorm.DB) {
		if p.TracerServer == nil {
			return
		}
		span := trace.SpanFromContext(tx.Statement.Context)
		if !span.IsRecording() {
			return
		}
		defer span.End()
		if tx.Error != nil {
			span.SetStatus(codes.Error, tx.Error.Error())
		}
		attrs := make([]attribute.KeyValue, 0, len(p.Attrs)+4)
		attrs = append(attrs, p.Attrs...)
		if sys := dbSystem(tx); sys.Valid() {
			attrs = append(attrs, sys)
		}
		vars := tx.Statement.Vars
		if p.ExcludeQueryVars {
			vars = make([]interface{}, len(tx.Statement.Vars))
			for i := 0; i < len(vars); i++ {
				vars[i] = "?"
			}
		}
		query := tx.Dialector.Explain(tx.Statement.SQL.String(), vars...)
		attrs = append(attrs, semconv.DBStatementKey.String(p.formatQuery(query)))
		if tx.Statement.Table != "" {
			attrs = append(attrs, semconv.DBSQLTableKey.String(tx.Statement.Table))
		}
		if tx.Statement.RowsAffected != -1 {
			attrs = append(attrs, dbRowsAffected.Int64(tx.Statement.RowsAffected))
		}
		span.SetAttributes(attrs...)

	}
}

func (p *Config) formatQuery(query string) string {
	if p.QueryFormatter != nil {
		return p.QueryFormatter(query)
	}
	return query
}

func dbSystem(tx *gorm.DB) attribute.KeyValue {
	switch tx.Dialector.Name() {
	case "mysql":
		return semconv.DBSystemMySQL
	case "mssql":
		return semconv.DBSystemMSSQL
	case "postgres", "postgresql":
		return semconv.DBSystemPostgreSQL
	case "sqlite":
		return semconv.DBSystemSqlite
	case "sqlserver":
		return semconv.DBSystemKey.String("sqlserver")
	case "clickhouse":
		return semconv.DBSystemKey.String("clickhouse")
	default:
		return attribute.KeyValue{}
	}
}

func beforeName(name string) string {
	return callBackBeforeName + "_" + name
}

func afterName(name string) string {
	return callBackAfterName + "_" + name
}

func (p *Config) spanName(tx *gorm.DB, operation string) string {
	table := ""
	if tx.Statement != nil && tx.Statement.Table != "" {
		table = ":" + tx.Statement.Table
	}
	return fmt.Sprintf("%s%s", operation, table)
}
