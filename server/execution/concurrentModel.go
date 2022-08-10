package execution

import (
	"fmt"
	"github.com/Shopify/sarama"
	"kp-runner/model"
	"kp-runner/model/plan"
	"kp-runner/model/request"
	"kp-runner/model/task"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionConcurrentModel 并发模式
func ExecutionConcurrentModel(kafkaProducer sarama.SyncProducer, wg *sync.WaitGroup, ch chan *model.TestResultDataMsg, plan plan.Plan) {
	startTime := time.Now().Unix()
	concurrent := plan.ConfigTask.TestModel.ConcurrentTest.Concurrent
	requests := plan.Scene.Requests
	timeUp := plan.ConfigTask.TestModel.ConcurrentTest.TimeUp
	testType := plan.ConfigTask.TestModel.ConcurrentTest.Type
	// go model.SendKafkaMsg(kafkaProducer, ch)
	fmt.Println("requests: ", requests)
	switch testType {
	case task.DurationType:
		duration := plan.ConfigTask.TestModel.ConcurrentTest.Duration
		for startTime+duration > time.Now().Unix() {
			for i := int64(0); i < concurrent; i++ {
				wg.Add(1)
				go func(requests []request.Request) {
					for _, request := range requests {
						golink.Send(ch, plan, wg, request)
					}
				}(requests)

				if timeUp != 0 && (concurrent/timeUp)%i == 0 && i != 0 {
					time.Sleep(1000)
				}
			}

		}

	case task.RoundsType:
		rounds := plan.ConfigTask.TestModel.ConcurrentTest.Rounds
		for i := int64(0); i < rounds; i++ {
			for j := int64(0); j < concurrent; j++ {
				wg.Add(1)
				go func() {
					for _, request := range requests {
						golink.Send(ch, plan, wg, request)
					}
				}()

				if timeUp != 0 && (concurrent/timeUp)%i == 0 && i != 0 {
					time.Sleep(1000)
				}

			}
		}

	}

}
