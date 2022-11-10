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

type ErrorRateData struct {
	PlanId  string `json:"planId"`
	SceneId string `json:"sceneId"`
	Apis    []Apis `json:"apis"`
}

type Apis struct {
	ApiName   string  `json:"apiName"`
	Threshold float64 `json:"threshold"`
	Actual    float64 `json:"actual"`
}

// ErrorRateModel 错误率模式
func ErrorRateModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, requestCollection *mongo.Collection, sharedMap *sync.Map) {
	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime

	planId := strconv.FormatInt(reportMsg.PlanId, 10)
	// 定义一个chan, 从es中获取当前错误率与阈值分别是多少

	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	concurrent := startConcurrent
	// 只要开始时间+持续时长大于当前时间就继续循环
	index, target := 0, 0
	// 创建es客户端，获取测试数据
	//es := model.NewEsClient(config.Conf.Es.Host, config.Conf.Es.UserName, config.Conf.Es.Password)
	//if es == nil {
	//	return
	//}
	key := fmt.Sprintf("%d:%s:reportData", reportMsg.PlanId, reportMsg.ReportId)
	currentWg := &sync.WaitGroup{}
	startTime, endTime := time.Now().Unix(), time.Now().Unix()
	for startTime+stepRunTime > endTime {
		// 查询任务是否结束
		_, status := model.QueryPlanStatus(reportMsg.ReportId + ":status")
		if status == "stop" {
			return
		}
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

		// 查询当前错误率时多少
		//GetErrorRate(planId+":"+sceneId+":"+"errorRate", errorRateData)
		res := model.QueryReportData(key)
		if res != "" {
			var result = new(model.RedisSceneTestResultDataMsg)
			err := json.Unmarshal([]byte(res), result)
			if err != nil {
				break
			}
			for _, resultData := range result.Results {
				if resultData.TotalRequestNum != 0 {
					if resultData.ErrorRate > resultData.ErrorThreshold {
						log.Logger.Info("计划:", planId, "——测试报告：", result.ReportId, "  接口：", resultData.Name, ": 错误率为： ", resultData.ErrorRate, "大于等于阈值：  ", resultData.ErrorThreshold, "   任务结束结束")
						return
					}
				}

			}
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
}
