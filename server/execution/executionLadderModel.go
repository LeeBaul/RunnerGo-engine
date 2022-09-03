package execution

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/golink"
	"sync"
	"time"
)

// ExecutionLadderModel 阶梯模式
func ExecutionLadderModel(ladderTest model.LadderTest, eventList []model.Event, resultDataMsgCh chan *model.ResultDataMsg,
	planId, planName, sceneId, sceneName, reportId, reportName string,

	configuration *model.Configuration, wg *sync.WaitGroup, sceneVariable *sync.Map, requestCollection *mongo.Collection) {

	defer close(resultDataMsgCh)

	startConcurrent := ladderTest.StartConcurrent
	length := ladderTest.Length
	maxConcurrent := ladderTest.MaxConcurrent
	lengthDuration := ladderTest.LengthDuration
	stableDuration := ladderTest.StableDuration
	timeUp := ladderTest.TimeUp

	// 连接es，并查询当前错误率为多少，并将其放入到chan中
	startTime := time.Now().Unix()
	// preConcurrent 是为了回退,此功能后续开发
	//preConcurrent := startConcurrent

	concurrent := startConcurrent

	// 只要开始时间+持续时长大于当前时间就继续循环
	for startTime+lengthDuration > time.Now().Unix() {
		// 查询任务是否结束
		_, status := model.QueryPlanStatus(planId + ":" + sceneId + ":" + "status")
		if status == "false" {
			log.Logger.Info("计划", planName, "结束")
			return
		}
		for i := int64(0); i < concurrent; i++ {
			wg.Add(1)
			go func(i, concurrent int64, wg *sync.WaitGroup) {
				gid := GetGid()
				golink.DisposeScene(gid, eventList, resultDataMsgCh, planId, planName, sceneId, sceneName, reportId, reportName, configuration, wg, sceneVariable, requestCollection, i, concurrent)
				wg.Done()
			}(i, concurrent, wg)

			// 如果设置了启动并发时长
			if timeUp != 0 && (startConcurrent/timeUp)%i == 0 && i != 0 {
				time.Sleep(1 * time.Second)
			}
		}
		if concurrent == maxConcurrent && lengthDuration == stableDuration && startTime+lengthDuration >= time.Now().Unix() {
			log.Logger.Info("计划", planName, "结束")
			return
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

}
