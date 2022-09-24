package execution

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
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
func ErrorRateModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {

	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime

	planId := strconv.FormatInt(reportMsg.PlanId, 10)
	// 定义一个chan, 从es中获取当前错误率与阈值分别是多少
	startTime := time.Now().Unix()
	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	concurrent := startConcurrent
	// 只要开始时间+持续时长大于当前时间就继续循环
	index := 0
	// 创建es客户端，获取测试数据
	es := model.NewEsClient(config.Conf.Es.Host, config.Conf.Es.UserName, config.Conf.Es.Password)
	if es == nil {
		return
	}
	for startTime+stepRunTime > time.Now().Unix() {
		// 查询任务是否结束
		_, status := model.QueryPlanStatus(reportMsg.ReportId + ":status")
		if status == "stop" {
			return
		}
		_, debug := model.QueryPlanStatus(reportMsg.ReportId + ":debug")
		if debug != "" {
			scene.Debug = debug
		} else {
			scene.Debug = ""
		}

		// 查询当前错误率时多少
		//GetErrorRate(planId+":"+sceneId+":"+"errorRate", errorRateData)
		res := model.QueryReport(es, config.Conf.Es.Index, reportMsg.ReportId)
		if res != nil && res.Results != nil {
			for _, result := range res.Results {
				errRate := float64(result.ErrorNum) / float64(result.TotalRequestNum)
				if errRate > result.ErrorThreshold {
					log.Logger.Info(result.Name, "接口：在", concurrent, "并发时,错误率", errRate, "大于所设阈值", result.ErrorThreshold)
					log.Logger.Info("计划:", planId, "...............结束")
					return
				}
			}
		}

		for i := int64(0); i < concurrent; i++ {
			wg.Add(1)
			go func(i, concurrent int64) {
				gid := GetGid()
				golink.DisposeScene(wg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
				wg.Done()
			}(i, concurrent)
			// 如果设置了启动并发时长
			if reheatTime > 0 && index == 0 {
				durationTime := time.Now().UnixMilli() - startTime
				if i%(concurrent/reheatTime) == 0 && durationTime < 1000 {
					time.Sleep(time.Duration(durationTime) * time.Millisecond)
				}
			}
		}
		index++

		if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime >= time.Now().Unix() {
			log.Logger.Info("计划:", planId, ".....................结束")
			return
		}
		// 如果当前并发数小于最大并发数，
		if concurrent <= maxConcurrent {
			// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
			if startTime+stepRunTime >= time.Now().Unix() {
				startTime = time.Now().Unix()
				//preConcurrent = concurrent
				if concurrent+step <= maxConcurrent {
					concurrent = concurrent + step
				} else {
					concurrent = maxConcurrent
					stepRunTime = stableDuration
				}

			}
		}
	}
}
