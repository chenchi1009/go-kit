package task_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chenchi1009/go-kit/task"
)

func TestTask_Run(t *testing.T) {
	// 创建一个测试任务函数
	testTaskFunc := func(ctx context.Context) error {
		return nil
	}

	// 创建一个新的任务实例
	testTask := task.NewTask(testTaskFunc)

	// 运行任务
	err := testTask.Run(context.Background())

	// 检查是否发生了错误
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestTask_Retry(t *testing.T) {
	// 创建一个测试任务函数
	testTaskFunc := func(ctx context.Context) error {
		return errors.New("error")
	}

	// 创建一个带重试的新任务实例
	testTask := task.NewTask(testTaskFunc).Retry(3, func(int) time.Duration { return time.Millisecond }, func(error) bool { return true })

	// 运行任务
	err := testTask.Run(context.Background())

	// 检查是否发生了错误
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestTask_Timeout(t *testing.T) {
	// 创建一个测试任务函数
	testTaskFunc := func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// 创建一个带超时的新任务实例
	testTask := task.NewTask(testTaskFunc).Timeout(50 * time.Millisecond)

	// 运行任务
	err := testTask.Run(context.Background())

	// 检查是否发生了错误
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestTask_WithCompletion(t *testing.T) {
	// 创建一个测试任务函数
	testTaskFunc := func(ctx context.Context) error {
		return nil
	}

	// 标志以检查是否调用了 onSuccess 函数
	onSuccessCalled := false

	// 创建一个带完成通知的新任务实例
	testTask := task.NewTask(testTaskFunc).WithCompletion(func() { onSuccessCalled = true }, nil)

	// 运行任务
	err := testTask.Run(context.Background())

	// 检查是否发生了错误
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// 检查是否调用了 onSuccess 函数
	if !onSuccessCalled {
		t.Error("expected onSuccess function to be called, but it wasn't")
	}
}

func TestTask_RunWithCron(t *testing.T) {
	// 创建一个测试任务函数
	testTaskFunc := func(ctx context.Context) error {
		return nil
	}

	// 创建一个带 cron 的新任务实例
	testTask := task.NewTask(testTaskFunc)

	// 使用 cron 运行任务
	err := testTask.RunWithCron(context.Background(), "* * * * *")

	// 检查是否发生了错误
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// 停止 cron
	testTask.StopCron()
}
