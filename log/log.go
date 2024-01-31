package log

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Fields 是 logrus.Fields 的别名
type Fields logrus.Fields

// Logger 是自定义日志接口
type Logger interface {
	Debug(args ...interface{})                  // 记录调试级别的日志
	Info(args ...interface{})                   // 记录信息级别的日志
	Warn(args ...interface{})                   // 记录警告级别的日志
	Error(args ...interface{})                  // 记录错误级别的日志
	ErrorWithDetails(message string, err error) // 记录带有额外字段的错误日志，自动添加 error 和 trace 字段
	Fatal(args ...interface{})                  // 记录致命错误级别的日志
	FatalWithDetails(message string, err error) // 记录带有额外字段的致命错误日志，自动添加 error 和 trace 字段
	WithFields(fields Fields) Logger            // 创建带有额外字段的新日志实例
	SetDefaultFields(fields Fields)             // 设置默认字段
}

// LogLevel 是自定义日志级别类型
type LogLevel int

const (
	// DebugLevel 表示 debug 日志级别
	DebugLevel LogLevel = iota
	// InfoLevel 表示 info 日志级别
	InfoLevel
	// WarnLevel 表示 warn 日志级别
	WarnLevel
	// ErrorLevel 表示 error 日志级别
	ErrorLevel
	// FatalLevel 表示 fatal 日志级别
	FatalLevel
)

// loggerImpl 是 Logger 接口的实现
type loggerImpl struct {
	logger *logrus.Entry
	path   string
	level  LogLevel
}

var logLevels = map[LogLevel]logrus.Level{
	DebugLevel: logrus.DebugLevel,
	InfoLevel:  logrus.InfoLevel,
	WarnLevel:  logrus.WarnLevel,
	ErrorLevel: logrus.ErrorLevel,
	FatalLevel: logrus.FatalLevel,
}

var logger Logger
var once sync.Once

// Option 是用于设置 Logger 的选项函数类型
type Option func(*loggerImpl)

// DefaultLogger 返回默认的日志实例
func DefaultLogger() Logger {
	once.Do(func() {
		if logger == nil {
			logger = NewLogger(WithPath("logs/default.log"), WithLevel(DebugLevel))
		}
	})
	return logger
}

// NewLogger 创建自定义日志实例
func NewLogger(options ...Option) Logger {
	return newLogger(options...)
}

// newLogger 创建自定义日志实例
func newLogger(options ...Option) Logger {
	// 默认参数
	logger := &loggerImpl{
		path:  "logs/default.log",
		level: InfoLevel,
	}

	// 应用选项函数
	for _, option := range options {
		option(logger)
	}

	// 使用 logrus.NewEntry 创建一个带有默认字段的 *logrus.Entry
	entry := logrus.NewEntry(logrus.New())

	// 设置控制台输出格式为文本
	entry.Logger.SetFormatter(&logrus.TextFormatter{})

	// 配置日志文件输出
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logger.path,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // 天数
	}

	lfHook := lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: lumberjackLogger,
			logrus.InfoLevel:  lumberjackLogger,
			logrus.WarnLevel:  lumberjackLogger,
			logrus.ErrorLevel: lumberjackLogger,
			logrus.FatalLevel: lumberjackLogger,
		},
		&logrus.JSONFormatter{},
	)

	entry.Logger.AddHook(lfHook)

	// 设置日志级别
	if logLevel, ok := logLevels[logger.level]; ok {
		entry.Logger.SetLevel(logLevel)
	} else {
		entry.Warn("提供的日志级别无效。")
	}

	return &loggerImpl{logger: entry}
}

// WithPath 设置日志路径的选项函数
func WithPath(path string) Option {
	return func(l *loggerImpl) {
		l.path = path
	}
}

// WithLevel 设置日志级别的选项函数
func WithLevel(level LogLevel) Option {
	return func(l *loggerImpl) {
		l.level = level
	}
}

// Debug 记录调试级别的日志
func (l *loggerImpl) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Info 记录信息级别的日志
func (l *loggerImpl) Info(args ...interface{}) {
	l.logger.Info(args...)
}

// Warn 记录警告级别的日志
func (l *loggerImpl) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Error 记录错误级别的日志，并添加堆栈信息
func (l *loggerImpl) Error(args ...interface{}) {
	l.logger.WithField("trace", getStackTrace()).Error(args...)
}

// ErrorWithDetails 记录带有额外字段的错误日志，自动添加 error 和 trace 字段
func (l *loggerImpl) ErrorWithDetails(message string, err error) {
	// 添加额外字段
	fields := logrus.Fields{"error": err.Error(), "trace": getStackTrace()}

	// 使用 WithFields 添加额外字段，然后调用 Error 记录错误
	l.logger.WithFields(fields).Error(message)
}

// Fatal 记录致命错误级别的日志，并添加堆栈信息
func (l *loggerImpl) Fatal(args ...interface{}) {
	l.logger.WithField("trace", getStackTrace()).Fatal(args...)
}

// FatalWithDetails 记录带有额外字段的致命错误日志，自动添加 error 和 trace 字段
func (l *loggerImpl) FatalWithDetails(message string, err error) {
	// 添加额外字段
	fields := logrus.Fields{"error": err.Error(), "trace": getStackTrace()}

	// 使用 WithFields 添加额外字段，然后调用 Fatal 记录致命错误
	l.logger.WithFields(fields).Fatal(message)
}

// SetDefaultFields 设置默认字段
func (l *loggerImpl) SetDefaultFields(fields Fields) {
	l.logger = l.logger.WithFields(logrus.Fields(fields))
}

// WithFields 创建带有额外字段的新日志实例
func (l *loggerImpl) WithFields(fields Fields) Logger {
	return &loggerImpl{logger: l.logger.WithFields(logrus.Fields(fields))}
}

// getStackTrace 返回当前堆栈跟踪信息
func getStackTrace() string {
	// 获取当前调用堆栈信息
	pc := make([]uintptr, 10) // 可以根据需要调整堆栈深度
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])

	// 构建堆栈信息字符串
	var traceInfo string
	for {
		frame, more := frames.Next()
		traceInfo += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return traceInfo
}
