package execution

import (
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionConcurrentModel 并发模式
func ExecutionConcurrentModel(statusCh chan bool, ch chan *model.ResultDataMsg, plan *model.Plan) {
	defer close(ch)
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
				}
				currenWg.Wait()

				if time.Now().UnixMilli()-currentTime < 1000 {
					sleepTime := time.Duration(time.Now().UnixMilli()-currentTime) * time.Millisecond
					time.Sleep(sleepTime)
				}
				currentTime = time.Now().UnixMilli()
			}

		}
		log.Logger.Info("计划", plan.PlanName, "结束")

	case model.RoundsType:
		rounds := plan.ConfigTask.TestModel.ConcurrentTest.Rounds
		for i := int64(0); i < rounds; i++ {
			select {
			case status := <-statusCh:
				if status == false {
					log.Logger.Info("计划", plan.PlanName, "结束")
					return
				}
			default:
				currentTime := time.Now().UnixMilli()
				var currenWg = &sync.WaitGroup{}
				for j := int64(0); j < concurrent; j++ {
					currenWg.Add(1)
					go func(i, concurrent int64) {
						globalVariable := plan.Variable.VariableMap
						golink.Dispose(i, concurrent, eventList, ch, plan, globalVariable)
						currenWg.Done()
					}(i, concurrent)
				}
				currenWg.Wait()
				if time.Now().UnixMilli()-currentTime < 1000 {
					sleepTime := time.Duration(time.Now().UnixMilli()-currentTime) * time.Millisecond
					time.Sleep(sleepTime)
				}
			}

		}
		log.Logger.Info("计划", plan.PlanName, "结束")
	}

}
