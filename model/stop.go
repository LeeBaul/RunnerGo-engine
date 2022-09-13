package model

type Stop struct {
	ReportIds []string `json:"report_ids" bson:"report_ids"`
}
