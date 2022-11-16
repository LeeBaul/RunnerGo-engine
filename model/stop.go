package model

type Stop struct {
	TeamId    int64    `json:"team_id" bson:"team_id"`
	PlanId    int64    `json:"plan_id" bson:"plan_id"`
	ReportIds []string `json:"report_ids" bson:"report_ids"`
}

type StopScene struct {
	SceneId int64 `json:"scene_id"`
}
