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

	nodes := scene.Nodes
	planId := reportMsg.PlanId
	sceneId := reportMsg.SceneId
	for _, node := range nodes {
		wg.Add(1)
		go func(event model.Event) {
			// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，知道上一级所有事件全部完成
			if event.PreList != nil && len(event.PreList) > 0 {
				// 如果该事件上一级有事件, 并且上一级事件中的第一个事件的权重不等于100，那么并发数就等于上一级的并发*权重
				for _, request := range nodes {
					if event.Id == event.PreList[0] {
						if request.Weight != 100 && request.Weight != 0 {
							options[1] = int64(float64(options[1]) * (float64(request.Weight) / 100))
						}

					}
				}

				var status = true
				for status {
					for _, eventId := range event.PreList {
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

			switch event.Type {
			case model.RequestType:
				var requestResults = &model.ResultDataMsg{}
				go DisposeRequest(wg, reportMsg, resultDataMsgCh, requestResults, scene.Configuration, event, requestCollection, options[0], options[1])
			case model.ControllerType:
				go DisposeController(wg, reportMsg, scene.Configuration, event, requestCollection, options[0], options[1])
			}

			// 如果该事件下还有事件，那么将该事件得状态发送到redis
			if event.NextList != nil && len(event.NextList) > 0 {
				expiration := 10 * time.Second
				err := model.InsertStatus(gid+":"+planId+":"+sceneId+":"+event.Id+":status", "true", expiration)
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

	// 计算接口权重，不通过此接口的比例 = 并发数 /（100 - 权重） 比如：150并发，权重为20， 那么不通过此几口的比例
	if event.Weight < 100 && event.Weight > 0 {
		if float64(options[0]) < float64(options[1])*(float64(100-event.Weight)/100) {
			return
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
	log.Logger.Info("api", event.Api)
	log.Logger.Info("api.TargetType", event.Api.TargetType)
	switch api.TargetType {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg = HttpSend(event.Id, api, configuration.Variable, mongoCollection)
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

// DisposeController 处理控制器
func DisposeController(wg *sync.WaitGroup, reportMsg *model.ResultDataMsg, configuration *model.Configuration, event model.Event, requestCollection *mongo.Collection, options ...int64) {
	defer wg.Done()
	controller := event.Controller
	switch controller.ControllerType {
	case model.IfControllerType:
		if v, ok := configuration.Variable.Load(controller.IfController.Key); ok {
			controller.IfController.PerForm(v.(string))
		}
	case model.CollectionType:
		// 集合点, 待开发
	case model.WaitControllerType: // 等待控制器
		timeWait, _ := strconv.Atoi(controller.WaitController.WaitTime)
		time.Sleep(time.Duration(timeWait) * time.Millisecond)
	}

}
