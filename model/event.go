package model

type Event struct {
	EventType  string      `json:"eventType"`  //   事件类型 "request" "controller"
	Request    *Request    `json:"request"`    //   请求类型
	Controller *Controller `json:"controller"` //   控制器
}
