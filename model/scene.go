package model

// Scene 场景结构体
type Scene struct {
	SceneId       string                 `json:"sceneId"`     // 场景Id
	SceneName     string                 `json:"name"`        // 场景名称
	CreateTime    int                    `json:"create_time"` // 创建时间
	CreateUUid    string                 `json:"create_uuid"`
	EventList     []Event                `json:"event_list"`    // 事件列表
	Configuration map[string]interface{} `json:"configuration"` // 场景配置
}
