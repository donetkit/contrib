package redis

import (
	"context"
	"fmt"
	"github.com/donetkit/contrib-log/glog"
	tracerServer "github.com/donetkit/contrib/tracer"
	"github.com/go-redis/redis/v8"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"strings"
)

type TracingHook struct {
	logger       glog.ILoggerEntry
	tracerServer *tracerServer.Server
	attrs        []attribute.KeyValue
}

func newTracingHook(logger glog.ILoggerEntry, tracerServer *tracerServer.Server, attrs []attribute.KeyValue) *TracingHook {
	hook := &TracingHook{
		logger:       logger,
		tracerServer: tracerServer,
		attrs:        attrs,
	}
	return hook
}

func (h *TracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	dbValue := ""
	db, ok := ctx.Value(redisClientDBKey).(int)
	if ok {
		dbValue = fmt.Sprintf("[%d]", db)
	}
	cmdName := getTraceFullName(cmd, dbValue)
	//if h.logger != nil {
	//	h.logger.Info(cmdName)
	//}
	if h.tracerServer == nil {
		return ctx, nil
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx, nil
	}
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(h.attrs...),
		trace.WithAttributes(
			semconv.DBStatementKey.String(cmdName),
		),
	}
	ctx, _ = h.tracerServer.Tracer.Start(ctx, cmd.FullName(), opts...)
	return ctx, nil
}

func (h *TracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	dbValue := ""
	db, ok := ctx.Value(redisClientDBKey).(int)
	if ok {
		dbValue = fmt.Sprintf("[%d]", db)
	}
	if h.logger != nil {
		h.logger.Debugf("db%s:redis:%s ", dbValue, cmd.String())
	}
	if h.tracerServer == nil {
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}
	defer span.End()
	span.SetName(getTraceFullName(cmd, dbValue))
	if err := cmd.Err(); err != nil {
		h.recordError(ctx, span, err)
	}
	return nil
}

func (h *TracingHook) BeforeProcessPipeline(ctx context.Context, cmd []redis.Cmder) (context.Context, error) {
	dbValue := ""
	db, ok := ctx.Value(redisClientDBKey).(int)
	if ok {
		dbValue = fmt.Sprintf("[%d]", db)
	}
	cmdName := getTraceFullNames(cmd, dbValue)
	//if h.logger != nil {
	//	h.logger.Info(cmdName)
	//}
	if h.tracerServer == nil {
		return ctx, nil
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return ctx, nil
	}
	summary, _ := CmdsString(cmd)
	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(h.attrs...),
		trace.WithAttributes(
			semconv.DBStatementKey.String(cmdName),
			attribute.Int("db.redis.num_cmd", len(cmd)),
		),
	}

	ctx, _ = h.tracerServer.Tracer.Start(ctx, "pipeline "+summary, opts...)

	return ctx, nil
}

func (h *TracingHook) AfterProcessPipeline(ctx context.Context, cmder []redis.Cmder) error {
	dbValue := ""
	db, ok := ctx.Value(redisClientDBKey).(int)
	if ok {
		dbValue = fmt.Sprintf("[%d]", db)
	}
	if h.logger != nil {
		for _, c := range cmder {
			h.logger.Debugf("db%s:redis:%s ", dbValue, c.String())
		}
	}
	if h.tracerServer == nil {
		return nil
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}
	defer span.End()
	span.SetName(getTraceFullNames(cmder, dbValue))
	if err := cmder[0].Err(); err != nil {
		h.recordError(ctx, span, err)
	}
	return nil
}

func (h *TracingHook) recordError(ctx context.Context, span trace.Span, err error) {
	if err != redis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if h.logger != nil {
			h.logger.Error(err.Error())
		}
	}
}

func getTraceFullName(cmd redis.Cmder, dbValue string) string {
	var args = cmd.Args()
	switch name := cmd.Name(); name {
	case "cluster", "command":
		if len(args) == 1 {
			return fmt.Sprintf("db%s:redis:%s", dbValue, name)
		}
		if s2, ok := args[1].(string); ok {
			return fmt.Sprintf("db%s:redis:%s => %s", dbValue, name, s2)
		}
		return fmt.Sprintf("db%s:redis:%s", dbValue, name)
	default:
		if len(args) == 1 {
			return fmt.Sprintf("db%s:redis:%s", dbValue, name)
		}
		if s2, ok := args[1].(string); ok {
			return fmt.Sprintf("db%s:redis:%s => %s", dbValue, name, s2)
		}
		return fmt.Sprintf("db%s:redis:%s", dbValue, name)
	}
}

func getTraceFullNames(cmd []redis.Cmder, dbValue string) string {
	var cmdStr []string
	for _, c := range cmd {
		cmdStr = append(cmdStr, getTraceFullName(c, dbValue))
	}
	return strings.Join(cmdStr, ", ")
}
