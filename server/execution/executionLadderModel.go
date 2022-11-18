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

// LadderModel 阶梯模式
func LadderModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection, sharedMap *sync.Map) string {
	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime
	// 连接es，并查询当前错误率为多少，并将其放入到chan中

	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	concurrent := startConcurrent
	index, target := 0, 0
	currentWg := &sync.WaitGroup{}
	targetTime, startTime, endTime := time.Now().Unix(), time.Now().Unix(), time.Now().Unix()

	// 只要开始时间+持续时长大于当前时间就继续循环
	for startTime+stepRunTime > endTime {
		// 查询任务是否结束
		_, status := model.QueryPlanStatus(fmt.Sprintf("StopPlan:%d:%d:%s", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
		if status == "stop" {
			err := model.DeleteKey(fmt.Sprintf("StopPlan:%d:%d:%s", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
			if err != nil {
				log.Logger.Error("停止性能任务key: ", fmt.Sprintf("StopPlan:%d:%d:%s", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId), " ; 删除失败")
			} else {
				log.Logger.Info("停止性能任务key: ", fmt.Sprintf("StopPlan:%d:%d:%s", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId), " ; 删除成功")
			}
			return fmt.Sprintf("测试报告：%s, 最大并发数：%d， 总运行时长%ds, 任务手动结束！", reportMsg.ReportId, concurrent, endTime-targetTime)
		}
		reportId, _ := strconv.Atoi(reportMsg.ReportId)
		debug := model.QueryDebugStatus(debugCollection, reportId)
		scene.Debug = debug
		// 查询是否在报告中对任务模式进行修改
		err, mode := model.QueryPlanStatus(fmt.Sprintf("adjust:%s:%d:%d:%s", heartbeat.LocalIp, reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
		if err == nil {
			var modeConf = new(model.ModeConf)
			_ = json.Unmarshal([]byte(mode), modeConf)
			if modeConf.StartConcurrency > 0 {
				concurrent = modeConf.StartConcurrency
			}
			if modeConf.StepRunTime > 0 {
				stepRunTime = modeConf.StepRunTime
				startTime = time.Now().Unix()
			}
			if modeConf.MaxConcurrency > 0 {
				maxConcurrent = modeConf.MaxConcurrency
			}
			if modeConf.Duration > 0 {
				stableDuration = modeConf.Duration
			}
			// 删除redis中的key
			model.DeleteKey(fmt.Sprintf("adjust:%s:%d:%d:%s", heartbeat.LocalIp, reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
		}

		startConcurrentTime := time.Now().UnixMilli()
		for i := int64(0); i < concurrent; i++ {
			wg.Add(1)
			currentWg.Add(1)
			go func(i, concurrent int64, wg *sync.WaitGroup) {
				gid := tools.GetGid()
				golink.DisposeScene(sharedMap, wg, currentWg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
				wg.Done()
				currentWg.Done()
			}(i, concurrent, wg)
			// 如果设置了启动并发时长
			if reheatTime > 0 && index == 0 {
				durationTime := time.Now().UnixMilli() - startConcurrentTime
				if i%(concurrent/reheatTime) == 0 && durationTime < 1000 {
					time.Sleep(time.Duration(durationTime) * time.Millisecond)
				}
			}
		}
		currentWg.Wait()
		endTime = time.Now().Unix()
		if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
			return fmt.Sprintf("测试报告：%s, 最大并发数：%d， 总运行时长%ds, 任务正常结束！", reportMsg.ReportId, concurrent, endTime-targetTime)
		}

		// 如果当前并发数小于最大并发数，
		if concurrent < maxConcurrent {
			if endTime-startTime >= stepRunTime {
				// 从开始时间算起，加上持续时长。如果大于现在的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
				concurrent = concurrent + step
				if concurrent > maxConcurrent {
					concurrent = maxConcurrent
				}
				if startTime+stepRunTime <= endTime && concurrent < maxConcurrent {
					startTime = endTime + stepRunTime
				}
			}
		}
		if concurrent == maxConcurrent {
			if target == 0 {
				stepRunTime = stableDuration
				startTime = endTime + stepRunTime
			}
			target++
		}
		index++
	}
	return fmt.Sprintf("测试报告：%s, 最大并发数：%d， 总运行时长%ds, 任务非正常结束！", reportMsg.ReportId, concurrent, time.Now().Unix()-targetTime)

}
