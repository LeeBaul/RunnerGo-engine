package execution

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"kp-runner/tools"
	"strconv"
	"sync"
	"time"
)

// ConcurrentModel 并发模式
func ConcurrentModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection, sharedMap *sync.Map) {

	concurrent := scene.ConfigTask.ModeConf.Concurrency
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime

	if scene.ConfigTask.ModeConf.Duration != 0 {
		log.Logger.Info("开始性能测试,持续时间", scene.ConfigTask.ModeConf.Duration, "秒")
		index := 0
		duration := scene.ConfigTask.ModeConf.Duration * 1000

		// 并发数的所有请求都完成后进行下一轮并发
		currentWg := &sync.WaitGroup{}
		startTime := time.Now().Unix()
		for startTime+duration >= time.Now().Unix() {
			// 查询是否停止
			_, status := model.QueryPlanStatus(reportMsg.ReportId + ":status")
			if status == "stop" {
				return
			}
			// 查询是否开启debug
			reportId, _ := strconv.Atoi(reportMsg.ReportId)
			debug := model.QueryDebugStatus(debugCollection, reportId)
			if debug != "" {
				scene.Debug = debug
			}
			startConcurrentTime := time.Now().UnixMilli()
			for i := int64(0); i < concurrent; i++ {
				wg.Add(1)
				currentWg.Add(1)
				go func(i, concurrent int64, planSharedMap *sync.Map) {
					gid := tools.GetGid()
					golink.DisposeScene(planSharedMap, wg, currentWg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
					wg.Done()
					currentWg.Done()

				}(i, concurrent, sharedMap)
				if reheatTime > 0 && index == 0 {
					durationTime := time.Now().UnixMilli() - startConcurrentTime
					if (concurrent/reheatTime) > 0 && i%(concurrent/reheatTime) == 0 && durationTime < 1000 {
						time.Sleep(time.Duration(durationTime) * time.Millisecond)
					}
				}
			}

			currentWg.Wait()
			index++
		}

	} else {
		index := 0
		rounds := scene.ConfigTask.ModeConf.RoundNum
		currentWg := &sync.WaitGroup{}
		startTime := time.Now().UnixMilli()
		for i := int64(0); i < rounds; i++ {
			_, status := model.QueryPlanStatus(reportMsg.ReportId + ":status")
			if status == "stop" {
				return
			}
			reportId, _ := strconv.Atoi(reportMsg.ReportId)
			debug := model.QueryDebugStatus(debugCollection, reportId)
			if debug != "" {
				scene.Debug = debug
			} else {
				scene.Debug = ""
			}
			currentTime := time.Now().UnixMilli()
			for j := int64(0); j < concurrent; j++ {
				wg.Add(1)
				currentWg.Add(1)
				go func(i, concurrent int64) {
					gid := tools.GetGid()
					golink.DisposeScene(sharedMap, wg, currentWg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent, currentTime)
					wg.Done()
					currentWg.Done()
				}(i, concurrent)
				if reheatTime > 0 && index == 0 {
					durationTime := time.Now().UnixMilli() - startTime
					if i%(concurrent/reheatTime) == 0 && durationTime < 1000 {
						time.Sleep(time.Duration(durationTime) * time.Millisecond)
					}
				}
			}
			currentWg.Wait()
			index++

			distance := time.Now().UnixMilli() - currentTime
			if distance < 1000 {
				sleepTime := time.Duration(1000-distance) * time.Millisecond
				time.Sleep(sleepTime)
			}

		}

	}

}
