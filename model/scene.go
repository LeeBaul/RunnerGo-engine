package model

import (
	uuid "github.com/satori/go.uuid"
	"sync"
)

// Scene 场景结构体
type Scene struct {
	SceneId                 int64          `json:"scene_id" bson:"scene_id"` // 场景Id
	Uuid                    uuid.UUID      `json:"uuid" bson:"uuid"`
	TeamId                  int64          `json:"team_id" bson:"team_id"`
	SceneName               string         `json:"scene_name" bson:"scene_name"` // 场景名称
	Version                 int64          `json:"version" bson:"version"`
	EnablePlanConfiguration bool           `json:"enable_plan_configuration" bson:"enable_plan_configuration"` // 是否启用计划的任务配置，默认为true，
	Nodes                   []Event        `json:"nodes" bson:"nodes"`                                         // 事件列表
	ConfigTask              *ConfigTask    `json:"config_task" bson:"config_task"`                             // 任务配置
	Configuration           *Configuration `json:"configuration" bson:"configuration"`                         // 场景配置

}

type Configuration struct {
	ParameterizedFile *ParameterizedFile `json:"parameterizedFile" bson:"parameterizedFile"`
	Variable          *sync.Map          `json:"variable" bson:"variable"`
	Mu                sync.Mutex         `json:"mu" bson:"mu"`
}
