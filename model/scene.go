package model

import "sync"

// Scene 场景结构体
type Scene struct {
	SceneId       string         `json:"sceneId"`       // 场景Id
	SceneName     string         `json:"sceneName"`     // 场景名称
	CreateTime    int            `json:"create_time"`   // 创建时间
	CreateUUid    string         `json:"create_uuid"`   //
	EventList     []Event        `json:"event_list"`    // 事件列表
	Configuration *Configuration `json:"configuration"` // 场景配置
}

type Configuration struct {
	ParameterizedFile *ParameterizedFile `json:"parameterizedFile"`
	Variable          *sync.Map          `json:"variable"`
	Mu                sync.Mutex         `json:"mu"`
}
