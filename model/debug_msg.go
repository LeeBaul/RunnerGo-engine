package model

type DebugMsg struct {
	Request   map[string]string `json:"request"  bson:"request"`
	Response  map[string]string `json:"response" bson:"response"`
	Assertion map[string]string `json:"assertion" bson:"assertion"`
}
