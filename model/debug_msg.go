package model

type DebugMsg struct {
	EventId   string            `json:"eventId" bson:"eventId""`
	ApiId     int64             `json:"apiId" bson:"apiId"`
	ApiName   string            `json:"apiName" bson:"apiName"`
	Request   map[string]string `json:"request"  bson:"request"`
	Response  map[string]string `json:"response" bson:"response"`
	Assertion map[string]string `json:"assertion" bson:"assertion"`
}
