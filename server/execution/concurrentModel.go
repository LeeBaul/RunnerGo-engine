package execution

import (
	"RunnerGo-engine/log"
	"RunnerGo-engine/model"
	"RunnerGo-engine/server/golink"
	"RunnerGo-engine/server/heartbeat"
	"RunnerGo-engine/tools"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"sync"
	"time"
)

// ConcurrentModel 并发模式
func ConcurrentModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection, sharedMap *sync.Map) string {

	concurrent := scene.ConfigTask.ModeConf.Concurrency
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime

	if scene.ConfigTask.ModeConf.Duration != 0 {
		log.Logger.Info("开始性能测试,持续时间", scene.ConfigTask.ModeConf.Duration, "秒")
		index := 0
		duration := scene.ConfigTask.ModeConf.Duration

		// 并发数的所有请求都完成后进行下一轮并发
		currentWg := &sync.WaitGroup{}
		targetTime, startTime := time.Now().Unix(), time.Now().Unix()
		for startTime+duration >= time.Now().Unix() {
			// 查询是否停止
			_, status := model.QueryPlanStatus(fmt.Sprintf("%d:%d:", reportMsg.TeamId, reportMsg.PlanId) + reportMsg.ReportId + ":status")
			if status == "stop" {
				return fmt.Sprintf("测试报告：%s, 并发数：%d, 总运行时长%ds, 任务手动结束！", reportMsg.ReportId, concurrent, time.Now().Unix()-targetTime)
			}
			// 查询是否开启debug
			reportId, _ := strconv.Atoi(reportMsg.ReportId)
			debug := model.QueryDebugStatus(debugCollection, reportId)
			if debug != "" {
				scene.Debug = debug
			}

			// 查询是否在报告中对任务模式进行修改
			err, mode := model.QueryPlanStatus(fmt.Sprintf("adjust:%s:%d:%d:%s", heartbeat.LocalIp, reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
			if err == nil {
				var modeConf = new(model.ModeConf)
				_ = json.Unmarshal([]byte(mode), modeConf)
				if modeConf.Duration > 0 {
					startTime = time.Now().Unix()
					duration = modeConf.Duration
				}
				if modeConf.Concurrency > 0 {
					concurrent = modeConf.Concurrency
				}
				model.DeleteKey(fmt.Sprintf("adjust:%s:%d:%d:%s", heartbeat.LocalIp, reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
			}

			startConcurrentTime := time.Now().UnixMilli()
			log.Logger.Debug("并发数：", concurrent)
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
		return fmt.Sprintf("测试报告：%s, 并发数：%d, 总运行时长%ds, 任务正常结束！", reportMsg.ReportId, concurrent, time.Now().Unix()-targetTime)

	} else {
		index := 0
		rounds := scene.ConfigTask.ModeConf.RoundNum
		currentWg := &sync.WaitGroup{}
		startTime := time.Now().UnixMilli()
		for i := int64(0); i < rounds; i++ {
			_, status := model.QueryPlanStatus(reportMsg.ReportId + ":status")
			if status == "stop" {
				return fmt.Sprintf("测试报告：%s, 并发数：%d， 运行了%d轮次, 任务手动结束！", reportMsg.ReportId, concurrent, i-1)
			}
			reportId, _ := strconv.Atoi(reportMsg.ReportId)
			debug := model.QueryDebugStatus(debugCollection, reportId)
			if debug != "" {
				scene.Debug = debug
			} else {
				scene.Debug = ""
			}

			// 查询是否在报告中对任务模式进行修改
			err, mode := model.QueryPlanStatus(fmt.Sprintf("adjust:%s:%d:%d:%s", reportMsg.MachineIp, reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
			if err == nil {
				var modeConf = new(model.ModeConf)
				_ = json.Unmarshal([]byte(mode), modeConf)
				if modeConf.RoundNum > 0 {
					rounds = modeConf.RoundNum
				}
				if modeConf.Concurrency > 0 {
					concurrent = modeConf.Concurrency
				}
				model.DeleteKey(fmt.Sprintf("adjust:%s:%d:%d:%s", reportMsg.MachineIp, reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
			}

			currentTime := time.Now().UnixMilli()
			log.Logger.Debug("并发数：", concurrent)
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
		return fmt.Sprintf("测试报告：%s, 并发数：%d， 运行了%d轮次, 任务正常结束！", reportMsg.ReportId, concurrent, rounds)
	}

}
