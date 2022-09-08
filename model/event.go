package model

import uuid "github.com/satori/go.uuid"

type Event struct {
	Id                   string      `json:"id" bson:"id"`
	Uuid                 uuid.UUID   `json:"uuid" bson:"uuid"`
	Type                 string      `json:"type" bson:"type"` //   事件类型 "request" "controller"
	PreList              []string    `json:"pre_list" bson:"pre_list"`
	NextList             []string    `json:"next_list"   bson:"next_list"`
	Weight               int64       `json:"weight" bson:"weight"`                             // 权重，并发分配的比例
	Tag                  bool        `json:"tag" bson:"tag"`                                   // Tps模式下，该标签代表以该接口为准
	ErrorThreshold       float32     `json:"errorThreshold" bson:"errorThreshold"`             // 错误率阈值
	CustomRequestTime    int64       `json:"customRequestTime" bson:"customRequestTime"`       // 自定义响应时间线
	RequestTimeThreshold int64       `json:"requestTimeThreshold" bson:"requestTimeThreshold"` // 响应时间阈值
	Api                  Api         `json:"api" bson:"api"`
	Controller           *Controller `json:"controller" bson:"controller"` // 控制器
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
