package ctxlog

import (
	"context"
	"sync"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

type ctxManagerKey struct{}

// From returns logger from context.
func From(ctx context.Context) *logrus.Entry { return ctxlogrus.Extract(ctx) }

// With adds logger to context.
func With(ctx context.Context, entry *logrus.Entry) context.Context {
	return ctxlogrus.ToContext(ctx, entry)
}

// WithFields adds fields to context logger.
func WithFields(ctx context.Context, fields logrus.Fields) context.Context {
	return With(ctx, From(ctx).WithFields(fields))
}

// WithField adds field to context logger.
func WithField(ctx context.Context, key string, value interface{}) context.Context {
	return With(ctx, From(ctx).WithField(key, value))
}

// Test returns context with logrus testing logger.
func Test(ctx context.Context) (context.Context, *test.Hook) {
	log, hook := test.NewNullLogger()
	e := logrus.NewEntry(log)
	return With(ctx, e), hook
}

type LogFieldManager struct {
	fields logrus.Fields
	mut    sync.RWMutex
}

func (l *LogFieldManager) AddFields(fields map[string]interface{}) {
	l.mut.Lock()
	defer l.mut.Unlock()

	for k, v := range fields {
		l.fields[k] = v
	}
}

// WithFieldManagerFields adds fields from FieldManager to logger.
func WithFieldManagerFields(ctx context.Context, log *logrus.Entry) *logrus.Entry {
	manager := FieldManagerFrom(ctx)
	manager.mut.RLock()
	newLog := log.WithFields(manager.fields)
	manager.mut.RUnlock()
	return newLog
}

func newLogFieldManager() *LogFieldManager {
	return &LogFieldManager{
		fields: make(map[string]interface{}),
	}
}

func NewContextWithFieldManager(ctx context.Context) context.Context {
	manager := newLogFieldManager()
	return context.WithValue(ctx, ctxManagerKey{}, manager)
}

func FieldManagerFrom(ctx context.Context) *LogFieldManager {
	manager, ok := ctx.Value(ctxManagerKey{}).(*LogFieldManager)
	if ok {
		return manager
	}
	return newLogFieldManager()
}

