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
func DisposeScene(gid string, eventList []model.Event, resultDataMsgCh chan *model.ResultDataMsg,
	planId, planName, sceneId, sceneName, reportId, reportName string,
	configuration *model.Configuration, wg *sync.WaitGroup, sceneVariable *sync.Map, requestCollection *mongo.Collection, options ...int64) {

	for _, event := range eventList {
		wg.Add(1)
		go func(event model.Event) {
			// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，知道上一级所有事件全部完成
			if event.PreEventIdList != nil && len(event.PreEventIdList) > 0 {
				// 如果该事件上一级有事件, 并且上一级事件中的第一个事件的权重不等于100，那么并发数就等于上一级的并发*权重
				for _, request := range eventList {
					if request.EventId == event.PreEventIdList[0] {
						if request.Request.Weight != 100 && request.Request.Weight != 0 {
							options[1] = int64(float64(options[1]) * (float64(request.Request.Weight) / 100))
						}

					}
				}

				var status = true
				for status {
					for _, eventId := range event.PreEventIdList {
						if eventId != "" {
							// 查询上一级状态，如果都完成，则进行该请求，如果未完成，继续查询，直到上一级请求完成
							err, preEventStatus := model.QueryPlanStatus(gid + ":" + planId + ":" + sceneId + ":" + eventId + ":status")
							if err != nil {
								status = true
								break
							}
							if preEventStatus == "true" {
								status = false
							}
						}
					}
				}
			}

			switch event.EventType {
			case model.RequestType:
				var requestResults = &model.ResultDataMsg{}
				var debugMsg = &model.DebugMsg{}
				go DisposeRequest(resultDataMsgCh, planId, planName, sceneId, sceneName, reportId, reportName, configuration, event, wg, requestResults, debugMsg, sceneVariable, requestCollection, options[0], options[1])
			case model.ControllerType:
				go DisposeController(gid, event, eventList, planId, sceneId, sceneVariable, wg, requestCollection, options[0], options[1])
			}

			// 如果该事件下还有事件，那么将该事件得状态发送到redis
			if event.NextEventIdList != nil && len(event.NextEventIdList) > 0 {
				expiration := 10 * time.Second
				err := model.InsertStatus(gid+":"+planId+":"+sceneId+":"+event.EventId+":status", "true", expiration)
				if err != nil {
					log.Logger.Error("事件状态写入数据库失败", err)
				}

			}
		}(event)

	}
}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(resultDataMsgCh chan *model.ResultDataMsg,
	planId, planName, sceneId, sceneName, reportId, reportName string, configuration *model.Configuration,
	event model.Event, wg *sync.WaitGroup, requestResults *model.ResultDataMsg, debugMsg *model.DebugMsg, sceneVariable *sync.Map, requestCollection *mongo.Collection, options ...int64) {
	defer wg.Done()

	request := event.Request

	if event.PreEventIdList != nil {

	}
	// 计算接口权重，不通过此接口的比例 = 并发数 /（100 - 权重） 比如：150并发，权重为20， 那么不通过此几口的比例
	if request.Weight < 100 && request.Weight > 0 {
		if float64(options[0]) < float64(options[1])*(float64(100-request.Weight)/100) {
			return
		}
	}

	if planId != "" {
		requestResults.PlanId = planId
		requestResults.MachineIp = heartbeat.LocalIp
		requestResults.MachineName = heartbeat.LocalHost
		requestResults.PlanName = planName
		requestResults.EventId = event.EventId
		requestResults.SceneId = sceneId
		requestResults.SceneName = sceneName
		requestResults.ReportId = reportId
		requestResults.ReportName = reportName
		requestResults.CustomRequestTimeLine = request.CustomRequestTime
		requestResults.ErrorThreshold = request.ErrorThreshold
	}
	requestResults.ApiId = request.ApiId
	requestResults.ApiName = request.ApiName

	// 将请求信息中所有用的变量添加到接口变量维护的map中
	request.FindParameterizes()

	// 如果请求中使用的变量在场景设置的全局变量中存在存在，则将其赋值给变量
	if configuration != nil {
		request.ReplaceParameters(configuration)
	}

	// 请求中所有的变量替换未真正的值
	request.ReplaceQueryParameterizes()

	var (
		isSucceed     = false
		errCode       = int64(0)
		requestTime   = uint64(0)
		sendBytes     = uint(0)
		contentLength = uint(0)
		errMsg        = ""
	)
	switch request.Form {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg = HttpSend(event.EventId, request, sceneVariable, requestCollection, debugMsg)
	case model.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, contentLength = webSocketSend(request)
	case model.FormTypeGRPC:
		//isSucceed, errCode, requestTime, sendBytes, contentLength := rpcSend(request)
	default:
		return
	}

	requestResults.ApiName = request.ApiName
	requestResults.RequestTime = requestTime
	requestResults.ErrorType = errCode
	requestResults.IsSucceed = isSucceed
	requestResults.SendBytes = uint64(sendBytes)
	requestResults.ReceivedBytes = uint64(contentLength)
	requestResults.ErrorMsg = errMsg
	if resultDataMsgCh != nil {
		resultDataMsgCh <- requestResults
	}

	//if event.NextEvent != nil {
	//	eventList := event.NextEvent
	//	DisposeScene(eventList, resultDataMsgCh, planId, planName, sceneId, sceneName, reportId, reportName, configuration, wg, sceneVariable, requestCollection, options[0], options[1])
	//}

}

// DisposeController 处理控制器
func DisposeController(gid string, event model.Event, eventList []model.Event, planId, sceneId string, sceneVariable *sync.Map, wg *sync.WaitGroup, requestCollection *mongo.Collection, options ...int64) {
	defer wg.Done()
	controller := event.Controller

	switch controller.ControllerType {
	case model.IfControllerType:
		if v, ok := sceneVariable.Load(controller.IfController.Key); ok {
			controller.IfController.PerForm(v.(string))
		}
	case model.CollectionType:
		// 集合点, 待开发
	case model.WaitControllerType: // 等待控制器
		timeWait, _ := strconv.Atoi(controller.WaitController.WaitTime)
		time.Sleep(time.Duration(timeWait) * time.Millisecond)
	}

}
