package execution

import (
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
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

// GetErrorRate 查询es，当前错误率
func GetErrorRate(key string, errorRateData *ErrorRateData) {
	value, err := model.RDB.Get(key).Result()
	if err != nil {
		return
	}
	_ = json.Unmarshal([]byte(value), errorRateData)

}

// ErrorRateModel 错误率模式
func ErrorRateModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {
	defer close(resultDataMsgCh)

	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime

	planId := strconv.FormatInt(reportMsg.PlanId, 10)
	sceneId := reportMsg.SceneId
	// 定义一个chan, 从es中获取当前错误率与阈值分别是多少
	errorRateData := new(ErrorRateData)
	startTime := time.Now().Unix()
	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	concurrent := startConcurrent
	// 只要开始时间+持续时长大于当前时间就继续循环
	index := 0
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
		GetErrorRate(planId+":"+sceneId+":"+"errorRate", errorRateData)
		apis := errorRateData.Apis
		for _, api := range apis {
			if api.Threshold < api.Actual {
				log.Logger.Info(api.ApiName, "接口：在", concurrent, "并发时,错误率", api.Actual, "大于所设阈值", api.Threshold)
				log.Logger.Info("计划:", planId, "...............结束")
				return
			}
		}
		for i := int64(0); i < concurrent; i++ {
			wg.Add(1)
			go func(i, concurrent int64) {
				gid := GetGid()
				golink.DisposeScene(wg, gid, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
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
