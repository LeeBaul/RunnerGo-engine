package task

// 任务，定义接收的参数

type Task struct {
	TaskId      int64
	TaskName    string
	Concurrency int64
	Scenarios   []*Scenario
}
