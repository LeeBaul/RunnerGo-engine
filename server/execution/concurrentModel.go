package execution

import (
	"fmt"
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
func ConcurrentModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection) {

	startTime := time.Now().UnixMilli()
	concurrent := scene.ConfigTask.ModeConf.Concurrency
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime

	if scene.ConfigTask.ModeConf.Duration != 0 {
		log.Logger.Info("开始性能测试,持续时间", scene.ConfigTask.ModeConf.Duration, "秒")
		index := 0
		duration := scene.ConfigTask.ModeConf.Duration * 1000
		currentTime := time.Now().UnixMilli()

		// 并发数的所有请求都完成后进行下一轮并发
		currentWg := &sync.WaitGroup{}
		for startTime+duration > currentTime {
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
			startCurrentTime := time.Now().UnixMilli()
			for i := int64(0); i < concurrent; i++ {
				wg.Add(1)
				currentWg.Add(1)
				go func(i, concurrent int64) {
					gid := tools.GetGid()
					golink.DisposeScene(wg, currentWg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
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
			// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
			distance := time.Now().UnixMilli() - startCurrentTime
			if distance < 1000 {
				sleepTime := time.Duration(1000-distance) * time.Millisecond
				time.Sleep(sleepTime)
			}
			currentTime = time.Now().UnixMilli()
		}

	} else {
		index := 0
		rounds := scene.ConfigTask.ModeConf.RoundNum
		currentWg := &sync.WaitGroup{}
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
					golink.DisposeScene(wg, currentWg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
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
			fmt.Println(fmt.Sprintf("第%d次并发", i))
			index++

			distance := time.Now().UnixMilli() - currentTime
			if distance < 1000 {
				sleepTime := time.Duration(1000-distance) * time.Millisecond
				time.Sleep(sleepTime)
			}

		}

	}

}
