// Package server 压测启动
package server

import (
	"context"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
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
	if plan.ConfigTask != nil && plan.Scene.EnablePlanConfiguration == true {
		plan.Scene.ConfigTask = plan.ConfigTask
	}
	switch plan.Scene.ConfigTask.TaskType {
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

	// 查询定时任务状态，如果redis中的状态变为停止，则关闭定时任务
	status := model.QueryTimingTaskStatus(plan.PlanId + ":" + plan.Scene.SceneId + ":" + "timing")
	if status == false {
		c.Remove(id)
	}

}

// ExecutionDebugRequest 接口调试
//func ExecutionDebugRequest(request model.Request, globalVariable *sync.Map, requestResults *model.ResultDataMsg, debugMsg *model.DebugMsg) {
//	golink.DisposeRequest(nil, "", "", "", "", "", "", nil, request, globalVariable, nil, requestResults, debugMsg, nil)
//}

// ExecutionPlan 执行计划
func ExecutionPlan(plan *model.Plan) {

	// 如果场景为空或者场景中的事件为空，直接结束该方法
	if plan.Scene == nil || plan.Scene.EventList == nil {
		log.Logger.Error("计划的场景不能为空: ", plan)
		return
	}

	// 设置kafka消费者，目的是像kafka中发送测试结果
	kafkaProducer, err := model.NewKafkaProducer([]string{config.Config["kafkaAddress"].(string)})
	if err != nil {
		log.Logger.Error("kafka连接失败", err)
		return
	}

	// 新建mongo客户端连接，用于发送debug数据
	mongoClient, err := model.NewMongoClient(
		config.Config["mongoUser"].(string),
		config.Config["mongoPassword"].(string),
		config.Config["mongoHost"].(string),
		config.Config["mongoDB"].(string))
	if err != nil {
		log.Logger.Error("连接mongo错误：", err)
		return
	}
	defer mongoClient.Disconnect(context.TODO())

	// 场景channel,用于各个event之间的通信
	//sceneCh := make(chan *model.Plan)

	// 设置接收数据缓存
	resultDataMsgCh := make(chan *model.ResultDataMsg, 10000)

	var wg = &sync.WaitGroup{}

	// 计算测试结果
	//sceneTestResultDataMsgCh := make(chan *model.SceneTestResultDataMsg, 10)
	//go ReceivingResults(ch, sceneTestResultDataMsgCh)
	// 向kafka发送消息
	go model.SendKafkaMsg(kafkaProducer, resultDataMsgCh, config.Config["Topic"].(string))

	requestCollection := model.NewCollection(config.Config["mongoDB"].(string), config.Config["mongoRequestTable"].(string), mongoClient)

	planId := plan.PlanId
	planName := plan.PlanName
	reportId := plan.ReportId
	reportName := plan.ReportName
	scene := plan.Scene

	// 如果场景中的任务配置勾选了全局任务配置，那么使用全局任务配置
	if scene.EnablePlanConfiguration == true {
		scene.ConfigTask = plan.ConfigTask
	}
	if scene.ConfigTask == nil {
		log.Logger.Error("任务配置不能为空", plan)
		return
	}
	// 循环读全局变量，如果场景变量不存在则将添加到场景变量中，如果有参数化数据则，将其替换
	if plan.Variable != nil {
		if scene.Configuration.Variable == nil {
			scene.Configuration.Variable = new(sync.Map)
		}
		plan.Variable.Range(func(key, value any) bool {
			if _, ok := scene.Configuration.Variable.Load(key); !ok {
				scene.Configuration.Variable.Store(key, value)
			}
			return true
		})
	}

	// 分解任务
	TaskDecomposition(planId, planName, reportId, reportName, scene, wg, resultDataMsgCh, scene.Configuration.Variable, requestCollection)
}

// TaskDecomposition 分解任务
func TaskDecomposition(planId, planName, reportId, reportName string, scene *model.Scene, wg *sync.WaitGroup, resultDataMsgCh chan *model.ResultDataMsg, sceneVariable *sync.Map, mongoCollection *mongo.Collection) {

	if scene.Configuration.ParameterizedFile != nil {
		var mu = sync.Mutex{}
		p := scene.Configuration.ParameterizedFile
		p.VariableNames.Mu = mu
		p.ReadFile()
	}

	configTask := scene.ConfigTask
	configuration := scene.Configuration
	eventList := scene.EventList
	sceneId := scene.SceneId
	sceneName := scene.SceneName
	switch configTask.TestModel.Type {
	case model.ConcurrentModel:
		execution.ExecutionConcurrentModel(
			configTask.TestModel.ConcurrentTest,
			resultDataMsgCh,
			eventList,
			planId,
			planName,
			reportId,
			reportName,
			sceneId,
			sceneName,
			configuration,
			wg,
			sceneVariable,
			mongoCollection)

	case model.ErrorRateModel:
		execution.ExecutionErrorRateModel(
			configTask.TestModel.ErrorRateTest,
			eventList,
			resultDataMsgCh,
			planId,
			planName,
			sceneId,
			sceneName,
			reportId,
			reportName,
			configuration,
			wg,
			sceneVariable,
			mongoCollection)
	case model.LadderModel:
		execution.ExecutionLadderModel(
			configTask.TestModel.LadderTest,
			eventList,
			resultDataMsgCh,
			planId,
			planName,
			sceneId,
			sceneName,
			reportId,
			reportName,
			configuration,
			wg,
			sceneVariable,
			mongoCollection)
		//case task.QpsModel:
		//	execution.ExecutionQpsModel()
	case model.RTModel:
		execution.ExecutionRTModel(
			configTask.TestModel.RTTest,
			eventList,
			resultDataMsgCh,
			planId,
			planName,
			sceneId,
			sceneName,
			reportId,
			reportName,
			configuration,
			wg,
			sceneVariable,
			mongoCollection)
	default:
		close(resultDataMsgCh)

	}
	wg.Wait()
	log.Logger.Info("计划", planName, "结束")
}

// ReceivingResults 计算并发送测试结果,
func ReceivingResults(resultDataMsgCh <-chan *model.ResultDataMsg, sceneTestResultDataMsgCh chan *model.SceneTestResultDataMsg) {
	var (
		sceneTestResultDataMsg = new(model.SceneTestResultDataMsg)
		requestTimeMap         = make(map[interface{}]model.RequestTimeList)
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

			requestTimeMap[resultDataMsg.TargetId] = append(requestTimeMap[resultDataMsg.TargetId], resultDataMsg.RequestTime)
			// 将各个接口的响应时间进行排序
			sort.Sort(requestTimeMap[resultDataMsg.TargetId])
			if sceneTestResultDataMsg.Results == nil {
				sceneTestResultDataMsg.Results = make(map[interface{}]*model.ApiTestResultDataMsg)
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId] == nil {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId] = new(model.ApiTestResultDataMsg)
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanId == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanId = resultDataMsg.PlanId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanName = resultDataMsg.PlanName
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].SceneId == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].SceneId = resultDataMsg.SceneId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].SceneName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].SceneName = resultDataMsg.SceneName
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].ReportId == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].ReportId = resultDataMsg.ReportId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].ReportName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].ReportName = resultDataMsg.ReportName
			}
			sceneTestResultDataMsg.Results[resultDataMsg.TargetId].TargetId = resultDataMsg.TargetId

			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].Name == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].Name = resultDataMsg.Name
			}
			sceneTestResultDataMsg.Results[resultDataMsg.TargetId].ReceivedBytes += resultDataMsg.ReceivedBytes
			sceneTestResultDataMsg.Results[resultDataMsg.TargetId].SendBytes += resultDataMsg.SendBytes
			if resultDataMsg.IsSucceed {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].SuccessNum += 1
			} else {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].ErrorNum += 1
			}
			sceneTestResultDataMsg.Results[resultDataMsg.TargetId].TotalRequestNum += 1
			sceneTestResultDataMsg.Results[resultDataMsg.TargetId].TotalRequestTime += resultDataMsg.RequestTime
			sceneTestResultDataMsg.Results[resultDataMsg.TargetId].CustomRequestTimeLineValue = resultDataMsg.CustomRequestTimeLine

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

//// DebugScene 场景调试
//func DebugScene() {
//	gid := GetGid()
//	golink.DisposeScene(gid, eventList, ch, planId, planName, sceneId, sceneName, reportId, reportName, configuration, wg, sceneVariable, requestCollection, i, concurrent)
//}
//

// DebugApi api调试
func DebugApi(Api model.Api) {
	event := model.Event{}
	event.Api = Api
	event.Api.Weight = 100
	event.EventId = "接口调试"
	wg := &sync.WaitGroup{}
	sceneVariable := new(sync.Map)
	// 新建mongo客户端连接，用于发送debug数据
	mongoClient, err := model.NewMongoClient(
		config.Config["mongoUser"].(string),
		config.Config["mongoPassword"].(string),
		config.Config["mongoHost"].(string),
		config.Config["mongoDB"].(string))
	if err != nil {
		log.Logger.Error("连接mongo错误：", err)
		return
	}
	defer mongoClient.Disconnect(context.TODO())
	var debugMsg = &model.DebugMsg{}
	mongoCollection := model.NewCollection(config.Config["mongoDB"].(string), config.Config["mongoRequestTable"].(string), mongoClient)

	wg.Add(1)
	go golink.DisposeRequest(nil, "", "", "", "", "", "", nil, event, wg, nil, debugMsg, sceneVariable, mongoCollection)
	wg.Wait()
}
