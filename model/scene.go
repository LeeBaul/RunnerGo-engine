package model

import "sync"

// Scene 场景结构体
type Scene struct {
	SceneId                 string         `json:"sceneId" bson:"sceneId"`                                 // 场景Id
	SceneName               string         `json:"sceneName" bson:"sceneName"`                             // 场景名称
	EnablePlanConfiguration bool           `json:"enablePlanConfiguration" bson:"enablePlanConfiguration"` // 是否启用计划的任务配置，默认为true，
	EventList               []Event        `json:"eventList" bson:"eventList"`                             // 事件列表
	ConfigTask              *ConfigTask    `json:"configTask" bson:"configTask"`                           // 任务配置
	Configuration           *Configuration `json:"configuration" bson:"configuration"`                     // 场景配置

}

type Configuration struct {
	ParameterizedFile *ParameterizedFile `json:"parameterizedFile" bson:"parameterizedFile"`
	Variable          *sync.Map          `json:"variable" bson:"variable"`
	Mu                sync.Mutex         `json:"mu" bson:"mu"`
}
