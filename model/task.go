package model

const (
	ConcurrentModel = iota // 并发数模式
	ErrorRateModel         // 错误率模式
	LadderModel            // 阶梯模式
	TpsModel               // 每秒事务数模式
	QpsModel               // 每秒请求数模式
	RTModel                // 响应时间模式
)

const (
	DurationType = int64(iota) // 按时长执行
	RoundsType                 // 按轮次执行
)

const (
	CommonTaskType = iota
	TimingTaskType
	CICDTaskType
)

// ConfigTask 任务配置
type ConfigTask struct {
	TaskType  int64     `json:"taskType"` // 任务类型：0. 普通任务； 1. 定时任务； 2. cicd任务
	Task      Task      `json:"task"`
	TestModel TestModel `json:"testModel"`
}

// Task 任务
type Task struct {
	TimingTask TimingTask `json:"timingTask"`
	CICDTask   CICDTask   `json:"CICDTask"`
}

// TimingTask 定时任务
type TimingTask struct {
	Spec string `json:"spec"`
}

// CICDTask cicd任务
type CICDTask struct {
}

// TestModel 压测模型
type TestModel struct {
	Type           int8           `json:"type"` // 0:ConcurrentModel; 1:ErrorRateModel; 2:LadderModel; 3:TpsModel; 4:QpsModel; 5:RTModel
	ConcurrentTest ConcurrentTest `json:"concurrentTest"`
	ErrorRateTest  ErrorRateTest  `json:"errorRatTest"`
	LadderTest     LadderTest     `json:"ladderTest"`
	TpsTest        TpsTest        `json:"tpsTest"`
	QpsTest        QpsTest        `json:"qpsTest"`
	RTTest         RTTest         `json:"rtTest"`
}

// ConcurrentTest 并发模式 0
type ConcurrentTest struct {
	Type       int64 `json:"type"`       // 0:DurationType; 1:RoundsType
	Concurrent int64 `json:"concurrent"` // 并发数
	Duration   int64 `json:"duration"`   // 持续时长
	Rounds     int64 `json:"rounds"`     // 轮次
	TimeUp     int64 `json:"timeUp"`     // 启动并发数时长
}

// ErrorRateTest 错误率模式 1
type ErrorRateTest struct {
	StartConcurrent int64 `json:"startConcurrent"` // 起始并发数
	Length          int64 `json:"length"`          // 步长
	LengthDuration  int64 `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64 `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64 `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64 `json:"timeUp"`          // 启动并发数时长
}

// LadderTest 阶梯模式 2
type LadderTest struct {
	StartConcurrent int64 `json:"startConcurrent"` // 起始并发数
	Length          int64 `json:"length"`          // 步长
	LengthDuration  int64 `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64 `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64 `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64 `json:"timeUp"`          // 启动并发数时长
}

//	TpsTest 每秒事务数模式 3
type TpsTest struct {
	Threshold       float32 `json:"threshold"`       // 阈值
	StartConcurrent int64   `json:"startConcurrent"` // 起始并发数
	Length          int64   `json:"length"`          // 步长
	LengthDuration  int64   `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64   `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64   `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64   `json:"timeUp"`          // 启动并发数时长
}

//	QpsTest 每秒请求数模式 4
type QpsTest struct {
	StartConcurrent int64 `json:"startConcurrent"` // 起始并发数
	Length          int64 `json:"length"`          // 步长
	LengthDuration  int64 `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64 `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64 `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64 `json:"timeUp"`          // 启动并发数时长
}

//	RTTest 响应时间模式 5
type RTTest struct {
	StartConcurrent int64 `json:"startConcurrent"` // 起始并发数
	Length          int64 `json:"length"`          // 步长
	LengthDuration  int64 `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64 `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64 `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64 `json:"timeUp"`          // 启动并发数时长
}
