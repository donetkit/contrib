package gorm

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

const (
	callBackBeforeName = "otel:before"
	callBackAfterName  = "otel:after"
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

const (
	eventMaxSize = 250
	maxChunks    = 4
)
