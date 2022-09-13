package model

// Plan 计划结构体
type Plan struct {
	PlanId     int64       `json:"plan_id" bson:"plan_id"`     // 计划id
	PlanName   string      `json:"plan_name" bson:"plan_name"` // 计划名称
	ReportId   string      `json:"report_id" bson:"report_id"` // 报告名称
	ReportName string      `json:"report_name" bson:"report_name"`
	ConfigTask *ConfigTask `json:"config_task" bson:"config_task"` // 任务配置
	Variable   []*KV       `json:"variable" bson:"variable"`       // 全局变量
	Scene      *Scene      `json:"scene" bson:"scene"`             // 场景
}
