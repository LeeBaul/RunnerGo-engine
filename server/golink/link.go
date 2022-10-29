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
func DisposeScene(wg, currentWg *sync.WaitGroup, gid string, runType string, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection, options ...int64) {
	nodes := scene.Nodes
	sceneId := fmt.Sprintf("%d", scene.SceneId)
	for _, node := range nodes {
		node.Uuid = scene.Uuid
		wg.Add(1)
		currentWg.Add(1)
		switch runType {
		case model.PlanType:
			go disposePlanNode(scene, sceneId, node, gid, wg, currentWg, reportMsg, resultDataMsgCh, requestCollection, options...)
		case model.SceneType:
			go disposeDebugNode(scene, sceneId, node, gid, wg, currentWg, reportMsg, resultDataMsgCh, requestCollection)

		}

	}
}

// disposePlanNode 处理node节点
func disposePlanNode(scene *model.Scene, sceneId string, event model.Event, gid string, wg, currentWg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection, disOptions ...int64) {
	defer wg.Done()
	defer currentWg.Done()

	var (
		goroutineId int64 // 启动的第几个协程
		current     int64 // 并发数
		machineIp   = ""
		reportId    = ""
	)
	if reportMsg != nil {
		machineIp = reportMsg.MachineIp
		reportId = reportMsg.ReportId
	}
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
				err, preEventStatus := model.QueryPlanStatus(machineIp + ":" + reportId + ":" + gid + ":" + sceneId + ":" + eventId + ":status")
				if err != nil {
					break
				}
				switch preEventStatus {
				case model.End:
					delete(preMap, eventId)
				case model.NotRun:
					expiration := 60 * time.Second
					err = model.InsertStatus(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", model.NotRun, expiration)
					if err != nil {
						log.Logger.Error("事件状态写入数据库失败", err)
					}
					delete(preMap, eventId)
				case model.NotHit:
					expiration := 60 * time.Second
					err = model.InsertStatus(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", model.NotRun, expiration)
					if err != nil {
						log.Logger.Error("事件状态写入数据库失败", err)
					}
					delete(preMap, eventId)
				}
			}
			if startTime+60000 < time.Now().UnixMilli() {
				return
			}
		}

		// 如果该事件上一级有事件, 并且上一级事件中的第一个事件的权重不等于100，那么并发数就等于上一级的并发*权重
		if disOptions != nil && len(disOptions) > 1 {

			// 上级事件的最大并发数
			goroutineId = disOptions[0]
			var preMaxCurrent = int64(0)
			// 从redis获取到上一级事件中的最大并发数
			var s = false

			for !s {
				for _, tempEvent := range event.PreList {

					err, result := model.QueryPlanStatus(machineIp + ":" + reportId + ":" + tempEvent + ":current")
					if err != nil {
						s = false
						break
					}

					tempMaxCurrent, err := strconv.ParseInt(result, 10, 64)
					if err != nil {
						s = false
						break
					}
					if preMaxCurrent <= tempMaxCurrent {
						preMaxCurrent = tempMaxCurrent
					}
					s = true
				}
			}
			// 将上一级的最大并发数赋值给该接口
			current = preMaxCurrent
			if current <= 0 {
				return
			}
			if event.Weight == 100 {
				current = preMaxCurrent
			}

			if event.Weight > 0 && event.Weight < 100 {
				current = int64(float64(preMaxCurrent) * (float64(event.Weight) / 100))
			}

		}

	} else {
		if disOptions != nil && len(disOptions) > 1 {
			if event.Weight == 100 {
				current = disOptions[1]
			}
			if event.Weight > 0 && event.Weight < 100 {
				current = int64(math.Ceil(float64(disOptions[1]) * (float64(event.Weight) / float64(100))))
			}
			if event.Weight == 100 {
				current = disOptions[1]
			}
		}

	}

	if event.NextList != nil && len(event.NextList) > 0 {
		// 将该接口的并发数写入到redis当中，由nextList中的接口去查询并计算自己的并发数
		result := fmt.Sprintf("%d", current)
		expiration := 60 * time.Second
		err := model.InsertStatus(machineIp+":"+reportId+":"+event.Id+":"+"current", result, expiration)
		if err != nil {
			log.Logger.Error(event.Id, " ：并发数状态写入redis失败：  ", err)
		}
	}
	event.TeamId = scene.TeamId
	event.Debug = scene.Debug
	event.ReportId = scene.ReportId
	switch event.Type {
	case model.RequestType:
		event.Api.Uuid = scene.Uuid
		var requestResults = &model.ResultDataMsg{}
		DisposeRequest(nil, reportMsg, resultDataMsgCh, requestResults, scene.Configuration, event, requestCollection, goroutineId, current)
		expiration := 60 * time.Second
		err := model.InsertStatus(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
		if err != nil {
			log.Logger.Error("事件状态写入redis数据库失败", err)
		}
	case model.IfControllerType:
		keys := tools.FindAllDestStr(event.Var, "{{(.*?)}}")

		if len(keys) > 0 {
			for _, val := range keys {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						event.Var = strings.Replace(event.Var, val[0], kv.Value, -1)
					}
				}
			}
		}
		values := tools.FindAllDestStr(event.Val, "{{(.*?)}}")
		if len(values) > 0 {
			for _, val := range values {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						event.Val = strings.Replace(event.Val, val[0], kv.Value, -1)
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
				result, msg = event.PerForm(kv.Value)
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

		expiration := 60 * time.Second
		if result == model.Failed {
			err := model.InsertStatus(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", model.NotHit, expiration)
			if err != nil {
				log.Logger.Error("事件状态写入redis数据库失败", err)
			}
		} else {
			err := model.InsertStatus(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
			if err != nil {
				log.Logger.Error("事件状态写入redis数据库失败", err)
			}
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
		expiration := 60 * time.Second
		err := model.InsertStatus(machineIp+":"+reportId+":"+gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
		if err != nil {
			log.Logger.Error("事件状态写入数据库失败", err)
		}
	}

}

func disposeDebugNode(scene *model.Scene, sceneId string, event model.Event, gid string, wg, currentWg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {
	defer wg.Done()
	defer currentWg.Done()

	var (
		machineIp = ""
	)
	if reportMsg != nil {
		machineIp = reportMsg.MachineIp
	}
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
			err, sceneStatus := model.QuerySceneStatus(sceneId + ":status")
			if err == nil && sceneStatus == "stop" {
				debugMsg := make(map[string]interface{})
				debugMsg["uuid"] = event.Uuid.String()
				debugMsg["event_id"] = event.Id
				debugMsg["status"] = model.NotRun
				debugMsg["type"] = event.Type
				debugMsg["report_id"] = event.ReportId
				debugMsg["next_list"] = event.NextList
				if requestCollection != nil {
					model.Insert(requestCollection, debugMsg)
				}
				return
			}

			for eventId, _ := range preMap {
				// 查询上一级状态，如果都完成，则进行该请求，如果未完成，继续查询，直到上一级请求完成
				err, preEventStatus := model.QueryPlanStatus(machineIp + ":" + gid + ":" + sceneId + ":" + eventId + ":status")
				if err != nil {
					break
				}
				switch preEventStatus {
				case model.End:
					log.Logger.Debug("event.id:                   ", event.Id)
					delete(preMap, eventId)
				case model.NotRun:
					expiration := 60 * time.Second
					err = model.InsertStatus(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", model.NotRun, expiration)
					if err != nil {
						log.Logger.Error("事件状态写入数据库失败", err)
					}
					debugMsg := make(map[string]interface{})
					debugMsg["uuid"] = event.Uuid.String()
					debugMsg["event_id"] = event.Id
					debugMsg["status"] = model.NotRun
					debugMsg["type"] = event.Type
					debugMsg["report_id"] = event.ReportId
					debugMsg["next_list"] = event.NextList
					if requestCollection != nil {
						model.Insert(requestCollection, debugMsg)
					}
					delete(preMap, eventId)
					return
				case model.NotHit:
					expiration := 60 * time.Second
					err = model.InsertStatus(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", model.NotRun, expiration)
					if err != nil {
						log.Logger.Error("事件状态写入数据库失败", err)
					}
					debugMsg := make(map[string]interface{})
					debugMsg["uuid"] = event.Uuid.String()
					debugMsg["event_id"] = event.Id
					debugMsg["type"] = event.Type
					debugMsg["status"] = model.NotRun
					debugMsg["report_id"] = event.ReportId
					debugMsg["next_list"] = event.NextList
					if requestCollection != nil {
						model.Insert(requestCollection, debugMsg)
					}
					delete(preMap, eventId)
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
		DisposeRequest(nil, reportMsg, resultDataMsgCh, nil, scene.Configuration, event, requestCollection)
		expiration := 60 * time.Second
		err := model.InsertStatus(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
		if err != nil {
			log.Logger.Error("事件状态写入redis数据库失败", err)
		}
	case model.IfControllerType:
		keys := tools.FindAllDestStr(event.Var, "{{(.*?)}}")

		if len(keys) > 0 {
			for _, val := range keys {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						event.Var = strings.Replace(event.Var, val[0], kv.Value, -1)
					}
				}
			}
		}
		values := tools.FindAllDestStr(event.Val, "{{(.*?)}}")
		if len(values) > 0 {
			for _, val := range values {
				for _, kv := range scene.Configuration.Variable {
					if kv.Key == val[1] {
						event.Val = strings.Replace(event.Val, val[0], kv.Value, -1)
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
				result, msg = event.PerForm(kv.Value)
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

		expiration := 60 * time.Second
		if result == model.Failed {
			err := model.InsertStatus(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", model.NotHit, expiration)
			if err != nil {
				log.Logger.Error("事件状态写入redis数据库失败", err)
			}
		} else {
			err := model.InsertStatus(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
			if err != nil {
				log.Logger.Error("事件状态写入redis数据库失败", err)
			}
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
		expiration := 60 * time.Second
		err := model.InsertStatus(machineIp+":"+gid+":"+sceneId+":"+event.Id+":status", model.End, expiration)
		if err != nil {
			log.Logger.Error("事件状态写入数据库失败", err)
		}
	}

}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(wg *sync.WaitGroup, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestResults *model.ResultDataMsg, configuration *model.Configuration,
	event model.Event, mongoCollection *mongo.Collection, options ...int64) {
	if wg != nil {
		defer wg.Done()
	}
	api := event.Api
	if api.Debug == "" {
		api.Debug = event.Debug
	}
	// 计算接口权重，只要并发id大于并发数则直接返回
	if options != nil && len(options) > 1 {
		if options[0] > options[1] {
			return
		}
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
		timestamp     = int64(0)
	)
	switch api.TargetType {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, receivedBytes, errMsg, timestamp = HttpSend(event, api, configuration, mongoCollection)
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
		requestResults.ErrorMsg = errMsg
		requestResults.Timestamp = timestamp
		resultDataMsgCh <- requestResults
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
