package execution

import (
	"bytes"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// ConcurrentModel 并发模式
func ConcurrentModel(wg *sync.WaitGroup, scene *model.Scene, reportMsg *model.ResultDataMsg, resultDataMsgCh chan *model.ResultDataMsg, requestCollection *mongo.Collection) {

	defer close(resultDataMsgCh)

	startTime := time.Now().UnixMilli()
	concurrent := scene.ConfigTask.ModeConf.Concurrency
	planId := strconv.FormatInt(reportMsg.PlanId, 10)
	sceneId := reportMsg.SceneId
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime
	if scene.ConfigTask.ModeConf.Duration != 0 {
		index := 0
		duration := scene.ConfigTask.ModeConf.Duration * 1000
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

				if reheatTime > 0 && index == 0 {
					durationTime := time.Now().UnixMilli() - startTime
					if i%(concurrent/reheatTime) == 0 && durationTime < 1000 {
						time.Sleep(time.Duration(durationTime) * time.Millisecond)
					}
				}
			}
			index++
			// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
			distance := time.Now().UnixMilli() - startCurrentTime
			if distance < 1000 {
				sleepTime := time.Duration(1000-distance) * time.Millisecond
				time.Sleep(sleepTime)
			}
			currentTime = time.Now().UnixMilli()
		}

	} else {
		rounds := scene.ConfigTask.ModeConf.RoundNum
		for i := int64(0); i < rounds; i++ {
			index := 0
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
				if reheatTime > 0 && index == 0 {
					durationTime := time.Now().UnixMilli() - startTime
					if i%(concurrent/reheatTime) == 0 && durationTime < 1000 {
						time.Sleep(time.Duration(durationTime) * time.Millisecond)
					}
				}
			}
			index++

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
