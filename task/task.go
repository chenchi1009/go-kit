package task

import (
	"context"
	"errors"
	"time"

	"github.com/robfig/cron/v3"
)

// TaskFunc 是一个接受 context.Context 参数并返回 error 的函数类型
type TaskFunc func(context.Context) error

// Task 是一个任务类型，包含了要执行的任务函数以及该任务的配置选项
type Task struct {
	taskFunc TaskFunc
	cron     *cron.Cron
}

// NewTask 创建一个新的任务实例
func NewTask(task TaskFunc) *Task {
	return &Task{taskFunc: task}
}

// Retry 添加重试逻辑到当前任务链中，并返回新的任务链
func (t *Task) Retry(retryCount int, backoff func(int) time.Duration, condition func(error) bool) *Task {
	return &Task{
		taskFunc: retry(t.taskFunc, retryCount, backoff, condition),
	}
}

// Timeout 添加超时逻辑到当前任务链中，并返回新的任务链
func (t *Task) Timeout(dur time.Duration) *Task {
	return &Task{
		taskFunc: timeout(t.taskFunc, dur),
	}
}

// WithCompletionNotification 添加完成通知功能到当前任务链中，并返回新的任务链
func (t *Task) WithCompletion(onSuccess func(), onFailure func(error)) *Task {
	return &Task{
		taskFunc: withCompletion(t.taskFunc, onSuccess, onFailure),
	}
}

// Run 执行任务链
func (t *Task) Run(ctx context.Context) error {
	return t.taskFunc(ctx)
}

// RunWithCron 使用 cron 表达式执行任务链
func (t *Task) RunWithCron(ctx context.Context, cronExpr string) error {
	t.cron = cron.New(cron.WithSeconds())
	_, err := t.cron.AddFunc(cronExpr, func() { _ = t.taskFunc(ctx) })
	if err != nil {
		return err
	}
	t.cron.Start()

	go func() {
		<-ctx.Done()
		t.cron.Stop()
	}()

	return nil
}

// StopCron 停止 cron 任务
func (t *Task) StopCron() {
	if t.cron != nil {
		t.cron.Stop()
	}
}

// retry 为一个任务添加重试逻辑，并根据条件函数决定是否进行重试
func retry(task TaskFunc, retryCount int, backoff func(int) time.Duration, condition func(error) bool) TaskFunc {
	return func(ctx context.Context) error {
		var err error
		for i := 0; i < retryCount; i++ {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				return ctx.Err() // 返回上下文取消的错误
			default:
				// 如果上下文未取消，则继续执行任务
			}

			err = task(ctx)
			if err == nil || !condition(err) {
				// 如果任务执行成功或者不满足重试条件，则立即返回
				break
			}
			time.Sleep(backoff(i))
		}
		return err
	}
}

// timeout 为一个任务添加超时逻辑
func timeout(task TaskFunc, dur time.Duration) TaskFunc {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, dur)
		defer cancel()

		done := make(chan error)
		go func() {
			done <- task(ctx)
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return errors.New("task timeout")
		}
	}
}

// withCompletionNotification 为一个任务添加完成通知功能
func withCompletion(task TaskFunc, onSuccess func(), onFailure func(error)) TaskFunc {
	return func(ctx context.Context) error {
		err := task(ctx)
		if err != nil && onFailure != nil {
			// 如果任务执行失败且存在失败通知函数，则调用失败通知函数
			onFailure(err)
		} else if err == nil && onSuccess != nil {
			// 如果任务执行成功且存在成功通知函数，则调用成功通知函数
			onSuccess()
		}
		return err
	}
}
