package model

type Event struct {
	EventType       string      `json:"eventType" bson:"EventType"` //   事件类型 "request" "controller"
	EventId         string      `json:"eventId" bson:"EventId"`
	PreEventIdList  []string    `json:"preEventIdList" bson:"preEventIdList"`
	NextEventIdList []string    `json:"nextEventIdList"   bson:"nextEventIdList"`
	Request         Request     `json:"request" bson:"request"`
	Controller      *Controller `json:"controller" bson:"controller"` // 控制器
}

type EventStatus struct {
	EventType string `json:"eventType" bson:"eventType"`
	EventId   string `json:"eventId" bson:"eventId"`
	Status    bool   `json:"status" bson:"status"`
}

//
//func GrpcReplaceParameterizes(r *pb.Request, globalVariable *sync.Map) {
//	for k, v := range r.Parameterizes {
//		// 查找header的key中是否存在变量{{****}}
//		keys := tools.FindAllDestStr(k, "{{(.*?)}}")
//		if keys != nil {
//			delete(r.Parameterizes, k)
//			for _, realKey := range keys {
//				if value, ok := globalVariable.Load(realKey[1]); ok {
//					k = strings.Replace(k, realKey[0], value.(string), -1)
//				}
//			}
//			r.Parameterizes[k] = v
//		}
//
//		values := tools.FindAllDestStr(v.String(), "{{(.*?)}}")
//		if values != nil {
//			for _, realValue := range values {
//				if value, ok := globalVariable.Load(realValue[1]); ok {
//					v.Value = []byte(strings.Replace(v.String(), realValue[0], value.(string), -1))
//				}
//			}
//			r.Parameterizes[k] = v
//		}
//	}
//}
