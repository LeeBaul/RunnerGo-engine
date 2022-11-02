package execution

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"kp-runner/tools"
	"math"
	"strconv"
	"sync"
	"time"
)

// RTModel 响应时间模式
func RTModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection, sharedMap *sync.Map) {
	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	timeUp := scene.ConfigTask.ModeConf.ReheatTime

	planId := strconv.FormatInt(reportMsg.PlanId, 10)
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

	//es := model.NewEsClient(config.Conf.Es.Host, config.Conf.Es.UserName, config.Conf.Es.Password)
	//if es == nil {
	//	return
	//}
	key := fmt.Sprintf("%d:%s:reportData", reportMsg.PlanId, reportMsg.ReportId)
	currentWg := &sync.WaitGroup{}
	// 只要开始时间+持续时长大于当前时间就继续循环
	startTime := time.Now().Unix()
	for startTime+stepRunTime > time.Now().Unix() {

		_, status := model.QueryPlanStatus(reportMsg.ReportId + ":status")
		if status == "stop" {
			return
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
					times := int64(math.Ceil(resultData.FiftyRequestTimelineValue))
					if resultData.RequestThreshold > 0 && times >= resultData.RequestThreshold {
						log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ":  ", resultData.PercentAge, "%响应时间线大于等于阈值：  ", resultData.RequestThreshold, "   任务结束结束")
						return
					}
				case 90:
					times := int64(math.Ceil(resultData.NinetyRequestTimeLineValue))
					if resultData.RequestThreshold > 0 && times >= resultData.RequestThreshold {
						log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ":  ", resultData.PercentAge, "%响应时间线大于等于阈值：  ", resultData.RequestThreshold, "   任务结束结束")
						return
					}
				case 95:
					times := int64(math.Ceil(resultData.NinetyFiveRequestTimeLineValue))
					if resultData.RequestThreshold > 0 && times >= resultData.RequestThreshold {
						log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ":  ", resultData.PercentAge, "%响应时间线大于等于阈值：  ", resultData.RequestThreshold, "   任务结束结束")
						return
					}
				case 99:
					times := int64(math.Ceil(resultData.NinetyNineRequestTimeLineValue))
					if resultData.RequestThreshold > 0 && times >= resultData.RequestThreshold {
						log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ":  ", resultData.PercentAge, "%响应时间线大于等于阈值：  ", resultData.RequestThreshold, "   任务结束结束")
						return
					}
				case 100:
					times := int64(math.Ceil(resultData.MaxRequestTime))
					if resultData.RequestThreshold > 0 && times >= resultData.RequestThreshold {
						log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ":  ", resultData.PercentAge, "%响应时间线大于等于阈值：  ", resultData.RequestThreshold, "   任务结束结束")
						return
					}
				case 101:
					times := int64(math.Ceil(resultData.AvgRequestTime))
					if resultData.RequestThreshold > 0 && times >= resultData.RequestThreshold {
						log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ":  ", "平均响应时间线大于等于阈值：  ", resultData.RequestThreshold, "   任务结束结束")
						return
					}
				default:
					if resultData.PercentAge == resultData.CustomRequestTimeLine {
						times := int64(math.Ceil(resultData.CustomRequestTimeLineValue))
						if resultData.RequestThreshold > 0 && times >= resultData.RequestThreshold {
							log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ":  ", resultData.PercentAge, "%响应时间线大于等于阈值：  ", resultData.RequestThreshold, "   任务结束结束")
							return
						}
					}

				}
			}
		}
		startConcurrentTime := time.Now().Unix()
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
		endTime := time.Now().Unix()
		// 当此时的并发等于最大并发数时，并且持续时长等于稳定持续时长且当前运行时长大于等于此时时结束
		if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime <= time.Now().Unix() {
			log.Logger.Info("计划: ", planId, "；报告：   ", reportId, "     :结束 ")
		}

		// 如果当前并发数小于最大并发数，
		if concurrent < maxConcurrent {
			if endTime-startTime >= stepRunTime {
				// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
				concurrent = concurrent + step
				if concurrent > maxConcurrent {
					concurrent = maxConcurrent
				}

				if startTime+stepRunTime <= time.Now().Unix() && concurrent < maxConcurrent {
					startTime = startTime + stepRunTime

				}
			}

		}
		if concurrent == maxConcurrent {
			if target == 0 {
				stepRunTime = stableDuration
				startTime = startTime + stepRunTime
			}
			target++
		}
		index++

	}

}
