package scene

import (
	"kp-runner/model/request"
)

// Scene 场景结构体
type Scene struct {
	SceneID       int                    `json:"sceneId"`       // 场景Id
	SceneName     string                 `json:"sceneName"`     // 场景名称
	Requests      []request.Request      `json:"requests"`      //
	Configuration map[string]interface{} `json:"configuration"` // 场景配置
}
