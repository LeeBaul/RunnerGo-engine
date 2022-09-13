package execution

import (
	"bytes"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"runtime"
	"sync"
	"time"
)

// ConcurrentModel 并发模式
func ConcurrentModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {

	defer close(resultDataMsgCh)

	startTime := time.Now().UnixMilli()
	concurrentTest := scene.ConfigTask.TestModel.ConcurrentTest
	concurrent := concurrentTest.Concurrent
	planId := reportMsg.PlanId
	sceneId := reportMsg.SceneId
	switch concurrentTest.Type {
	case model.DurationType:
		index := 0
		duration := concurrentTest.Duration * 1000
		currentTime := time.Now().UnixMilli()

		for startTime+duration > currentTime {
			_, status := model.QueryPlanStatus(planId + ":" + sceneId + ":status")
			if status == "false" {
				return
			}
			_, debug := model.QueryPlanStatus(planId + ":" + sceneId + ":debug")
			if debug != "" {
				scene.Debug = debug
			} else {
				scene.Debug = ""
			}
			startCurrentTime := time.Now().UnixMilli()
			for i := int64(0); i < concurrent; i++ {
				wg.Add(1)
				go func(i, concurrent int64) {
					gid := GetGid()
					golink.DisposeScene(wg, gid, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
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
		rounds := concurrentTest.Rounds
		for i := int64(0); i < rounds; i++ {
			_, status := model.QueryPlanStatus(planId + ":" + sceneId + ":" + "status")
			if status == "false" {
				return
			}
			currentTime := time.Now().UnixMilli()
			for j := int64(0); j < concurrent; j++ {
				wg.Add(1)
				go func(i, concurrent int64) {
					gid := GetGid()
					golink.DisposeScene(wg, gid, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
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

func GetGid() (gid string) {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	gid = string(b)
	return
}
