# Task Package

`task` 包提供了创建、管理和执行任务的方式。它包括重试任务、为任务设置超时以及添加完成回调等功能。

## 使用

```go
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
        // 设置超时时间，如果 Timeout 在 Retry 前被调用指单个任务限时，
        // 如果在 Retry 之后代表总限时
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
    err = taskChain.RunWithCron(context.Background(), "*/5 * * * * *    ") // 每隔5秒钟执行一次
    if err != nil {
        fmt.Printf("Cron task execution failed: %v\n", err)
    }

    // 模拟任务链执行中止
    time.Sleep(10 * time.Second)

    // 停止 cron 任务
    taskChain.StopCron()
}
```

## cron 表达式

使用 [github.com/robfig/cron/v3](https://github.com/robfig/cron/tree/v3) 实现 `cron` 表达式功能， `cron` 语法与 `Linux` 下的略有不同，有六位参数，具体文档见 [go doc](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format)

### 翻译

一个 cron 表达式代表了一组时间，使用 6 个由空格分隔的字段。

| 字段名称     | 是否必须? | 允许的值        | 允许的特殊字符 |
| ------------ | --------- | --------------- | -------------- |
| 秒           | 是        | 0-59            | \* / , -       |
| 分钟         | 是        | 0-59            | \* / , -       |
| 小时         | 是        | 0-23            | \* / , -       |
| 月份中的日期 | 是        | 1-31            | \* / , - ?     |
| 月份         | 是        | 1-12 或 JAN-DEC | \* / , -       |
| 星期中的日期 | 是        | 0-6 或 SUN-SAT  | \* / , - ?     |

_注意:_ 月份和星期中的日期字段值不区分大小写。"SUN", "Sun", 和 "sun" 都被接受。

#### 特殊字符

- 星号 ( \* )

  星号表示 cron 表达式将匹配字段的所有值；例如，在第 5 个字段（月份）使用星号将表示每个月。

- 斜线 ( / )

  斜线用于描述范围的增量。例如，在第 1 个字段（分钟）中使用 3-59/15 将表示每小时的第 3 分钟和之后的每 15 分钟。形式 "\*/..." 等同于形式 "first-last/..."，即，字段可能的最大范围的增量。形式 "N/..." 被接受为 "N-MAX/..."，即，从 N 开始，使用增量直到该特定范围的结束。它不会回绕。

- 逗号 ( , )

  逗号用于分隔列表的项。例如，在第 5 个字段（星期中的日期）中使用 "MON,WED,FRI" 将表示星期一、星期三和星期五。

- 连字符 ( - )

  连字符用于定义范围。例如，9-17 将表示上午 9 点到下午 5 点之间的每个小时，包括两者。

- 问号 ( ? )

  问号可以用 '\*' 的位置代替，以留空月份中的日期或星期中的日期。

#### 预定义的时间表

你可以使用几个预定义的时间表代替 cron 表达式。

| 条目                   | 描述                            | 等价于          |
| ---------------------- | ------------------------------- | --------------- |
| @yearly (或 @annually) | 每年运行一次，1 月 1 日午夜     | 0 0 0 1 1 \*    |
| @monthly               | 每月运行一次，每月第一天午夜    | 0 0 0 1 \* \*   |
| @weekly                | 每周运行一次，星期六/星期日午夜 | 0 0 0 \* \* 0   |
| @daily (或 @midnight)  | 每天运行一次，午夜              | 0 0 0 \* \* \*  |
| @hourly                | 每小时运行一次，每小时开始时    | 0 0 \* \* \* \* |

#### 间隔

你也可以安排一个作业以固定的间隔执行，从添加它或运行 cron 的时间开始。这是通过像这样格式化 cron 规范来支持的：

@every 其中 "duration" 是一个被 time.ParseDuration (http://golang.org/pkg/time/#ParseDuration) 接受的字符串。

例如，"@every 1h30m10s" 将表示一个在 1 小时、30 分钟、10 秒后激活的时间表，然后是之后的每个间隔。

注意: 间隔不考虑作业运行时间。例如，如果一个作业需要 3 分钟才能运行，而它被安排每 5 分钟运行一次，那么它在每次运行之间只有 2 分钟的空闲时间。
