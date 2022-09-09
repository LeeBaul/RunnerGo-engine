package golink

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/heartbeat"
	"strconv"
	"sync"
	"time"
)

// DisposeScene 对场景进行处理
func DisposeScene(wg *sync.WaitGroup, gid string, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection, options ...int64) {

	log.Logger.Info("djflkasjlfjsjdflks", scene.Uuid)
	nodes := scene.Nodes
	sceneId := strconv.FormatInt(scene.SceneId, 10)
	for _, node := range nodes {
		node.Uuid = scene.Uuid
		wg.Add(1)
		go func(event model.Event) {

			// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，知道上一级所有事件全部完成
			if event.PreList != nil && len(event.PreList) > 0 {
				// 如果该事件上一级有事件, 并且上一级事件中的第一个事件的权重不等于100，那么并发数就等于上一级的并发*权重
				if options != nil && len(options) > 1 {
					var preMaxCon = int64(0)
					for _, request := range nodes {
						for _, tempEvent := range event.PreList {
							if request.Id == tempEvent {
								if request.Weight < 100 && request.Weight > 0 {
									tempWeight := request.Weight
									if tempWeight > preMaxCon {
										preMaxCon = tempWeight
									}
								}
							}
						}
					}

					if preMaxCon != 0 {
						options[1] = int64(float64(preMaxCon) * (float64(event.Weight) / 100))
					} else {
						if event.Weight > 0 && event.Weight < 100 {
							options[1] = int64(float64(options[1]) * (float64(event.Weight) / 100))
						}
					}
				}

				var status = true
				startTime := time.Now().UnixMilli()
				for status {
					for _, eventId := range event.PreList {
						if eventId != "" {
							// 查询上一级状态，如果都完成，则进行该请求，如果未完成，继续查询，直到上一级请求完成
							err, preEventStatus := model.QueryPlanStatus(gid + ":" + sceneId + ":" + eventId + ":status")
							if err != nil {
								status = true
								break
							}
							switch preEventStatus {
							case model.End:
								status = false
							case model.NotRun:
								expiration := 60 * time.Second
								err = model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.NotRun, expiration)
								if err != nil {
									log.Logger.Error("事件状态写入数据库失败", err)
								}
								debugMsg := make(map[string]interface{})
								debugMsg["uuid"] = event.Uuid.String()
								debugMsg["event_id"] = eventId
								debugMsg["status"] = model.NotRun
								debugMsg["next_list"] = event.NextList
								if requestCollection != nil {
									model.Insert(requestCollection, debugMsg)
								}
								return
							}
						}
					}
					if startTime+6000 < time.Now().UnixMilli() {
						status = false
					}
				}
			} else {
				if event.Weight > 0 && event.Weight < 100 {
					if options != nil && len(options) > 0 {
						options[1] = int64(float64(options[1]) * (float64(event.Weight) / 100))
					}

				}
			}

			switch event.Type {
			case model.RequestType:
				event.Api.Uuid = scene.Uuid
				event.Debug = scene.Debug
				if options != nil && len(options) > 0 {
					var requestResults = &model.ResultDataMsg{}
					DisposeRequest(wg, reportMsg, resultDataMsgCh, requestResults, scene.Configuration, event, requestCollection, options[0], options[1])
				} else {
					DisposeRequest(wg, reportMsg, resultDataMsgCh, nil, scene.Configuration, event, requestCollection)
				}

			case model.IfControllerType:
				if v, ok := scene.Configuration.Variable.Load(event.Key); ok {
					event.PerForm(v.(string))
				}
			case model.WaitControllerType:
				log.Logger.Info("..................", event.WaitTime)
				time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
			}

			// 如果该事件下还有事件，那么将该事件得状态发送到redis
			if event.NextList != nil && len(event.NextList) > 0 {
				expiration := 60 * time.Second
				err := model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
				if err != nil {
					log.Logger.Error("事件状态写入数据库失败", err)
				}

			}
		}(node)
	}
}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(wg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, configuration *model.Configuration,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	if wg != nil {
		defer wg.Done()
	}

	api := event.Api
	if event.Debug == true {
		api.Debug = true
	}
	log.Logger.Error("23123123123", event.Debug)
	log.Logger.Error("56756756756756", api.Debug)
	// 计算接口权重，不通过此接口的比例 = 并发数 /（100 - 权重） 比如：150并发，权重为20， 那么不通过此接口口的比例
	if event.Weight < 100 && event.Weight > 0 {
		if options != nil && len(options) > 0 {
			if float64(options[0]) < float64(options[1])*(float64(100-event.Weight)/100) {
				return
			}
		}
	}

	if requestResults != nil {
		requestResults.PlanId = reportMsg.PlanId
		requestResults.MachineIp = heartbeat.LocalIp
		requestResults.MachineName = heartbeat.LocalHost
		requestResults.PlanName = reportMsg.PlanName
		requestResults.EventId = event.Id
		requestResults.SceneId = reportMsg.SceneId
		requestResults.SceneName = reportMsg.SceneName
		requestResults.ReportId = reportMsg.ReportId
		requestResults.ReportName = reportMsg.ReportName
		requestResults.CustomRequestTimeLine = event.CustomRequestTime
		requestResults.ErrorThreshold = event.ErrorThreshold
		requestResults.TargetId = api.TargetId
		requestResults.Name = api.Name
	}

	// 将请求信息中所有用的变量添加到接口变量维护的map中
	api.FindParameterizes()

	// 如果请求中使用的变量在场景设置的全局变量中存在存在，则将其赋值给变量
	if configuration != nil {
		api.ReplaceParameters(configuration)
	}

	// 请求中所有的变量替换未真正的值
	api.ReplaceQueryParameterizes()

	var (
		isSucceed     = false
		errCode       = int64(0)
		requestTime   = uint64(0)
		sendBytes     = uint(0)
		contentLength = uint(0)
		errMsg        = ""
	)
	log.Logger.Error("configuration.Variable", configuration.Variable)
	switch api.TargetType {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg = HttpSend(event, api, configuration.Variable, mongoCollection)
	case model.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, contentLength = webSocketSend(api)
	case model.FormTypeGRPC:
		//isSucceed, errCode, requestTime, sendBytes, contentLength := rpcSend(request)
	default:
		return
	}

	if resultDataMsgCh != nil {
		requestResults.Name = api.Name
		requestResults.RequestTime = requestTime
		requestResults.ErrorType = errCode
		requestResults.IsSucceed = isSucceed
		requestResults.SendBytes = uint64(sendBytes)
		requestResults.ReceivedBytes = uint64(contentLength)
		requestResults.ErrorMsg = errMsg
		resultDataMsgCh <- requestResults
	}

}
