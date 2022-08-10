package task

const (
	ConcurrentModel = iota // 并发数模式
	ErrorRateModel         // 错误率模式
	LadderModel            // 阶梯模式
	TpsModel               // 每秒事务数模式
	QpsModel               // 每秒请求数模式
	RTModel                // 响应时间模式
)

const (
	DurationType = iota // 按时长执行
	RoundsType          // 按轮次执行
)

const (
	CommonTask = iota
	TimingTask
	CICDTask
)

// ConfigTask 任务配置
type ConfigTask struct {
	TaskType  int8      `json:"taskType"` // 任务类型：0. 普通任务； 1. 定时任务； 2. cicd任务
	TestModel TestModel `json:"testModel"`
}

// TestModel 压测模型
type TestModel struct {
	Type           int8           `json:"type"` // 0:ConcurrentModel; 1:ErrorRateModel; 2:LadderModel; 3:TpsModel; 4:QpsModel; 5:RTModel
	ConcurrentTest ConcurrentTest `json:"concurrentTest"`
	ErrorRatTest   ErrorRatTest   `json:"errorRatTest"`
	LadderTest     LadderTest     `json:"ladderTest"`
	TpsTest        TpsTest        `json:"tpsTest"`
	QpsTest        QpsTest        `json:"qpsTest"`
	RTTest         RTTest         `json:"rtTest"`
}

// ConcurrentTest 并发模式 0
type ConcurrentTest struct {
	Type       int8  `json:"type"`       // 0:DurationType; 1:RoundsType
	Concurrent int64 `json:"concurrent"` // 并发数
	Duration   int64 `json:"duration"`   // 持续时长
	Rounds     int64 `json:"rounds"`     // 轮次
	TimeUp     int64 `json:"timeUp"`     // 启动并发数时长
}

// ErrorRatTest 错误率模式 1
type ErrorRatTest struct {
	Threshold       float64 `json:"threshold"`       // 阈值
	StartConcurrent int64   `json:"startConcurrent"` // 起始并发数
	Length          int64   `json:"length"`          // 步长
	LengthDuration  int64   `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64   `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64   `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64   `json:"timeUp"`          // 启动并发数时长
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
	Threshold       float64 `json:"threshold"`       // 阈值
	StartConcurrent int64   `json:"startConcurrent"` // 起始并发数
	Length          int64   `json:"length"`          // 步长
	LengthDuration  int64   `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64   `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64   `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64   `json:"timeUp"`          // 启动并发数时长
}

//	QpsTest 每秒请求数模式 4
type QpsTest struct {
	Threshold       float64 `json:"threshold"`       // 阈值
	StartConcurrent int64   `json:"startConcurrent"` // 起始并发数
	Length          int64   `json:"length"`          // 步长
	LengthDuration  int64   `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64   `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64   `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64   `json:"timeUp"`          // 启动并发数时长
}

//	RTTest 响应时间模式 5
type RTTest struct {
	Standard        int   `json:"standard"`        // 0:平均响应时间；1. 90%rt; 2. 95%rt; 3. 99%rt; 4. 自定义
	Threshold       int   `json:"threshold"`       // 阈值
	StartConcurrent int64 `json:"startConcurrent"` // 起始并发数
	Length          int64 `json:"length"`          // 步长
	LengthDuration  int64 `json:"lengthDuration"`  // 步长持续时间
	MaxConcurrent   int64 `json:"MaxConcurrent"`   // 最大并发数
	StableDuration  int64 `json:"stableDuration"`  // 稳定持续时长
	TimeUp          int64 `json:"timeUp"`          // 启动并发数时长
}
