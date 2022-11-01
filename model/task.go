package model

const (
	ConcurrentModel = 1 // 并发数模式
	LadderModel     = 2 // 阶梯模式
	ErrorRateModel  = 3 // 错误率模式
	RTModel         = 4 // 响应时间模式
	QpsModel        = 5 // 每秒请求数模式

)

const (
	CommonTaskType = 1
	TimingTaskType = 2
	CICDTaskType   = 3
)

// ConfigTask 任务配置
type ConfigTask struct {
	TaskType int64    `json:"task_type" bson:"task_type"` // 任务类型：0. 普通任务； 1. 定时任务； 2. cicd任务
	Mode     int64    `json:"mode" bson:"mode"`           // 压测模式 1:并发模式，2:阶梯模式，3:错误率模式，4:响应时间模式，5:每秒请求数模式
	Remark   string   `json:"remark" bson:"remark"`       // 备注
	CronExpr string   `json:"cron_expr" bson:"cron_expr"` // 定时任务表达式
	ModeConf ModeConf `json:"mode_conf" bson:"mode_conf"` // 任务配置
}

type ModeConf struct {
	ReheatTime       int64 `json:"reheat_time" bson:"reheat_time"`             //预热时长
	RoundNum         int64 `json:"round_num" bson:"round_num"`                 // 轮次
	Concurrency      int64 `json:"concurrency" bson:"concurrency"`             // 并发数
	StartConcurrency int64 `json:"start_concurrency" bson:"start_concurrency"` // 起始并发数
	Step             int64 `json:"step" bson:"step"`                           // 并发步长
	StepRunTime      int64 `json:"step_run_time" bson:"step_run_time"`         // 步长持续时间
	MaxConcurrency   int64 `json:"max_concurrency" bson:"max_concurrency"`     // 最大并发数
	Duration         int64 `json:"duration" bson:"duration"`                   // 稳定持续市场
}
