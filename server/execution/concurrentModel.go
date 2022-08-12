package execution

import (
	"github.com/Shopify/sarama"
	"kp-runner/model"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionConcurrentModel 并发模式
func ExecutionConcurrentModel(kafkaProducer sarama.SyncProducer, wg *sync.WaitGroup, ch chan *model.TestResultDataMsg, plan model.Plan) {
	startTime := time.Now().Unix()
	concurrent := plan.ConfigTask.TestModel.ConcurrentTest.Concurrent
	eventList := plan.Scene.EventList
	timeUp := plan.ConfigTask.TestModel.ConcurrentTest.TimeUp
	testType := plan.ConfigTask.TestModel.ConcurrentTest.Type
	// go model.SendKafkaMsg(kafkaProducer, ch)

	switch testType {
	case model.DurationType:
		duration := plan.ConfigTask.TestModel.ConcurrentTest.Duration
		currentTime := time.Now().Unix()
		for startTime+duration > currentTime {
			for i := int64(0); i < concurrent; i++ {
				wg.Add(1)
				go func(eventList []model.Event) {
					globalVariable := plan.Variable.VariableMap
					for _, event := range eventList {
						switch event.EventType {
						case model.RequestType:
							golink.Send(ch, plan, wg, event.Request, globalVariable)
						case model.CollectionType:
							switch event.Controller.ControllerType {
							case model.IfControllerType:
								if v, ok := globalVariable[event.Controller.IfController.Key]; ok {
									event.Controller.IfController.PerForm(v)
								}
							case model.CollectionType:

							}
						}

					}
				}(eventList)

				if timeUp != 0 && (concurrent/timeUp)%i == 0 && i != 0 {
					time.Sleep(1000)
				}
			}
			if time.Now().Unix()-currentTime < 1 {
				time.Sleep(1 * time.Second)
			}
			currentTime = time.Now().Unix()

		}

	case model.RoundsType:
		rounds := plan.ConfigTask.TestModel.ConcurrentTest.Rounds
		for i := int64(0); i < rounds; i++ {
			for j := int64(0); j < concurrent; j++ {
				wg.Add(1)
				go func() {
					globalVariable := plan.Variable.VariableMap
					for _, event := range eventList {
						golink.Send(ch, plan, wg, event.Request, globalVariable)
					}
				}()

				if timeUp != 0 && (concurrent/timeUp)%i == 0 && i != 0 {
					time.Sleep(1000)
				}

			}
		}

	}

}
