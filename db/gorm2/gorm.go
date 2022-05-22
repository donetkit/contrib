package gorm2

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/donetkit/contrib/db/db_sql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"strings"
)

var dbRowsAffected = attribute.Key("db.rows_affected")

func NewPlugin(plugin *sqlConfig) *sqlConfig {
	return plugin
}

func (p *sqlConfig) Name() string {
	return "gorm"
}

type gormHookFunc func(tx *gorm.DB)

type gormRegister interface {
	Register(name string, fn func(*gorm.DB)) error
}

func (p *sqlConfig) Initialize(db *gorm.DB) (err error) {
	if !p.excludeMetrics {
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

func (p *sqlConfig) before(operation string) gormHookFunc {
	return func(tx *gorm.DB) {
		//var name string
		if p.tracerServer == nil {
			return
		}
		//var name1 = tx.Dialector.Name()
		//fmt.Println(name1)
		//if tx.Statement.Table != "" {
		//	name = fmt.Sprintf("db:gorm:%s:%s", tx.Statement.Table, spanName)
		//} else {
		//	name = fmt.Sprintf("db:gorm:%s", spanName)
		//}
		tx.Statement.Context, _ = p.tracerServer.Tracer.
			Start(tx.Statement.Context, p.spanName(tx, operation), trace.WithSpanKind(trace.SpanKindClient))
	}
}

func (op *sqlConfig) after(operation string) gormHookFunc {
	return func(tx *gorm.DB) {
		if op.tracerServer == nil {
			return
		}
		span := trace.SpanFromContext(tx.Statement.Context)
		if !span.IsRecording() {
			// skip the reporting if not recording
			return
		}
		defer span.End()

		span.SetName(op.spanName(tx, operation))
		// Error
		if tx.Error != nil {
			span.SetStatus(codes.Error, tx.Error.Error())
		}
		// extract the db operation
		query := strings.ToValidUTF8(extractQuery(tx), "")
		// If query is longer then max size log it as chunked event, otherwise log it in attribute
		if len(query) > eventMaxSize {
			chunkBy(query, eventMaxSize, span.AddEvent)
		} else {
			span.SetAttributes(dbStatement(query))
		}

		operation = operationForQuery(query, operation)
		if tx.Statement.Table != "" {
			span.SetAttributes(dbTable(tx.Statement.Table))
		}

		span.SetAttributes(
			dbOperation(operation),
			dbCount(tx.Statement.RowsAffected),
		)
	}
}

func (p *sqlConfig) formatQuery(query string) string {
	if p.queryFormatter != nil {
		return p.queryFormatter(query)
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

func extractQuery(tx *gorm.DB) string {
	if shouldOmit, _ := tx.Statement.Context.Value(omitVarsKey).(bool); shouldOmit {
		return tx.Statement.SQL.String()
	}
	return tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...)
}

func chunkBy(val string, size int, callback func(string, ...trace.EventOption)) {
	if len(val) > maxChunks*size {
		return
	}

	for i := 0; i < maxChunks*size; i += size {
		end := len(val)
		if end > size {
			end = size
		}
		callback(val[0:end])
		if end > len(val)-1 {
			break
		}
		val = val[end:]
	}
}

func dbTable(name string) attribute.KeyValue {
	return dbTableKey.String(name)
}

func dbStatement(stmt string) attribute.KeyValue {
	return dbStatementKey.String(stmt)
}

func dbCount(n int64) attribute.KeyValue {
	return dbRowsAffectedKey.Int64(n)
}

func dbOperation(op string) attribute.KeyValue {
	return dbOperationKey.String(op)
}

func WithOmitVariablesFromTrace(ctx context.Context) context.Context {
	return context.WithValue(ctx, omitVarsKey, true)
}

func operationForQuery(query, op string) string {
	if op != "" {
		return op
	}

	return strings.ToUpper(strings.Split(query, " ")[0])
}

func (op *sqlConfig) spanName(tx *gorm.DB, operation string) string {
	query := extractQuery(tx)

	operation = operationForQuery(query, operation)

	target := "" // op.cfg.dbName
	if target == "" {
		target = tx.Dialector.Name()
	}

	if tx.Statement != nil && tx.Statement.Table != "" {
		target += "." + tx.Statement.Table
	}

	return fmt.Sprintf("%s %s", operation, target)
}
