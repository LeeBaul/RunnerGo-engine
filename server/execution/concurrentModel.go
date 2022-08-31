package execution

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionConcurrentModel 并发模式
func ExecutionConcurrentModel(eventList []model.Event,
	ch chan *model.ResultDataMsg,
	concurrent, testType, roundOrTime int64,
	planId, planName, sceneId, sceneName, reportId, reportName string,
	configuration *model.Configuration,
	globalVariable *sync.Map,
	wg *sync.WaitGroup,
	requestCollection *mongo.Collection) {

	defer close(ch)
	defer wg.Wait()
	startTime := time.Now().UnixMilli()

	switch testType {
	case model.DurationType:
		index := 0
		duration := roundOrTime * 1000
		currentTime := time.Now().UnixMilli()
		for startTime+duration > currentTime {
			_, status := model.QueryPlanStatus(planId + ":" + sceneId + ":" + "status")
			if status == "false" {
				return
			}
			startCurrentTime := time.Now().UnixMilli()
			for i := int64(0); i < concurrent; i++ {
				wg.Add(1)
				go func(i, concurrent int64) {
					golink.DisposeScene(eventList, ch, planId, planName, sceneId, sceneName, reportId, reportName, configuration, globalVariable, wg, requestCollection, i, concurrent)
					wg.Done()
				}(i, concurrent)
				index++
			}

			// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
			distance := time.Now().UnixMilli() - startCurrentTime
			if distance < 1000 {
				sleepTime := time.Duration(1000-distance) * time.Millisecond
				time.Sleep(sleepTime)
			}
			currentTime = time.Now().UnixMilli()
		}
		log.Logger.Info(index)

	case model.RoundsType:
		for i := int64(0); i < roundOrTime; i++ {
			_, status := model.QueryPlanStatus(planId + ":" + sceneId + ":" + "status")
			if status == "false" {
				return
			}
			currentTime := time.Now().UnixMilli()
			for j := int64(0); j < concurrent; j++ {
				wg.Add(1)
				go func(i, concurrent int64) {
					golink.DisposeScene(eventList, ch, planId, planName, sceneId, sceneName, reportId, reportName, configuration, globalVariable, wg, requestCollection, i, concurrent)
					wg.Done()
				}(i, concurrent)
			}

			distance := time.Now().UnixMilli() - currentTime
			if distance < 1000 {
				sleepTime := time.Duration(1000-distance) * time.Millisecond
				log.Logger.Info("睡眠", distance)
				time.Sleep(sleepTime)
			}

		}
	}

}
