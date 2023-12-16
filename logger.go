package starriver

import log "github.com/sirupsen/logrus"

var (
	PanicLevel = log.PanicLevel
	FatalLevel = log.FatalLevel
	ErrorLevel = log.ErrorLevel
	WarnLevel  = log.WarnLevel
	InfoLevel  = log.InfoLevel
	DebugLevel = log.DebugLevel
	TraceLevel = log.TraceLevel
)

type (
	LogLevel = log.Level

	Logger interface {
		SetLoggerLevel(logLevel LogLevel)
		Debug(ctx DataContext, msg string)
		Debugf(ctx DataContext, msg string, args ...interface{})
		Info(ctx DataContext, msg string)
		Infof(ctx DataContext, msg string, args ...interface{})
		Warn(ctx DataContext, msg string)
		Warnf(ctx DataContext, msg string, args ...interface{})
		Error(ctx DataContext, msg string)
		Errorf(ctx DataContext, msg string, args ...interface{})
		Fatal(ctx DataContext, msg string)
		Fatalf(ctx DataContext, msg string, args ...interface{})
	}

	dataContextLogger interface {
		Debug(msg string)
		Debugf(msg string, args ...interface{})
		Info(msg string)
		Infof(msg string, args ...interface{})
		Warn(msg string)
		Warnf(msg string, args ...interface{})
		Error(msg string)
		Errorf(msg string, args ...interface{})
		Fatal(msg string)
		Fatalf(msg string, args ...interface{})
	}
)
