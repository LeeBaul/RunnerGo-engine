package execution

import (
	"github.com/olivere/elastic"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/client"
	"kp-runner/server/golink"
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
func GetErrorRate(esClient *elastic.Client, errorRateData *ErrorRateData) {
	for {
		esClient.Get()

		time.Sleep(1 * time.Second)
	}
}

// ExecutionErrorRateModel 错误率模式
func ExecutionErrorRateModel(statusCh chan bool, plan *model.Plan, ch chan *model.ResultDataMsg) {
	defer close(ch)
	// 定义一个chan, 从es中获取当前错误率与阈值分别是多少
	errorRateData := new(ErrorRateData)
	// 连接es，并查询当前错误率为多少，并将其放入到chan中
	err, esClient := client.NewEsClient(config.Config["esHost"].(string))
	if err != nil {
		return
	}
	go GetErrorRate(esClient, errorRateData)
	startTime := time.Now().Unix()
	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	startConcurrent := plan.ConfigTask.TestModel.ErrorRatTest.StartConcurrent
	concurrent := startConcurrent
	length := plan.ConfigTask.TestModel.ErrorRatTest.Length
	maxConcurrent := plan.ConfigTask.TestModel.ErrorRatTest.MaxConcurrent
	lengthDuration := plan.ConfigTask.TestModel.ErrorRatTest.LengthDuration
	stableDuration := plan.ConfigTask.TestModel.ErrorRatTest.StableDuration
	timeUp := plan.ConfigTask.TestModel.ErrorRatTest.TimeUp
	eventList := plan.Scene.EventList
	// 只要开始时间+持续时长大于当前时间就继续循环
	for startTime+lengthDuration > time.Now().Unix() {
		select {
		case status := <-statusCh:
			if status == false {
				log.Logger.Info("计划", plan.PlanName, "结束")
				return
			}
		default:
			var currenWg = &sync.WaitGroup{}
			for i := int64(0); i < concurrent; i++ {
				currenWg.Add(1)
				go func(i, concurrent int64) {
					if plan.Variable.VariableMap == nil {
						plan.Variable.VariableMap = new(sync.Map)
					}
					globalVariable := plan.Variable.VariableMap
					golink.Dispose(i, concurrent, eventList, ch, plan, globalVariable)
					currenWg.Done()
				}(i, concurrent)
				currenWg.Done()
				// 如果设置了启动并发时长
				if timeUp != 0 && (startConcurrent/timeUp)%i == 0 && i != 0 {
					time.Sleep(1 * time.Second)
				}
			}
			currenWg.Wait()

			if concurrent == maxConcurrent && lengthDuration == stableDuration && startTime+lengthDuration >= time.Now().Unix() {
				log.Logger.Info("计划", plan.PlanName, "结束")
				return
			}
			// 如果当前并发数小于最大并发数，
			if concurrent <= maxConcurrent {
				// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
				if startTime+int64(lengthDuration) >= time.Now().Unix() {
					startTime = time.Now().Unix()
					//preConcurrent = concurrent
					if concurrent+length <= maxConcurrent {
						concurrent = concurrent + length
					} else {
						concurrent = maxConcurrent
						lengthDuration = stableDuration
					}
					apis := errorRateData.Apis
					for _, api := range apis {
						if api.Threshold < api.Actual {
							log.Logger.Info(api.ApiName, "接口：在", concurrent, "并发时,错误率", api.Actual, "大于所设阈值", api.Threshold)
							log.Logger.Info("计划", plan.PlanName, "结束")
							return
						}
					}

				}
			}
		}

	}
}
