package runtime

import (
	"context"
	"errors"
	"strings"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// filteredGormLogger suppresses expected request-abort noise while preserving
// normal GORM warning/error behavior.
type filteredGormLogger struct {
	gormlogger.Interface
}

func newFilteredGormLogger(base gormlogger.Interface) gormlogger.Interface {
	return filteredGormLogger{Interface: base}
}

func (l filteredGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return filteredGormLogger{Interface: l.Interface.LogMode(level)}
}

func (l filteredGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if shouldIgnoreGormTraceError(err) {
		return
	}
	l.Interface.Trace(ctx, begin, fc, err)
}

func shouldIgnoreGormTraceError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "context canceled") ||
		strings.Contains(errMsg, "context already done: context canceled")
}
