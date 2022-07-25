package task

import (
	"kp-runner/model"
	"kp-runner/tools"
)

const (
	ConcurrencyModel  = iota // 并发模式
	LadderModel              // 阶梯模式
	ErrorRateModel           // 错误率模式
	ResponseTimeModel        // 响应时间模式
	QPSModel                 // 每秒请求数模式
	TPSModel                 // 每秒事务数模式

	GeneralTask = iota // 普通任务
	TimingTask         // 定时任务
	CICDTask           // 持续集成任务
)

//为场景中每个请求的节点，使用链表组成树形结构

type RequestNode struct {
	Request    *model.Request   // 请求
	Controller *tools.Condition // 控制器

	NextRequests  []*RequestNode // 后面的请求
	Assertions    []string       // 本请求的断言
	Parameterizes []string       // 参数话数据
}

// 场景

type Scenario struct {
	ScenarioId        int64          // 场景Id
	ScenarioName      string         // 场景名称
	FirstRequestNodes []*RequestNode // 场景的第一个请求，并列关系，可以是多个
	Path              string         // 参数化文件地址
	Model             int64          // 压测模式
	TaskType          int64          // 任务类型
}
