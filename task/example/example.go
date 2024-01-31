package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chenchi1009/go-kit/task"
)

func main() {
	// 定义一个简单的任务函数
	myTask := func(ctx context.Context) error {
		fmt.Println("Executing the task...")
		// 模拟任务执行时间
		time.Sleep(2 * time.Second)
		return nil
	}

	// 创建任务链
	taskChain := task.NewTask(myTask).
		Timeout(5*time.Second).
		Retry(3, func(attempt int) time.Duration {
			return time.Duration(attempt) * time.Second
		}, func(err error) bool {
			// 重试条件：只有在出现特定错误时才重试
			return err != nil && err.Error() == "specific error"
		}).
		WithCompletion(func() {
			fmt.Println("Task completed successfully!")
		}, func(err error) {
			fmt.Printf("Task failed with error: %v\n", err)
		})

	// 执行任务链
	err := taskChain.Run(context.Background())
	if err != nil {
		fmt.Printf("Task chain execution failed: %v\n", err)
	}

	// 使用 cron 表达式执行任务链
	err = taskChain.RunWithCron(context.Background(), "*/5 * * * * *	") // 每隔5秒钟执行一次
	if err != nil {
		fmt.Printf("Cron task execution failed: %v\n", err)
	}

	// 模拟任务链执行中止
	time.Sleep(10 * time.Second)

	// 停止 cron 任务
	taskChain.StopCron()
}
