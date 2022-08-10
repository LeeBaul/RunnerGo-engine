package plan

import (
	"kp-runner/model/scene"
	"kp-runner/model/task"
	"kp-runner/model/variable"
)

// Plan 计划结构体
type Plan struct {
	PlanID     int               `json:"planId"`   // 计划id
	PlanName   string            `json:"planName"` // 计划名称
	ReportId   int               `json:"reportId"` // 报告名称
	ReportName string            `json:"reportName"`
	ConfigTask task.ConfigTask   `json:"configTask"` // 任务配置
	Variable   variable.Variable `json:"variable"`   // 全局变量
	Scene      scene.Scene       `json:"scene"`      // 场景列表
}
