package execution

import (
	"github.com/Shopify/sarama"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionLadderModel 阶梯模式
func ExecutionLadderModel(kafkaProducer sarama.SyncProducer, wg *sync.WaitGroup, plan model.Plan, ch chan *model.TestResultDataMsg) {

	//go model.SendKafkaMsg(kafkaProducer, ch)
	// 连接es，并查询当前错误率为多少，并将其放入到chan中
	startTime := time.Now().Unix()
	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent
	startConcurrent := plan.ConfigTask.TestModel.LadderTest.StartConcurrent
	length := plan.ConfigTask.TestModel.LadderTest.Length
	maxConcurrent := plan.ConfigTask.TestModel.LadderTest.MaxConcurrent
	lengthDuration := plan.ConfigTask.TestModel.LadderTest.LengthDuration
	stableDuration := plan.ConfigTask.TestModel.LadderTest.StableDuration
	timeUp := plan.ConfigTask.TestModel.LadderTest.TimeUp
	concurrent := startConcurrent
	requests := plan.Scene.Requests
	// 只要开始时间+持续时长大于当前时间就继续循环
	for startTime+int64(lengthDuration) > time.Now().Unix() {
		for i := int64(0); i < concurrent; i++ {
			wg.Add(1)
			go func() {
				for _, request := range requests {
					golink.Send(ch, plan, wg, request)
				}
			}()
			// 如果设置了启动并发时长
			if timeUp != 0 && (startConcurrent/timeUp)%i == 0 && i != 0 {
				time.Sleep(1 * time.Second)
			}
		}

		if concurrent == maxConcurrent && lengthDuration == stableDuration && startTime+int64(lengthDuration) >= time.Now().Unix() {
			goto end
		}
		// 如果当前并发数小于最大并发数，
		if concurrent <= maxConcurrent {
			// 从开始时间算起，加上持续时长。如果大于现在是的时间，说明已经运行了持续时长的时间，那么就要将开始时间的值，变为现在的时间值
			if startTime+int64(lengthDuration) >= time.Now().Unix() {
				startTime = time.Now().Unix()
				if concurrent+length <= maxConcurrent {
					concurrent = concurrent + length
				} else {
					concurrent = maxConcurrent
					lengthDuration = stableDuration
				}
			}
		}

	}
end:
	log.Logger.Info("计划", plan.PlanName, "结束")
}
