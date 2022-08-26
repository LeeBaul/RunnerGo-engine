package execution

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/model"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionConcurrentModel 并发模式
func ExecutionConcurrentModel(ch chan *model.ResultDataMsg, plan *model.Plan, wg *sync.WaitGroup, requestCollection, responseCollection *mongo.Collection) {
	defer close(ch)
	defer wg.Wait()
	startTime := time.Now().UnixMilli()
	concurrent := plan.ConfigTask.TestModel.ConcurrentTest.Concurrent
	eventList := plan.Scene.EventList
	testType := plan.ConfigTask.TestModel.ConcurrentTest.Type
	if plan.Scene.Configuration.ParameterizedFile.Path != "" {
		p := plan.Scene.Configuration.ParameterizedFile
		p.VariableNames.Mu = sync.Mutex{}
		p.ReadFile()
	}
	switch testType {
	case model.DurationType:
		duration := plan.ConfigTask.TestModel.ConcurrentTest.Duration * 1000
		currentTime := time.Now().UnixMilli()
		for startTime+duration > currentTime {
			_, status := model.QueryPlanStatus(plan.PlanID + ":" + plan.Scene.SceneId + ":" + "status")
			if status == "false" {
				return
			}

			startCurrentTime := time.Now().UnixMilli()
			for i := int64(0); i < concurrent; i++ {
				wg.Add(1)
				go func(i, concurrent int64) {
					if plan.Variable == nil {
						plan.Variable = new(sync.Map)
					}
					globalVariable := plan.Variable
					golink.DisposeScene(eventList, ch, plan, globalVariable, wg, requestCollection, responseCollection, i, concurrent)
					wg.Done()
				}(i, concurrent)
			}

			// 如果发送的并发数时间小于1000ms，那么休息剩余的时间;也就是说每秒只发送concurrent个请求
			if time.Now().UnixMilli()-startCurrentTime < 1000 {
				sleepTime := time.Duration(time.Now().UnixMilli()-currentTime) * time.Millisecond
				time.Sleep(sleepTime)
			}
			currentTime = time.Now().UnixMilli()
		}

	case model.RoundsType:
		rounds := plan.ConfigTask.TestModel.ConcurrentTest.Rounds
		for i := int64(0); i < rounds; i++ {
			_, status := model.QueryPlanStatus(plan.PlanID + ":" + plan.Scene.SceneId + ":" + "status")
			if status == "false" {
				return
			}
			currentTime := time.Now().UnixMilli()
			for j := int64(0); j < concurrent; j++ {
				wg.Add(1)
				go func(i, concurrent int64) {
					globalVariable := plan.Variable
					golink.DisposeScene(eventList, ch, plan, globalVariable, wg, requestCollection, responseCollection, i, concurrent)
					wg.Done()
				}(i, concurrent)
			}

			if time.Now().UnixMilli()-currentTime < 1000 {
				sleepTime := time.Duration(time.Now().UnixMilli()-currentTime) * time.Millisecond
				time.Sleep(sleepTime)
			}

		}
	}

}
