package model

import "sync"

// Plan 计划结构体
type Plan struct {
	PlanId     string      `json:"plan_id" bson:"plan_id"`     // 计划id
	PlanName   string      `json:"plan_name" bson:"plan_name"` // 计划名称
	ReportId   string      `json:"report_id" bson:"report_id"` // 报告名称
	ReportName string      `json:"report_name" bson:"report_name"`
	ConfigTask *ConfigTask `json:"config_task" bson:"config_task"` // 任务配置
	Variable   *sync.Map   `json:"variable" bson:"variable"`       // 全局变量
	Scene      *Scene      `json:"scene" bson:"scene"`             // 场景
}

// Group 分组
type Group struct {
	Group    *Group    `json:"groups" bson:"groups"`
	Variable *sync.Map `json:"variable"` // 全局变量
	Scene    *Scene    `json:"scene"`    // 场景
}
