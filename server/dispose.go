// Package server 压测启动
package server

import (
	"context"
	"github.com/robfig/cron/v3"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/execution"
	"kp-runner/server/golink"
	"kp-runner/server/heartbeat"
	"sort"
	"sync"
	"time"
)

// DisposeTask 处理任务
func DisposeTask(plan *model.Plan) {
	switch plan.ConfigTask.TaskType {
	case model.CommonTaskType:
		ExecutionPlan(plan)
	case model.TimingTaskType:
		TimingExecutionPlan(plan, func() {
			ExecutionPlan(plan)
		})
	case model.CICDTaskType:

	}
}

// TimingExecutionPlan 定时任务
func TimingExecutionPlan(plan *model.Plan, job func()) {
	if plan.ConfigTask.Task.TimingTask.Spec == "" {
		log.Logger.Error("定时任务，执行时间不能为空")
		return
	}
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithSeconds(),
	)

	id, err := c.AddFunc(plan.ConfigTask.Task.TimingTask.Spec, job)
	if err != nil {
		log.Logger.Error("定时任务执行失败", err)
		return
	}
	c.Start()

	status := model.QueryTimingTaskStatus(plan.PlanId + ":" + plan.Scene.SceneId + ":" + "timing")
	if status == false {
		c.Remove(id)
	}

}

func ExecutionDebugRequest(request model.Request, globalVariable *sync.Map, requestResults *model.ResultDataMsg, debugMsg *model.DebugMsg) {
	golink.DisposeRequest(nil, "", "", "", "", "", "", nil, request, globalVariable, nil, requestResults, debugMsg, nil)
}

// ExecutionPlan 执行计划
func ExecutionPlan(plan *model.Plan) {

	// 设置kafka消费者
	kafkaProducer, err := model.NewKafkaProducer([]string{config.Config["kafkaAddress"].(string)})
	if err != nil {
		log.Logger.Error("kafka连接失败", err)
		return
	}
	// 设置接收数据缓存
	ch := make(chan *model.ResultDataMsg, 10000)
	sceneTestResultDataMsgCh := make(chan *model.SceneTestResultDataMsg, 10)
	var wg = &sync.WaitGroup{}

	// 计算测试结果
	go ReceivingResults(ch, sceneTestResultDataMsgCh)
	// 向kafka发送消息
	go model.SendKafkaMsg(kafkaProducer, sceneTestResultDataMsgCh)
	mongoClient, err := model.NewMongoClient(
		config.Config["mongoUser"].(string),
		config.Config["mongoPassword"].(string),
		config.Config["mongoHost"].(string),
		config.Config["mongoDB"].(string))
	if err != nil {
		close(ch)
		log.Logger.Error("连接mongo错误：", err)
		return
	}
	defer mongoClient.Disconnect(context.TODO())
	requestCollection := model.NewCollection(config.Config["mongoDB"].(string), config.Config["mongoRequestTable"].(string), mongoClient)
	if plan.Scene.Configuration.ParameterizedFile.Path != "" {
		var mu = sync.Mutex{}
		plan.Scene.Configuration.ParameterizedFile.VariableNames.Mu = mu
		p := plan.Scene.Configuration.ParameterizedFile
		p.ReadFile()
	}

	configuration := plan.Scene.Configuration

	globalVariable := plan.Variable
	eventList := plan.Scene.EventList
	switch plan.ConfigTask.TestModel.Type {
	case model.ConcurrentModel:
		concurrent := plan.ConfigTask.TestModel.ConcurrentTest.Concurrent
		testType := plan.ConfigTask.TestModel.ConcurrentTest.Type
		roundOrTime := int64(0)
		switch testType {
		case model.DurationType:
			roundOrTime = plan.ConfigTask.TestModel.ConcurrentTest.Duration
		case model.RoundsType:
			roundOrTime = plan.ConfigTask.TestModel.ConcurrentTest.Rounds
		}
		execution.ExecutionConcurrentModel(
			eventList,
			ch,
			concurrent,
			testType,
			roundOrTime,
			plan.PlanId,
			plan.PlanName,
			plan.Scene.SceneId,
			plan.Scene.SceneName,
			plan.ReportId,
			plan.ReportName,
			configuration,
			globalVariable,
			wg,
			requestCollection)

	case model.ErrorRateModel:
		startConcurrent := plan.ConfigTask.TestModel.ErrorRatTest.StartConcurrent
		length := plan.ConfigTask.TestModel.ErrorRatTest.Length
		maxConcurrent := plan.ConfigTask.TestModel.ErrorRatTest.MaxConcurrent
		lengthDuration := plan.ConfigTask.TestModel.ErrorRatTest.LengthDuration
		stableDuration := plan.ConfigTask.TestModel.ErrorRatTest.StableDuration
		timeUp := plan.ConfigTask.TestModel.ErrorRatTest.TimeUp
		execution.ExecutionErrorRateModel(
			eventList,
			ch,
			plan.PlanId,
			plan.PlanName,
			plan.Scene.SceneId,
			plan.Scene.SceneName,
			plan.ReportId,
			plan.ReportName,
			startConcurrent,
			length,
			maxConcurrent,
			lengthDuration,
			stableDuration,
			timeUp,
			configuration,
			globalVariable,
			wg,
			requestCollection)
	case model.LadderModel:
		startConcurrent := plan.ConfigTask.TestModel.LadderTest.StartConcurrent
		length := plan.ConfigTask.TestModel.LadderTest.Length
		maxConcurrent := plan.ConfigTask.TestModel.LadderTest.MaxConcurrent
		lengthDuration := plan.ConfigTask.TestModel.LadderTest.LengthDuration
		stableDuration := plan.ConfigTask.TestModel.LadderTest.StableDuration
		timeUp := plan.ConfigTask.TestModel.LadderTest.TimeUp
		execution.ExecutionLadderModel(
			eventList,
			ch,
			plan.PlanId,
			plan.PlanName,
			plan.Scene.SceneId,
			plan.Scene.SceneName,
			plan.ReportId,
			plan.ReportName,
			startConcurrent,
			length,
			maxConcurrent,
			lengthDuration,
			stableDuration,
			timeUp,
			configuration,
			wg,
			globalVariable,
			requestCollection)
		//case task.TpsModel:
		//	execution.ExecutionTpsModel()
		//case task.QpsModel:
		//	execution.ExecutionQpsModel()
	case model.RTModel:
		startConcurrent := plan.ConfigTask.TestModel.RTTest.StartConcurrent
		length := plan.ConfigTask.TestModel.RTTest.Length
		maxConcurrent := plan.ConfigTask.TestModel.RTTest.MaxConcurrent
		lengthDuration := plan.ConfigTask.TestModel.RTTest.LengthDuration
		stableDuration := plan.ConfigTask.TestModel.RTTest.StableDuration
		timeUp := plan.ConfigTask.TestModel.RTTest.TimeUp
		execution.ExecutionRTModel(
			eventList,
			ch,
			wg,
			plan.PlanId,
			plan.PlanName,
			plan.Scene.SceneId,
			plan.Scene.SceneName,
			plan.ReportId,
			plan.ReportName,
			startConcurrent,
			length,
			maxConcurrent,
			lengthDuration,
			stableDuration,
			timeUp,
			configuration,
			globalVariable,
			requestCollection)
	default:
		close(ch)

	}

	log.Logger.Info("计划", plan.PlanName, "结束")
}

// ReceivingResults 计算并发送测试结果
func ReceivingResults(resultDataMsgCh <-chan *model.ResultDataMsg, sceneTestResultDataMsgCh chan *model.SceneTestResultDataMsg) {
	var (
		sceneTestResultDataMsg = new(model.SceneTestResultDataMsg)
		requestTimeMap         = make(map[string]model.RequestTimeList)
	)
	sceneTestResultDataMsg.MachineIp = heartbeat.LocalIp
	// 关闭通道
	defer close(sceneTestResultDataMsgCh)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case resultDataMsg, ok := <-resultDataMsgCh:
			if !ok {
				goto end
			}
			if sceneTestResultDataMsg.PlanId == "" {
				sceneTestResultDataMsg.PlanId = resultDataMsg.PlanId
			}
			if sceneTestResultDataMsg.PlanName == "" {
				sceneTestResultDataMsg.PlanName = resultDataMsg.PlanName
			}
			if sceneTestResultDataMsg.SceneId == "" {
				sceneTestResultDataMsg.SceneId = resultDataMsg.SceneId
			}
			if sceneTestResultDataMsg.SceneName == "" {
				sceneTestResultDataMsg.SceneName = resultDataMsg.SceneName
			}
			if sceneTestResultDataMsg.ReportId == "" {
				sceneTestResultDataMsg.ReportId = resultDataMsg.ReportId
			}
			if sceneTestResultDataMsg.ReportName == "" {
				sceneTestResultDataMsg.ReportName = resultDataMsg.ReportName
			}

			requestTimeMap[resultDataMsg.ApiId] = append(requestTimeMap[resultDataMsg.ApiId], resultDataMsg.RequestTime)
			// 将各个接口的响应时间进行排序
			sort.Sort(requestTimeMap[resultDataMsg.ApiId])
			if sceneTestResultDataMsg.Results == nil {
				sceneTestResultDataMsg.Results = make(map[string]*model.ApiTestResultDataMsg)
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId] == nil {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId] = new(model.ApiTestResultDataMsg)
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].PlanId == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].PlanId = resultDataMsg.PlanId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].PlanName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].PlanName = resultDataMsg.PlanName
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].SceneId == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].SceneId = resultDataMsg.SceneId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].SceneName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].SceneName = resultDataMsg.SceneName
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ReportId == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ReportId = resultDataMsg.ReportId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ReportName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ReportName = resultDataMsg.ReportName
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ApiId == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ApiId = resultDataMsg.ApiId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ApiName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ApiName = resultDataMsg.ApiName
			}
			sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ReceivedBytes += resultDataMsg.ReceivedBytes
			sceneTestResultDataMsg.Results[resultDataMsg.ApiId].SendBytes += resultDataMsg.SendBytes
			if resultDataMsg.IsSucceed {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].SuccessNum += 1
			} else {
				sceneTestResultDataMsg.Results[resultDataMsg.ApiId].ErrorNum += 1
			}
			sceneTestResultDataMsg.Results[resultDataMsg.ApiId].TotalRequestNum += 1
			sceneTestResultDataMsg.Results[resultDataMsg.ApiId].TotalRequestTime += resultDataMsg.RequestTime
			sceneTestResultDataMsg.Results[resultDataMsg.ApiId].CustomRequestTimeLineValue = resultDataMsg.CustomRequestTimeLine

		// 定时每秒发送一次场景的测试结果
		case <-ticker.C:
			for k, v := range requestTimeMap {
				if v != nil {
					sort.Sort(v)
					sceneTestResultDataMsg.Results[k].MinRequestTime = v[0]
					sceneTestResultDataMsg.Results[k].MaxRequestTime = v[len(v)-1]
					sceneTestResultDataMsg.Results[k].AvgRequestTime = sceneTestResultDataMsg.Results[k].TotalRequestTime / sceneTestResultDataMsg.Results[k].TotalRequestNum
					sceneTestResultDataMsg.Results[k].NinetyRequestTimeLine = timeLineCalculate(90, v)
					sceneTestResultDataMsg.Results[k].NinetyFiveRequestTimeLine = timeLineCalculate(95, v)
					sceneTestResultDataMsg.Results[k].NinetyNineRequestTimeLine = timeLineCalculate(99, v)
					sceneTestResultDataMsg.Results[k].CustomRequestTimeLine = timeLineCalculate(sceneTestResultDataMsg.Results[k].CustomRequestTimeLineValue, v)
				}

			}
			sceneTestResultDataMsgCh <- sceneTestResultDataMsg
		}
	}
end:
	return

}

// 根据响应时间线，计算该线的值
func timeLineCalculate(line int64, requestTimeList model.RequestTimeList) (requestTime uint64) {
	if line > 0 && line < 100 {
		proportion := float64(line) / 100
		value := proportion * float64(len(requestTimeList))
		requestTime = requestTimeList[int(value)]
	}
	return

}
