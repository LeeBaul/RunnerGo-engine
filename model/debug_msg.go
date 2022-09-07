package model

type DebugMsg struct {
	EventId   string                    `json:"eventId" bson:"eventId"`
	ApiId     int64                     `json:"apiId" bson:"apiId"`
	ApiName   string                    `json:"apiName" bson:"apiName"`
	Request   map[string]interface{}    `json:"request"  bson:"request"`
	Response  map[string]interface{}    `json:"response" bson:"response"`
	Assertion map[string][]AssertionMsg `json:"assertion" bson:"assertion"`
	Regex     []map[string]interface{}  `json:"regex" bson:"regex"`
}
