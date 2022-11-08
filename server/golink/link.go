package golink

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/tools"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DisposeScene 对场景进行处理
func DisposeScene(sharedMap *sync.Map, wg, currentWg *sync.WaitGroup, gid string, runType string, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection, options ...int64) {
	nodes := scene.Nodes
	sceneId := fmt.Sprintf("%d", scene.SceneId)

	for _, node := range nodes {
		node.Uuid = scene.Uuid
		wg.Add(1)
		currentWg.Add(1)

		switch runType {
		case model.PlanType:
			go disposePlanNode(sharedMap, scene, sceneId, node, gid, wg, currentWg, reportMsg, resultDataMsgCh, requestCollection, options...)

		case model.SceneType:
			go disposeDebugNode(sharedMap, scene, sceneId, node, gid, wg, currentWg, reportMsg, resultDataMsgCh, requestCollection)
		}

	}
}

// disposePlanNode 处理node节点
func disposePlanNode(sharedMap *sync.Map, scene *model.Scene, sceneId string, event model.Event, gid string, wg, currentWg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection, disOptions ...int64) {
	defer wg.Done()
	defer currentWg.Done()

	var (
		goroutineId int64 // 启动的第几个协程
		machineIp   = ""
		reportId    = ""
	)
	if reportMsg != nil {
		machineIp = reportMsg.MachineIp
		reportId = reportMsg.ReportId
	}
	var eventResult = model.EventResult{}
	// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，直到上一级所有事件全部完成
	if event.PreList != nil && len(event.PreList) > 0 {
		var preMaxConcurrent = int64(0) // 上一级最大并发数
		var preMaxWeight = int64(0)
		// 将上一级事件放入一个map中进行维护
		var preMap = make(map[string]bool)

		for _, eventId := range event.PreList {
			if eventId != "" {
				preMap[eventId] = false
			}
		}

		startTime := time.Now().UnixMilli()

		for len(preMap) > 0 {
			for eventId, _ := range preMap {
				// 查询上一级状态，如果都完成，则进行该请求，如果未完成，继续查询，直到上一级请求完成
				preEventStatus := model.EventResult{}
				if value, ok := sharedMap.Load(machineIp + ":" + reportId + ":" + gid + ":" + sceneId + ":" + eventId + ":status"); !ok {
					break
				} else {
					if preEventStatus, ok = value.(model.EventResult); !ok {
						break
					}

				}
				switch preEventStatus.Status {
				case model.End:
					goroutineId = disOptions[0]
					if preEventStatus.Concurrent >= preMaxConcurrent {
						preMaxConcurrent = preEventStatus.Concurrent
					}

					if event.Type == model.IfControllerType || event.Type == model.WaitControllerType {
						if preEventStatus.Weight >= preMaxWeight {
							preMaxWeight = preEventStatus.Weight
						}
					}
					delete(preMap, eventId)
				case model.NotRun:
					eventResult.Status = model.NotRun
					eventResult.Weight = event.Weight
					sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
					delete(preMap, eventId)
					return
				case model.NotHit:
					eventResult.Status = model.NotRun
					eventResult.Weight = event.Weight
					sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
					delete(preMap, eventId)
					return
				}
			}
			if startTime+60000 < time.Now().UnixMilli() {
				return
			}
		}

		if event.Type == model.WaitControllerType || event.Type == model.IfControllerType {
			event.Weight = preMaxWeight
		}
		if event.Weight > 0 && event.Weight < 100 {
			eventResult.Concurrent = int64(math.Ceil(float64(event.Weight) * float64(preMaxConcurrent) / 100))
		}

		if event.Weight == 100 {
			eventResult.Concurrent = preMaxConcurrent
		}

		// 如果该事件上一级有事件, 并且上一级事件中的第一个事件的权重不等于100，那么并发数就等于上一级的并发*权重

	} else {
		if event.Type == model.WaitControllerType || event.Type == model.IfControllerType {
			event.Weight = 100
		}
		if disOptions != nil && len(disOptions) > 1 {
			if event.Weight == 100 {
				eventResult.Concurrent = disOptions[1]
			}
			if event.Weight > 0 && event.Weight < 100 {
				eventResult.Concurrent = int64(math.Ceil(float64(disOptions[1]) * (float64(event.Weight) / float64(100))))
			}
		}

	}

	if eventResult.Concurrent == 0 {
		eventResult.Status = model.NotRun
		eventResult.Weight = event.Weight
		sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
		return
	}
	if goroutineId > eventResult.Concurrent {
		eventResult.Status = model.NotRun
		eventResult.Weight = event.Weight
		sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
		return
	}

	event.TeamId = scene.TeamId
	event.Debug = scene.Debug
	event.ReportId = scene.ReportId
	switch event.Type {
	case model.RequestType:
		event.Api.Uuid = scene.Uuid
		var requestResults = &model.ResultDataMsg{}
		DisposeRequest(reportMsg, resultDataMsgCh, requestResults, scene.Configuration, event, requestCollection, goroutineId, eventResult.Concurrent)
		eventResult.Status = model.End
		eventResult.Weight = event.Weight
		sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
	case model.IfControllerType:
		keys := tools.FindAllDestStr(event.Var, "{{(.*?)}}")
		if len(keys) > 0 {
			for _, val := range keys {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						if kv.Value != nil {
							event.Var = strings.Replace(event.Var, val[0], kv.Value.(string), -1)
						}

					}
				}
			}
		}
		values := tools.FindAllDestStr(event.Val, "{{(.*?)}}")
		if len(values) > 0 {
			for _, val := range values {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						if kv.Value != nil {
							event.Val = strings.Replace(event.Val, val[0], kv.Value.(string), -1)
						}

					}
				}
			}
		}
		var result = model.Failed
		var msg = ""

		var temp = false
		for _, kv := range scene.Configuration.Variable {
			if kv.Key == event.Var {
				temp = true
				if kv.Value != nil {
					result, msg = event.PerForm(kv.Value.(string))
				}

				break
			}
		}
		if temp == false {
			result, msg = event.PerForm(event.Var)
		}
		if event.Debug != "" {
			debugMsg := make(map[string]interface{})
			debugMsg["uuid"] = event.Uuid.String()
			debugMsg["event_id"] = event.Id
			debugMsg["status"] = result
			debugMsg["msg"] = msg
			debugMsg["type"] = model.IfControllerType
			debugMsg["report_id"] = event.ReportId
			debugMsg["next_list"] = event.NextList
			if requestCollection != nil {
				model.Insert(requestCollection, debugMsg)
			}
		}

		if result == model.Failed {
			eventResult.Status = model.End
			eventResult.Weight = event.Weight
			sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
		} else {
			eventResult.Status = model.End
			eventResult.Weight = event.Weight
			sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
		}
	case model.WaitControllerType:
		time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
		if scene.Debug != "" {
			debugMsg := make(map[string]interface{})
			debugMsg["uuid"] = event.Uuid.String()
			debugMsg["event_id"] = event.Id
			debugMsg["report_id"] = event.ReportId
			debugMsg["status"] = model.Success
			debugMsg["type"] = model.WaitControllerType
			debugMsg["msg"] = "等待了" + strconv.Itoa(event.WaitTime) + "毫秒"
			debugMsg["next_list"] = event.NextList
			if requestCollection != nil {
				model.Insert(requestCollection, debugMsg)
			}
		}
		eventResult.Status = model.End
		eventResult.Weight = event.Weight
		sharedMap.Store(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
	}

}

func disposeDebugNode(sharedMap *sync.Map, scene *model.Scene, sceneId string, event model.Event, gid string, wg, currentWg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {
	defer wg.Done()
	defer currentWg.Done()

	var (
		machineIp = ""
	)
	if reportMsg != nil {
		machineIp = reportMsg.MachineIp
	}
	var eventResult = model.EventResult{}
	// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，直到上一级所有事件全部完成
	if event.PreList != nil && len(event.PreList) > 0 {
		// 将上一级事件放入一个map中进行维护
		var preMap = make(map[string]bool)
		for _, eventId := range event.PreList {
			if eventId != "" {
				preMap[eventId] = false
			}
		}
		startTime := time.Now().UnixMilli()
		for len(preMap) > 0 {
			for eventId, _ := range preMap {
				// 查询上一级状态，如果都完成，则进行该请求，如果未完成，继续查询，直到上一级请求完成
				preEventStatus := model.EventResult{}
				if value, ok := sharedMap.Load(machineIp + ":" + gid + ":" + sceneId + ":" + eventId + ":status"); !ok {
					break
				} else {
					if preEventStatus, ok = value.(model.EventResult); !ok {
						break
					}
				}
				switch preEventStatus.Status {
				case model.End:
					delete(preMap, eventId)
				case model.NotRun:
					eventResult.Status = model.NotRun
					sharedMap.Store(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
					delete(preMap, eventId)
					debugMsg := make(map[string]interface{})
					debugMsg["uuid"] = event.Uuid.String()
					debugMsg["event_id"] = event.Id
					debugMsg["status"] = model.NotRun
					debugMsg["msg"] = "未运行"
					debugMsg["type"] = event.Type
					debugMsg["next_list"] = event.NextList
					if requestCollection != nil {
						model.Insert(requestCollection, debugMsg)
					}

					return
				case model.NotHit:

					eventResult.Status = model.NotRun
					sharedMap.Store(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
					delete(preMap, eventId)
					debugMsg := make(map[string]interface{})
					debugMsg["uuid"] = event.Uuid.String()
					debugMsg["event_id"] = event.Id
					debugMsg["status"] = model.NotRun
					debugMsg["msg"] = "未运行"
					debugMsg["type"] = event.Type
					debugMsg["next_list"] = event.NextList
					if requestCollection != nil {
						model.Insert(requestCollection, debugMsg)
					}

					return
				}
			}
			if startTime+60000 < time.Now().UnixMilli() {
				return
			}
		}

	}

	event.TeamId = scene.TeamId
	event.Debug = scene.Debug
	event.ReportId = scene.ReportId
	switch event.Type {
	case model.RequestType:
		event.Api.Uuid = scene.Uuid
		DisposeRequest(reportMsg, resultDataMsgCh, nil, scene.Configuration, event, requestCollection)
		eventResult.Status = model.End
		sharedMap.Store(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
	case model.IfControllerType:
		keys := tools.FindAllDestStr(event.Var, "{{(.*?)}}")

		if len(keys) > 0 {
			for _, val := range keys {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						if kv.Value != nil {
							event.Var = strings.Replace(event.Var, val[0], kv.Value.(string), -1)
						}

					}
				}
			}
		}
		values := tools.FindAllDestStr(event.Val, "{{(.*?)}}")
		if len(values) > 0 {
			for _, val := range values {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						if kv.Value != nil {
							event.Val = strings.Replace(event.Val, val[0], kv.Value.(string), -1)
						}
					}
				}
			}
		}
		var result = model.Failed
		var msg = ""

		var temp = false
		for _, kv := range scene.Configuration.Variable {
			if kv.Key == event.Var {
				temp = true
				if kv.Value != nil {
					result, msg = event.PerForm(kv.Value.(string))
				}
				break
			}
		}
		if temp == false {
			result, msg = event.PerForm(event.Var)
		}
		if event.Debug != "" {
			debugMsg := make(map[string]interface{})
			debugMsg["uuid"] = event.Uuid.String()
			debugMsg["event_id"] = event.Id
			debugMsg["status"] = result
			debugMsg["msg"] = msg
			debugMsg["type"] = model.IfControllerType
			debugMsg["next_list"] = event.NextList
			if requestCollection != nil {
				model.Insert(requestCollection, debugMsg)
			}
		}

		if result == model.Failed {
			eventResult.Status = model.NotHit
			sharedMap.Store(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
		} else {
			eventResult.Status = model.End
			sharedMap.Store(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
		}
	case model.WaitControllerType:
		time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
		if scene.Debug != "" {
			debugMsg := make(map[string]interface{})
			debugMsg["uuid"] = event.Uuid.String()
			debugMsg["event_id"] = event.Id
			debugMsg["status"] = model.Success
			debugMsg["type"] = model.WaitControllerType
			debugMsg["msg"] = "等待了" + strconv.Itoa(event.WaitTime) + "毫秒"
			debugMsg["next_list"] = event.NextList
			if requestCollection != nil {
				model.Insert(requestCollection, debugMsg)
			}
		}
		eventResult.Status = model.End
		sharedMap.Store(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", eventResult)
	}

}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, configuration *model.Configuration,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	api := event.Api
	if api.Debug == "" {
		api.Debug = event.Debug
	}

	if requestResults != nil {
		requestResults.PlanId = reportMsg.PlanId
		requestResults.PlanName = reportMsg.PlanName
		requestResults.EventId = event.Id
		requestResults.PercentAge = event.PercentAge
		requestResults.ResponseThreshold = event.ResponseThreshold
		requestResults.TeamId = event.TeamId
		requestResults.SceneId = reportMsg.SceneId
		requestResults.MachineIp = reportMsg.MachineIp
		requestResults.Concurrency = options[1]
		requestResults.SceneName = reportMsg.SceneName
		requestResults.ReportId = reportMsg.ReportId
		requestResults.ReportName = reportMsg.ReportName
		requestResults.PercentAge = event.PercentAge
		requestResults.RequestThreshold = event.RequestThreshold
		requestResults.ResponseThreshold = event.ResponseThreshold
		requestResults.ErrorThreshold = event.ErrorThreshold
		requestResults.TargetId = api.TargetId
		requestResults.Name = api.Name
		requestResults.MachineNum = reportMsg.MachineNum
	}

	// 将请求信息中所有用的变量添加到接口变量维护的map中
	api.FindParameterizes()

	// 如果请求中使用的变量在场景设置的全局变量中存在存在，则将其赋值给变量
	if configuration != nil {
		api.ReplaceParameters(configuration)
	}

	// 请求中所有的变量替换成真正的值
	api.ReplaceQueryParameterizes()

	var (
		isSucceed     = false
		errCode       = int64(0)
		requestTime   = uint64(0)
		sendBytes     = float64(0)
		receivedBytes = float64(0)
		errMsg        = ""
	)
	switch api.TargetType {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, receivedBytes, errMsg = HttpSend(event, api, configuration, mongoCollection)
	case model.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, receivedBytes = webSocketSend(api)
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
		requestResults.SendBytes = sendBytes
		requestResults.ReceivedBytes = receivedBytes
		requestResults.Timestamp = time.Now().UnixMilli()
		requestResults.ErrorMsg = errMsg
		resultDataMsgCh <- requestResults
	}

	if resultDataMsgCh == nil {
		log.Logger.Debug("接口: ", event.Api.Name, "   调试结束！")
	}

}

//func executionStrategy(event model.Event, nodes []model.Event, wgTemp *sync.WaitGroup, disOptions ...int64) {
//	gid1 := tools.GetGid()
//	log.Logger.Debug("gid: ", gid1, "           disOptions", disOptions)
//	// 如果该事件上一级有事件，那么就一直查询上一级事件的状态，直到上一级所有事件全部完成
//	if event.PreList != nil && len(event.PreList) > 0 {
//		// 如果该事件上一级有事件, 并且上一级事件中的第一个事件的权重不等于100，那么并发数就等于上一级的并发*权重
//		if disOptions != nil && len(disOptions) > 1 {
//			// 上级事件的最大并发数
//			var preMaxCon = int64(0)
//			for _, request := range nodes {
//				for _, tempEvent := range event.PreList {
//					if request.Id == tempEvent {
//						if request.Weight < 100 && request.Weight > 0 {
//							tempWeight := request.Weight
//							if tempWeight > preMaxCon {
//								preMaxCon = tempWeight
//							}
//						}
//					}
//				}
//			}
//
//			if preMaxCon != 0 {
//				disOptions[1] = int64(float64(preMaxCon) * (float64(event.Weight) / 100))
//			} else {
//				if event.Weight > 0 && event.Weight < 100 {
//					disOptions[1] = int64(float64(disOptions[1]) * (float64(event.Weight) / 100))
//				}
//			}
//		}
//
//		var preMap = make(map[string]bool)
//		for _, eventId := range event.PreList {
//			if eventId != "" {
//				preMap[eventId] = false
//			}
//		}
//
//		startTime := time.Now().UnixMilli()
//
//		for len(preMap) > 0 {
//			if runType == "scene" {
//				err, sceneStatus := model.QuerySceneStatus(sceneId + ":status")
//				if err == nil && sceneStatus == "stop" {
//					debugMsg := make(map[string]interface{})
//					debugMsg["uuid"] = event.Uuid.String()
//					debugMsg["event_id"] = event.Id
//					debugMsg["status"] = model.NotRun
//					debugMsg["type"] = event.Type
//					debugMsg["report_id"] = event.ReportId
//					debugMsg["next_list"] = event.NextList
//					if requestCollection != nil {
//						model.Insert(requestCollection, debugMsg)
//					}
//					wgTemp.Done()
//					return
//				}
//			}
//			for eventId, _ := range preMap {
//				if eventId != "" {
//					// 查询上一级状态，如果都完成，则进行该请求，如果未完成，继续查询，直到上一级请求完成
//					err, preEventStatus := model.QueryPlanStatus(gid + ":" + sceneId + ":" + eventId + ":status")
//					if err != nil {
//						break
//					}
//					switch preEventStatus {
//					case model.End:
//						delete(preMap, eventId)
//					case model.NotRun:
//						expiration := 60 * time.Second
//						err = model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.NotRun, expiration)
//						if err != nil {
//							log.Logger.Error("事件状态写入数据库失败", err)
//						}
//						debugMsg := make(map[string]interface{})
//						debugMsg["uuid"] = event.Uuid.String()
//						debugMsg["event_id"] = event.Id
//						debugMsg["status"] = model.NotRun
//						debugMsg["type"] = event.Type
//						debugMsg["report_id"] = event.ReportId
//						debugMsg["next_list"] = event.NextList
//						if requestCollection != nil {
//							model.Insert(requestCollection, debugMsg)
//						}
//						wgTemp.Done()
//						return
//					case model.NotHit:
//						expiration := 60 * time.Second
//						err = model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.NotRun, expiration)
//						if err != nil {
//							log.Logger.Error("事件状态写入数据库失败", err)
//						}
//						debugMsg := make(map[string]interface{})
//						debugMsg["uuid"] = event.Uuid.String()
//						debugMsg["event_id"] = event.Id
//						debugMsg["type"] = event.Type
//						debugMsg["status"] = model.NotRun
//						debugMsg["report_id"] = event.ReportId
//						debugMsg["next_list"] = event.NextList
//						if requestCollection != nil {
//							model.Insert(requestCollection, debugMsg)
//						}
//						wgTemp.Done()
//						return
//					}
//				}
//			}
//			if startTime+60000 < time.Now().UnixMilli() {
//				break
//			}
//		}
//	} else {
//		if event.Weight > 0 && event.Weight < 100 {
//			if disOptions != nil && len(disOptions) > 0 && event.Id == "07f9822c-4fbb-4453-8bbc-9e90276b23ce" {
//				log.Logger.Debug("gid: ", gid1, "event:          ", event.Id, "             前             ", disOptions)
//				disOptions[1] = int64(math.Ceil(float64(disOptions[1]) * (float64(event.Weight) / float64(100))))
//				log.Logger.Debug("gid: ", gid1, "event:          ", event.Id, "后            ", disOptions)
//			}
//
//		}
//	}
//
//	if disOptions != nil {
//		if disOptions[1] <= 0 {
//			return
//		} else if disOptions[1] < 100 {
//			if disOptions[1] != options[1] && disOptions[0] > disOptions[1] {
//				return
//			}
//		}
//	}
//
//	event.TeamId = scene.TeamId
//	event.Debug = scene.Debug
//	event.ReportId = scene.ReportId
//	switch event.Type {
//	case model.RequestType:
//		event.Api.Uuid = scene.Uuid
//		if disOptions != nil && len(disOptions) > 0 {
//			var requestResults = &model.ResultDataMsg{}
//			DisposeRequest(wgTemp, reportMsg, resultDataMsgCh, requestResults, scene.Configuration, event, requestCollection, disOptions[0], disOptions[1])
//		} else {
//			DisposeRequest(wgTemp, reportMsg, resultDataMsgCh, nil, scene.Configuration, event, requestCollection)
//		}
//
//		expiration := 60 * time.Second
//		err := model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
//		if err != nil {
//			log.Logger.Error("事件状态写入redis数据库失败", err)
//		}
//	case model.IfControllerType:
//		keys := tools.FindAllDestStr(event.Var, "{{(.*?)}}")
//		if len(keys) > 0 {
//			for _, val := range keys {
//				for _, kv := range scene.Configuration.Variable {
//					if kv.Key == val[1] {
//						event.Var = strings.Replace(event.Var, val[0], kv.Value, -1)
//					}
//				}
//			}
//		}
//		values := tools.FindAllDestStr(event.Val, "{{(.*?)}}")
//		if len(values) > 0 {
//			for _, val := range values {
//				for _, kv := range scene.Configuration.Variable {
//					if kv.Key == val[1] {
//						event.Val = strings.Replace(event.Val, val[0], kv.Value, -1)
//					}
//				}
//			}
//		}
//		var result = model.Failed
//		var msg = ""
//
//		var temp = false
//		for _, kv := range scene.Configuration.Variable {
//
//			if kv.Key == event.Var {
//				temp = true
//				result, msg = event.PerForm(kv.Value)
//				break
//			}
//		}
//		if temp == false {
//			result, msg = event.PerForm(event.Var)
//		}
//		if event.Debug != "" {
//			debugMsg := make(map[string]interface{})
//			debugMsg["uuid"] = event.Uuid.String()
//			debugMsg["event_id"] = event.Id
//			debugMsg["status"] = result
//			debugMsg["msg"] = msg
//			debugMsg["type"] = model.IfControllerType
//			debugMsg["report_id"] = event.ReportId
//			debugMsg["next_list"] = event.NextList
//			if requestCollection != nil {
//				model.Insert(requestCollection, debugMsg)
//			}
//		}
//
//		expiration := 60 * time.Second
//		if result == model.Failed {
//			err := model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.NotHit, expiration)
//			if err != nil {
//				log.Logger.Error("事件状态写入redis数据库失败", err)
//			}
//		} else {
//			err := model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
//			if err != nil {
//				log.Logger.Error("事件状态写入redis数据库失败", err)
//			}
//		}
//		wgTemp.Done()
//	case model.WaitControllerType:
//		time.Sleep(time.Duration(event.WaitTime) * time.Millisecond)
//		if scene.Debug != "" {
//			debugMsg := make(map[string]interface{})
//			debugMsg["uuid"] = event.Uuid.String()
//			debugMsg["event_id"] = event.Id
//			debugMsg["report_id"] = event.ReportId
//			debugMsg["status"] = model.Success
//			debugMsg["type"] = model.WaitControllerType
//			debugMsg["msg"] = "等待了" + strconv.Itoa(event.WaitTime) + "毫秒"
//			debugMsg["next_list"] = event.NextList
//			if requestCollection != nil {
//				model.Insert(requestCollection, debugMsg)
//			}
//		}
//		expiration := 60 * time.Second
//		err := model.InsertStatus(gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
//		if err != nil {
//			log.Logger.Error("事件状态写入数据库失败", err)
//		}
//		wgTemp.Done()
//	}
//
//}
