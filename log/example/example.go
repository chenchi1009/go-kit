package main

import (
	"github.com/chenchi1009/go-kit/log"
	"github.com/pkg/errors"
)

func main() {
	// 创建日志实例
	logger := log.NewLogger(log.WithPath("logs/example.log"), log.WithLevel(log.DebugLevel))

	// 设置默认字段
	logger.SetDefaultFields(log.Fields{
		"app":    "exampleApp",
		"env":    "development",
		"server": "localhost",
	})

	// 模拟应用启动
	logger.Info("Application started")

	// 模拟应用运行中的日志记录
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")

	// 模拟错误和致命错误
	err := simulateError()
	if err != nil {
		logger.ErrorWithDetails("Error occurred", err)
		// logger.FatalWithDetails("Fatal error occurred", err)
	}

	// 模拟应用关闭
	logger.Info("Application shutting down")
}

func simulateError() error {
	// 模拟一个错误
	return errors.New("This is a custom error")
}
