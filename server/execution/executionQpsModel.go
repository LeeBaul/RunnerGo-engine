package execution

import (
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

// QPSModel 响应时间模式
func QPSModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection, sharedMap *sync.Map) string {
	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	timeUp := scene.ConfigTask.ModeConf.ReheatTime

	// 定义一个chan, 从es中获取当前错误率与阈值分别是多少
	// 连接es，并查询当前错误率为多少，并将其放入到chan中
	//err, esClient := client.NewEsClient(config.Config["esHost"].(string))
	//if err != nil {
	//	return
	//}
	//go GetRequestTime(esClient, requestTimeData)

	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	stepRunTime *= 1000
	step *= 1000
	stableDuration *= 1000
	concurrent := startConcurrent

	index, target := 0, 0

	key := fmt.Sprintf("%d:%s:reportData", reportMsg.PlanId, reportMsg.ReportId)
	// 创建es客户端

	currentWg := &sync.WaitGroup{}
	targetTime, startTime, endTime := time.Now().Unix(), time.Now().Unix(), time.Now().Unix()

	qpsTag := false
	// 只要开始时间+持续时长大于当前时间就继续循环
	for startTime+stepRunTime > endTime {
		_, status := model.QueryPlanStatus(fmt.Sprintf("%d:%d:%s:status", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
		if status == "stop" {
			model.DeleteKey(fmt.Sprintf("%d:%d:%s:status", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
			return fmt.Sprintf("测试报告：%s, 最大并发数：%d， 总运行时长%ds, 任务正常结束！", reportMsg.ReportId, concurrent, endTime-targetTime)
		}
		reportId, _ := strconv.Atoi(reportMsg.ReportId)
		debug := model.QueryDebugStatus(debugCollection, reportId)
		if debug != "" {
			scene.Debug = debug
		}

		// 如果所有的接口qps都达到阈值，则不在进行查询当前qps
		if !qpsTag {
			res := model.QueryReportData(key)
			if res != "" {
				var result = new(model.RedisSceneTestResultDataMsg)
				err := json.Unmarshal([]byte(res), result)
				if err != nil {
					break
				}

				tag, apiLen := 0, len(result.Results)
				for _, resultData := range result.Results {
					if resultData.TotalRequestNum != 0 {
						if resultData.ResponseThreshold == 0 {
							tag++
						}
						if resultData.Qps >= float64(resultData.RequestThreshold) && resultData.ResponseThreshold != 0 {
							tag++
						}
						if resultData.Qps < float64(resultData.RequestThreshold) && resultData.ResponseThreshold != 0 {
							tag--
						}

					}

				}
				// 如果所有的接口qps都达到阈值，那么直接进入最大并发数
				if tag == apiLen {
					concurrent = maxConcurrent
					stepRunTime = stableDuration
					qpsTag = true
				}
			}

		}

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
			go func(i, concurrent int64) {
				gid := tools.GetGid()
				golink.DisposeScene(sharedMap, wg, currentWg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
				wg.Done()
				currentWg.Done()
			}(i, concurrent)
			// 如果设置了启动并发时长
			if index == 0 && timeUp != 0 && i%(startConcurrent/timeUp) == 0 && i != 0 {
				distance := time.Now().UnixMilli() - startConcurrentTime
				if distance < 1000 {
					time.Sleep(time.Duration(distance) * time.Millisecond)
				}

			}
		}

		currentWg.Wait()
		// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
		endTime = time.Now().Unix()
		if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
			return fmt.Sprintf("测试报告：%s, 最大并发数：%d， 总运行时长%ds, 任务正常结束！", reportMsg.ReportId, concurrent, endTime-targetTime)
		}

		// 如果当前并发数小于最大并发数，
		if concurrent < maxConcurrent {
			if endTime-startTime >= stepRunTime {
				// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
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
