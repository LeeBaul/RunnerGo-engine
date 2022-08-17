package execution

import (
	"github.com/Shopify/sarama"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"strconv"
	"sync"
	"time"
)

// ExecutionLadderModel 阶梯模式
func ExecutionLadderModel(kafkaProducer sarama.SyncProducer, plan *model.Plan, ch chan *model.TestResultDataMsg) {

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
	eventList := plan.Scene.EventList

	if plan.Scene.Configuration.ParameterizedFile.Path != "" {
		var mu = sync.Mutex{}
		plan.Scene.Configuration.ParameterizedFile.VariableNames.Mu = mu
		p := plan.Scene.Configuration.ParameterizedFile
		p.ReadFile()
	}

	// 只要开始时间+持续时长大于当前时间就继续循环
	for startTime+lengthDuration > time.Now().Unix() {
		var currenWg = &sync.WaitGroup{}
		for i := int64(0); i < concurrent; i++ {
			currenWg.Add(1)
			go func() {
				globalVariable := plan.Variable.VariableMap
				for _, event := range eventList {
					switch event.EventType {
					case model.RequestType:
						golink.Send(ch, plan, event.Request, globalVariable)
					case model.CollectionType:
						switch event.Controller.ControllerType {
						case model.IfControllerType:
							if v, ok := globalVariable.Load(event.Controller.IfController.Key); ok {
								event.Controller.IfController.PerForm(v.(string))
							}
						case model.CollectionType:
							// 集合点, 待开发
						case model.WaitControllerType:
							timeWait, _ := strconv.Atoi(event.Controller.WaitController.WaitTime)
							time.Sleep(time.Duration(timeWait) * time.Millisecond)
						}
					}
				}
			}()
			currenWg.Done()
			// 如果设置了启动并发时长
			if timeUp != 0 && (startConcurrent/timeUp)%i == 0 && i != 0 {
				time.Sleep(1 * time.Second)
			}
		}
		currenWg.Wait()
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
