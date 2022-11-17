package execution

import (
	"RunnerGo-engine/model"
	"RunnerGo-engine/server/golink"
	"RunnerGo-engine/server/heartbeat"
	"RunnerGo-engine/tools"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"math"
	"strconv"
	"sync"
	"time"
)

// RTModel 响应时间模式
func RTModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection, sharedMap *sync.Map) (msg string) {
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
	concurrent := startConcurrent

	index, target := 0, 0

	//es := model.NewEsClient(config.Conf.Es.Host, config.Conf.Es.UserName, config.Conf.Es.Password)
	//if es == nil {
	//	return
	//}
	key := fmt.Sprintf("%d:%s:reportData", reportMsg.PlanId, reportMsg.ReportId)
	currentWg := &sync.WaitGroup{}
	// 只要开始时间+持续时长大于当前时间就继续循环
	targetTime, startTime, endTime := time.Now().Unix(), time.Now().Unix(), time.Now().Unix()
	for startTime+stepRunTime > endTime {

		_, status := model.QueryPlanStatus(fmt.Sprintf("StopPlan:%d:%d:%s", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
		if status == "stop" {
			model.DeleteKey(fmt.Sprintf("StopPlan:%d:%d:%s", reportMsg.TeamId, reportMsg.PlanId, reportMsg.ReportId))
			return fmt.Sprintf("测试报告：%s, 最大并发数：%d， 总运行时长%ds, 任务手动结束！", reportMsg.ReportId, concurrent, endTime-targetTime)
		}
		reportId, _ := strconv.Atoi(reportMsg.ReportId)
		debug := model.QueryDebugStatus(debugCollection, reportId)
		if debug != "" {
			scene.Debug = debug
		}

		res := model.QueryReportData(key)
		if res != "" {
			var result = new(model.RedisSceneTestResultDataMsg)
			err := json.Unmarshal([]byte(res), result)
			if err != nil {
				break
			}
			for _, resultData := range result.Results {
				switch resultData.PercentAge {
				case 50:
					times := int64(math.Ceil(resultData.FiftyRequestTimelineValue / float64(time.Millisecond)))
					if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
						return fmt.Sprintf("测试报告：%d, 最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", reportId, concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
					}
				case 90:
					times := int64(math.Ceil(resultData.NinetyRequestTimeLineValue / float64(time.Millisecond)))
					if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
						return fmt.Sprintf("测试报告：%d, 最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", reportId, concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
					}
				case 95:
					times := int64(math.Ceil(resultData.NinetyFiveRequestTimeLineValue / float64(time.Millisecond)))
					if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {

						return fmt.Sprintf("测试报告：%d, 最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", reportId, concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
					}
				case 99:
					times := int64(math.Ceil(resultData.NinetyNineRequestTimeLineValue / float64(time.Millisecond)))
					if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
						return fmt.Sprintf("测试报告：%d, 最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", reportId, concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
					}
				case 100:
					times := int64(math.Ceil(resultData.MaxRequestTime / float64(time.Millisecond)))
					if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
						return fmt.Sprintf("测试报告：%d, 最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", reportId, concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
					}
				case 101:
					times := int64(math.Ceil(resultData.AvgRequestTime / float64(time.Millisecond)))
					if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
						return fmt.Sprintf("测试报告：%d, 最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", reportId, concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
					}
				default:
					if resultData.PercentAge == resultData.CustomRequestTimeLine {
						times := int64(math.Ceil(resultData.CustomRequestTimeLineValue / float64(time.Millisecond)))
						if resultData.ResponseThreshold > 0 && times >= resultData.ResponseThreshold {
							return fmt.Sprintf("测试报告：%d, 最大并发数：%d， 总运行时长%ds,  接口：%s: %d 响应时间线大于等于阈值%d, 任务结束！", reportId, concurrent, endTime-targetTime, resultData.Name, resultData.PercentAge, resultData.RequestThreshold)
						}
					}

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
		endTime = time.Now().Unix()
		// 当此时的并发等于最大并发数时，并且持续时长等于稳定持续时长且当前运行时长大于等于此时时结束
		if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= endTime {
			return fmt.Sprintf("测试报告：%s, 到达最大并发数：%d, 总运行时长%d秒, 任务正常结束！", reportMsg.ReportId, maxConcurrent, endTime-targetTime)
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
	return fmt.Sprintf("测试报告：%s, 到达最大并发数：%d, 总运行时长%d秒, 任务非正常结束！", reportMsg.ReportId, maxConcurrent, endTime-targetTime)

}
