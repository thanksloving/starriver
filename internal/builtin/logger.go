package builtin

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/thanksloving/starriver"
)

func NewLogger() starriver.Logger {
	std := log.New()
	std.SetLevel(starriver.DebugLevel)
	std.SetFormatter(&log.TextFormatter{
		DisableSorting:   false,
		QuoteEmptyFields: true,
		FullTimestamp:    true,
	})
	return &defaultLogger{
		std,
	}
}

type defaultLogger struct {
	*log.Logger
}

func (dl *defaultLogger) SetLoggerLevel(logLevel starriver.LogLevel) {
	dl.SetLevel(logLevel)
}

func (dl *defaultLogger) getLogger(ctx starriver.DataContext) *log.Entry {
	return dl.WithFields(map[string]interface{}{
		"request_id": ctx.GetRequestID(),
	}).WithContext(ctx).WithTime(time.Now())
}

func (dl *defaultLogger) Debug(ctx starriver.DataContext, msg string) {
	dl.getLogger(ctx).Debug(msg)
}

func (dl *defaultLogger) Debugf(ctx starriver.DataContext, msg string, args ...interface{}) {
	dl.getLogger(ctx).Debugf(msg, args...)
}

func (dl *defaultLogger) Info(ctx starriver.DataContext, msg string) {
	dl.getLogger(ctx).Info(msg)
}

func (dl *defaultLogger) Infof(ctx starriver.DataContext, msg string, args ...interface{}) {
	dl.getLogger(ctx).Infof(msg, args...)
}

func (dl *defaultLogger) Warn(ctx starriver.DataContext, msg string) {
	dl.getLogger(ctx).Warn(msg)
}

func (dl *defaultLogger) Warnf(ctx starriver.DataContext, msg string, args ...interface{}) {
	dl.getLogger(ctx).Warnf(msg, args...)
}

func (dl *defaultLogger) Error(ctx starriver.DataContext, msg string) {
	dl.getLogger(ctx).Error(msg)
}

func (dl *defaultLogger) Errorf(ctx starriver.DataContext, msg string, args ...interface{}) {
	dl.getLogger(ctx).Errorf(msg, args...)
}

func (dl *defaultLogger) Fatal(ctx starriver.DataContext, msg string) {
	dl.getLogger(ctx).Fatal(msg)
}

func (dl *defaultLogger) Fatalf(ctx starriver.DataContext, msg string, args ...interface{}) {
	dl.getLogger(ctx).Fatalf(msg, args...)
}
