package golink

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"strconv"
	"sync"
	"time"
)

// DisposeScene 对场景进行处理
func DisposeScene(gid string, eventList []model.Event, resultDataMsgCh chan *model.ResultDataMsg,
	planId, planName, sceneId, sceneName, reportId, reportName string,
	configuration *model.Configuration, wg *sync.WaitGroup, sceneVariable *sync.Map, requestCollection *mongo.Collection, options ...int64) {

	for _, event := range eventList {
		switch event.EventType {
		case model.RequestType:
			var requestResults = &model.ResultDataMsg{}
			var debugMsg = &model.DebugMsg{}
			wg.Add(1)
			go DisposeRequest(gid, resultDataMsgCh, planId, planName, sceneId, sceneName, reportId, reportName, configuration, event, wg, requestResults, debugMsg, sceneVariable, requestCollection, options[0], options[1])
		case model.ControllerType:
			wg.Add(1)
			go DisposeController(gid, event, planId, sceneId, sceneVariable, wg, requestCollection, options[0], options[1])
		}
	}
}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(gid string, resultDataMsgCh chan *model.ResultDataMsg,
	planId, planName, sceneId, sceneName, reportId, reportName string, configuration *model.Configuration,
	event model.Event, wg *sync.WaitGroup, requestResults *model.ResultDataMsg, debugMsg *model.DebugMsg, sceneVariable *sync.Map, requestCollection *mongo.Collection, options ...int64) {
	defer wg.Done()
	// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，知道上一级所有事件全部完成
	if event.PreEventIdList != nil && len(event.PreEventIdList) > 0 {
		var status = true
		for status {
			for _, eventId := range event.PreEventIdList {
				if eventId != "" {
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

	request := event.Request

	if request.Weight != 100 && request.Weight != 0 {
		proportion := options[1] / (100 - request.Weight)
		if options[0]%proportion == 0 {
			return
		}

	}

	if planId != "" {
		requestResults.PlanId = planId
		requestResults.PlanName = planName
		requestResults.SceneId = sceneId
		requestResults.SceneName = sceneName
		requestResults.ReportId = reportId
		requestResults.ReportName = reportName
		requestResults.CustomRequestTimeLine = request.CustomRequestTime
		requestResults.ErrorThreshold = request.ErrorThreshold
	}
	requestResults.ApiId = request.ApiId
	requestResults.ApiName = request.ApiName

	// 如果接口的变量中没有全局变量中的key，那么将全局变量添加到接口变量中
	if sceneVariable != nil {
		sceneVariable.Range(func(key, value any) bool {

			return true
		})
	}

	request.ReplaceBodyParameterizes(configuration.Variable)
	request.ReplaceUrlParameterizes(configuration.Variable)

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
	if planId != "" {
		resultDataMsgCh <- requestResults
	}

	// 如果该事件下还有事件，那么将该事件得状态发送到redis
	if event.NextEventIdList != nil && len(event.NextEventIdList) > 0 {
		expiration := 10 * time.Second
		err := model.InsertStatus(gid+":"+planId+":"+sceneId+":"+event.EventId+":status", "true", expiration)
		if err != nil {
			log.Logger.Error("事件状态写入数据库失败", err)
		}

	}

	//if event.NextEvent != nil {
	//	eventList := event.NextEvent
	//	DisposeScene(eventList, resultDataMsgCh, planId, planName, sceneId, sceneName, reportId, reportName, configuration, wg, sceneVariable, requestCollection, options[0], options[1])
	//}

}

// DisposeController 处理控制器
func DisposeController(gid string, event model.Event, planId, sceneId string, sceneVariable *sync.Map, wg *sync.WaitGroup, requestCollection *mongo.Collection, options ...int64) {
	// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，知道上一级所有事件全部完成
	defer wg.Done()
	if event.PreEventIdList != nil && len(event.NextEventIdList) > 0 {
		var status = true
		for status {
			for _, eventId := range event.PreEventIdList {
				if eventId != "" {
					err, preEventStatus := model.QueryPlanStatus(gid + ":" + planId + ":" + sceneId + ":" + eventId + ":" + "status")
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

	log.Logger.Info("gid", gid, "eventId", event.EventId)
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

	// 如果该事件下还有事件，那么将该事件得状态发送到redis
	if event.NextEventIdList != nil && len(event.NextEventIdList) > 0 {
		expiration := 10 * time.Second
		err := model.InsertStatus(gid+":"+planId+":"+sceneId+":"+event.EventId+":status", "true", expiration)
		if err != nil {
			log.Logger.Error("事件状态写入数据库失败", err)
		}
	}
}
