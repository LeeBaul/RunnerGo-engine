package model

type Stop struct {
	ReportIds []string `json:"report_ids" bson:"report_ids"`
}

type StopScene struct {
	SceneId int64 `json:"scene_id"`
}
