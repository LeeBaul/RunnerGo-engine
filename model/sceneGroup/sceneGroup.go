package sceneGroup

import (
	"kp-runner/model/scene"
)

// SceneGroup 场景组结构体，分组
type SceneGroup struct {
	GroupName string      `json:"groupName"`
	GroupID   int         `json:"groupId"`
	Scene     scene.Scene `json:"scenes"`
}
