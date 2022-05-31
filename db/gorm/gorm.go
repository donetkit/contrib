package gorm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/donetkit/contrib/db/db_sql"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"io"
	"strings"
)

const (
	callBackBeforeName = "gorm:before"
	callBackAfterName  = "gorm:after"
	opCreate           = "INSERT"
	opQuery            = "SELECT"
	opDelete           = "DELETE"
	opUpdate           = "UPDATE"
)

type internalCtxKey string

const (
	dbTableKey        = attribute.Key("db.sql.table")
	dbRowsAffectedKey = attribute.Key("db.rows_affected")
	dbOperationKey    = semconv.DBOperationKey
	dbStatementKey    = semconv.DBStatementKey
	omitVarsKey       = internalCtxKey("omit_vars")
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
	cb := db.Callback()
	hooks := []struct {
		callback gormRegister
		hook     gormHookFunc
		name     string
	}{
		// before hooks
		{cb.Create().Before("gorm:before_create"), p.before(opCreate), beforeName("create")},
		{cb.Query().Before("gorm:before_query"), p.before(opQuery), beforeName("query")},
		{cb.Delete().Before("gorm:before_delete"), p.before(opDelete), beforeName("delete")},
		{cb.Update().Before("gorm:before_update"), p.before(opUpdate), beforeName("update")},
		{cb.Row().Before("gorm:before_row"), p.before(opQuery), beforeName("row")},
		{cb.Raw().Before("gorm:before_raw"), p.before(opQuery), beforeName("raw")},

		// after hooks
		{cb.Create().After("gorm:after_create"), p.after(opCreate), afterName("create")},
		{cb.Query().After("gorm:after_query"), p.after(opQuery), afterName("select")},
		{cb.Delete().After("gorm:after_delete"), p.after(opDelete), afterName("delete")},
		{cb.Update().After("gorm:after_update"), p.after(opUpdate), afterName("update")},
		{cb.Row().After("gorm:after_row"), p.after(opQuery), afterName("row")},
		{cb.Raw().After("gorm:after_raw"), p.after(opQuery), afterName("raw")},
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
		if p.tracerServer == nil {
			return
		}
		opts := []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindClient),
		}
		spanName := p.spanName(tx, operation)
		spanName = fmt.Sprintf("db:gorm:%s", spanName)
		p.logger.Info(spanName)
		tx.Statement.Context, _ = p.tracerServer.Tracer.Start(tx.Statement.Context, spanName, opts...)
	}
}

func (p *sqlConfig) after(operation string) gormHookFunc {
	return func(tx *gorm.DB) {
		if p.tracerServer == nil {
			return
		}
		span := trace.SpanFromContext(tx.Statement.Context)
		if !span.IsRecording() {
			return
		}
		defer span.End()
		attrs := make([]attribute.KeyValue, 0, len(p.attrs)+4)
		attrs = append(attrs, p.attrs...)

		if sys := dbSystem(tx); sys.Valid() {
			attrs = append(attrs, sys)
		}
		vars := tx.Statement.Vars
		if p.excludeQueryVars {
			// Replace query variables with '?' to mask them
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
		//p.logger.Info(tx.Statement.RowsAffected)
		switch tx.Error {
		case nil,
			gorm.ErrRecordNotFound,
			driver.ErrSkip,
			io.EOF, // end of rows iterator
			sql.ErrNoRows:
			// ignore
		default:
			span.RecordError(tx.Error)
			span.SetStatus(codes.Error, tx.Error.Error())
		}
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

func (p *sqlConfig) spanName(tx *gorm.DB, operation string) string {
	query := extractQuery(tx)
	operation = operationForQuery(query, operation)
	table := ""
	if tx.Statement != nil && tx.Statement.Table != "" {
		table = ":" + tx.Statement.Table
	}
	operation = strings.ToLower(operation)
	return fmt.Sprintf("%s%s", operation, table)
}

func extractQuery(tx *gorm.DB) string {
	if shouldOmit, _ := tx.Statement.Context.Value(omitVarsKey).(bool); shouldOmit {
		return tx.Statement.SQL.String()
	}
	return tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...)
}

func operationForQuery(query, op string) string {
	if op != "" {
		return op
	}
	return strings.ToUpper(strings.Split(query, " ")[0])
}
