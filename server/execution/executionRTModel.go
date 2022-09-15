package execution

import (
	"github.com/olivere/elastic"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"strconv"
	"sync"
	"time"
)

type RequestTimeData struct {
	PlanId  string           `json:"planId"`
	SceneId string           `json:"sceneId"`
	Apis    []RequestTimeApi `json:"apis"`
}

type RequestTimeApi struct {
	ApiName   string  `json:"apiName"`
	Threshold float64 `json:"threshold"`
	Actual    float64 `json:"actual"`
	Custom    int     `json:"custom"`
}

func GetRequestTime(esClient *elastic.Client, requestTimeData *RequestTimeData) {

}

// RTModel 响应时间模式
func RTModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {

	startConcurrent := scene.ConfigTask.ModeConf.StartConcurrency
	step := scene.ConfigTask.ModeConf.Step
	maxConcurrent := scene.ConfigTask.ModeConf.MaxConcurrency
	stepRunTime := scene.ConfigTask.ModeConf.StepRunTime
	stableDuration := scene.ConfigTask.ModeConf.Duration
	timeUp := scene.ConfigTask.ModeConf.ReheatTime

	planId := strconv.FormatInt(reportMsg.PlanId, 10)
	// 定义一个chan, 从es中获取当前错误率与阈值分别是多少
	requestTimeData := new(RequestTimeData)
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

	startTime := time.Now().UnixMilli()
	startCurrentTime := startTime
	currentTime := startTime
	index := 0
	// 只要开始时间+持续时长大于当前时间就继续循环
	for startTime+stepRunTime > currentTime {

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
		startConcurrentTime := time.Now().UnixMilli()
		for i := int64(0); i < concurrent; i++ {
			wg.Add(1)
			go func(i, concurrent int64) {
				gid := GetGid()
				golink.DisposeScene(wg, gid, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
				wg.Done()
			}(i, concurrent)
			// 如果设置了启动并发时长
			if index == 0 && timeUp != 0 && i%(startConcurrent/timeUp) == 0 && i != 0 {
				distance := time.Now().UnixMilli() - startConcurrentTime
				if distance < 1000 {
					time.Sleep(time.Duration(distance) * time.Millisecond)
				}

			}
		}
		// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
		distance := time.Now().UnixMilli() - startCurrentTime
		if distance < 1000 {
			sleepTime := time.Duration(1000-distance) * time.Millisecond
			time.Sleep(sleepTime)
		}
		currentTime = time.Now().UnixMilli()

		// 当此时的并发等于最大并发数时，并且持续时长等于稳定持续时长且当前运行时长大于等于此时时结束
		if concurrent == maxConcurrent && stepRunTime == stableDuration && startTime+stepRunTime >= time.Now().UnixMilli() {
			log.Logger.Info("计划: ", planId, "..............结束")
			return
		}

		// 如果当前并发数小于最大并发数，
		if concurrent <= maxConcurrent {
			// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
			if startTime+stepRunTime >= time.Now().UnixMilli() {
				startTime = time.Now().UnixMilli()
				//preConcurrent = concurrent
				if concurrent+step <= maxConcurrent {
					concurrent = concurrent + step
				} else {
					concurrent = maxConcurrent
					stepRunTime = stableDuration
				}
				apis := requestTimeData.Apis
				for _, api := range apis {
					if api.Threshold < api.Actual {
						log.Logger.Info(api.ApiName, "接口：在", concurrent, "并发时", api.Custom, "线响应时间", api.Actual, "大于所设阈值", api.Threshold)
						log.Logger.Info("计划:", planId, "...............结束")
						return

					}
				}

			}
		}
		index++

	}

}
