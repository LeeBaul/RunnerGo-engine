package execution

import (
	"github.com/Shopify/sarama"
	"kp-runner/model"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionConcurrentModel 并发模式
func ExecutionConcurrentModel(kafkaProducer sarama.SyncProducer, ch chan *model.ResultDataMsg, plan *model.Plan) {
	startTime := time.Now().UnixMilli()
	concurrent := plan.ConfigTask.TestModel.ConcurrentTest.Concurrent
	eventList := plan.Scene.EventList
	testType := plan.ConfigTask.TestModel.ConcurrentTest.Type
	// go model.SendKafkaMsg(kafkaProducer, ch)
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
			var currenWg = &sync.WaitGroup{}
			for i := int64(0); i < concurrent; i++ {
				currenWg.Add(1)
				go func() {
					if plan.Variable.VariableMap == nil {
						plan.Variable.VariableMap = new(sync.Map)
					}
					globalVariable := plan.Variable.VariableMap
					golink.Dispose(eventList, ch, plan, globalVariable)
					currenWg.Done()
				}()
			}
			currenWg.Wait()
			if time.Now().UnixMilli()-currentTime < 1000 {
				sleepTime := time.Duration(time.Now().UnixMilli()-currentTime) * time.Millisecond
				time.Sleep(sleepTime)
			}
			currentTime = time.Now().UnixMilli()
		}

	case model.RoundsType:
		rounds := plan.ConfigTask.TestModel.ConcurrentTest.Rounds
		for i := int64(0); i < rounds; i++ {
			currentTime := time.Now().UnixMilli()
			var currenWg = &sync.WaitGroup{}
			for j := int64(0); j < concurrent; j++ {
				currenWg.Add(1)
				go func() {
					globalVariable := plan.Variable.VariableMap
					golink.Dispose(eventList, ch, plan, globalVariable)
					currenWg.Done()
				}()
			}
			currenWg.Wait()
			if time.Now().UnixMilli()-currentTime < 1000 {
				sleepTime := time.Duration(time.Now().UnixMilli()-currentTime) * time.Millisecond
				time.Sleep(sleepTime)
			}
		}

	}

}
