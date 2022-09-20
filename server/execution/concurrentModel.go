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

	startTime := time.Now().UnixMilli()
	concurrent := scene.ConfigTask.ModeConf.Concurrency
	reheatTime := scene.ConfigTask.ModeConf.ReheatTime

	if scene.ConfigTask.ModeConf.Duration != 0 {
		log.Logger.Info("开始性能测试,持续时间", scene.ConfigTask.ModeConf.Duration, "秒")
		index := 0
		duration := scene.ConfigTask.ModeConf.Duration * 1000
		currentTime := time.Now().UnixMilli()

		for startTime+duration > currentTime {
			// 查询是否停止
			_, status := model.QueryPlanStatus(reportMsg.ReportId + ":status")
			if status == "stop" {
				return
			}
			// 查询是否开启debug
			_, debug := model.QueryPlanStatus(reportMsg.ReportId + ":debug")
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
					golink.DisposeScene(wg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
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
		index := 0
		rounds := scene.ConfigTask.ModeConf.RoundNum
		for i := int64(0); i < rounds; i++ {

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
			currentTime := time.Now().UnixMilli()
			for j := int64(0); j < concurrent; j++ {
				wg.Add(1)
				go func(i, concurrent int64) {
					gid := GetGid()
					golink.DisposeScene(wg, gid, model.PlanType, scene, reportMsg, resultDataMsgCh, requestCollection, i, concurrent)
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
