package model

type EventResult struct {
	Status     string `json:"status"`
	Concurrent int64  `json:"concurrent"`
}
