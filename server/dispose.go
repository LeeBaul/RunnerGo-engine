// Package server 压测启动
package server

import (
	"context"
	"encoding/json"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/execution"
	"kp-runner/server/golink"
	"kp-runner/server/heartbeat"
	"sort"
	"strconv"
	"sync"
	"time"
)

// DisposeTask 处理任务
func DisposeTask(plan *model.Plan) {
	if plan.ConfigTask != nil {
		plan.Scene.ConfigTask = plan.ConfigTask
	}
	configTaskJson, _ := json.Marshal(plan.Scene.ConfigTask)
	log.Logger.Info("plan.scene.Config:", string(configTaskJson))
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
	if plan.ConfigTask.CronExpr == "" {
		log.Logger.Error("定时任务，执行时间不能为空")
		return
	}
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithSeconds(),
	)

	id, err := c.AddFunc(plan.ConfigTask.CronExpr, job)
	if err != nil {
		log.Logger.Error("定时任务执行失败", err)
		return
	}
	c.Start()

	// 查询定时任务状态，如果redis中的状态变为停止，则关闭定时任务
	status := model.QueryTimingTaskStatus(strconv.FormatInt(plan.PlanId, 10) + ":" + strconv.FormatInt(plan.Scene.SceneId, 10) + ":" + "timing")
	if status == false {
		c.Remove(id)
	}

}

// ExecutionPlan 执行计划
func ExecutionPlan(plan *model.Plan) {

	// 如果场景为空或者场景中的事件为空，直接结束该方法
	if plan.Scene == nil || plan.Scene.Nodes == nil {
		log.Logger.Error("计划的场景不能为空: ", plan)
		return
	}

	// 设置kafka消费者，目的是像kafka中发送测试结果
	kafkaProducer, err := model.NewKafkaProducer([]string{config.Conf.Kafka.Address})
	if err != nil {
		log.Logger.Error("kafka连接失败", err)
		return
	}

	// 新建mongo客户端连接，用于发送debug数据
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.User,
		config.Conf.Mongo.Password,
		config.Conf.Mongo.Address,
		config.Conf.Mongo.DB)
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
	go model.SendKafkaMsg(kafkaProducer, resultDataMsgCh, config.Conf.Kafka.TopIc)

	requestCollection := model.NewCollection(config.Conf.Mongo.DB, config.Conf.Mongo.StressDebugTable, mongoClient)
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
			scene.Configuration.Variable = []*model.KV{}
		}
		for _, value := range plan.Variable {
			var target = false
			for _, kv := range scene.Configuration.Variable {
				if value == kv {
					target = true
					break
				}
			}
			if target == false {
				scene.Configuration.Variable = append(scene.Configuration.Variable, value)
			}
		}
	}

	// 分解任务
	TaskDecomposition(plan, wg, resultDataMsgCh, requestCollection)
}

// TaskDecomposition 分解任务
func TaskDecomposition(plan *model.Plan, wg *sync.WaitGroup, resultDataMsgCh chan *model.ResultDataMsg, mongoCollection *mongo.Collection) {
	defer close(resultDataMsgCh)
	scene := plan.Scene
	scene.ReportId = plan.ReportId
	if scene.Configuration == nil {
		scene.Configuration = new(model.Configuration)
	}
	if scene.Configuration.Variable == nil {
		scene.Configuration.Variable = []*model.KV{}
	}
	if scene.Configuration.ParameterizedFile == nil {
		scene.Configuration.ParameterizedFile = new(model.ParameterizedFile)
	}
	if scene.Configuration.ParameterizedFile.VariableNames == nil {
		scene.Configuration.ParameterizedFile.VariableNames = new(model.VariableNames)
	}
	if scene.Configuration.ParameterizedFile.VariableNames.VarMapList == nil {
		scene.Configuration.ParameterizedFile.VariableNames.VarMapList = make(map[string][]string)
	}
	if scene.Configuration.ParameterizedFile != nil {
		p := scene.Configuration.ParameterizedFile
		p.VariableNames.Mu = sync.Mutex{}
		p.ReadFile()
	}

	var reportMsg = &model.ResultDataMsg{}
	if plan.MachineNum <= 0 {
		plan.MachineNum = 1
	}
	reportMsg.PlanId = plan.PlanId
	reportMsg.SceneId = scene.SceneId
	reportMsg.SceneName = scene.SceneName
	reportMsg.PlanName = plan.PlanName
	reportMsg.ReportId = plan.ReportId
	reportMsg.ReportName = plan.ReportName
	reportMsg.MachineNum = plan.MachineNum
	testModelJson, _ := json.Marshal(scene.ConfigTask.ModeConf)
	log.Logger.Info("plan.scene.Config:", string(testModelJson))
	switch scene.ConfigTask.Mode {
	case model.ConcurrentModel:
		execution.ConcurrentModel(wg, scene, reportMsg, resultDataMsgCh, mongoCollection)
	case model.ErrorRateModel:
		execution.ErrorRateModel(wg, scene, reportMsg, resultDataMsgCh, mongoCollection)
	case model.LadderModel:
		execution.LadderModel(wg, scene, reportMsg, resultDataMsgCh, mongoCollection)
		//case task.QpsModel:
		//	execution.ExecutionQpsModel()
	case model.RTModel:
		execution.RTModel(wg, scene, reportMsg, resultDataMsgCh, mongoCollection)
	case model.QpsModel:
		execution.QPSModel(wg, scene, reportMsg, resultDataMsgCh, mongoCollection)

	}
	wg.Wait()
	debugMsg := make(map[string]interface{})
	debugMsg["reportId"] = plan.ReportId
	debugMsg["end"] = true
	model.Insert(mongoCollection, debugMsg)

	log.Logger.Info("计划:", plan.PlanId, ".............结束")

}

// DebugScene 场景调试
func DebugScene(scene *model.Scene) {
	gid := execution.GetGid()
	wg := &sync.WaitGroup{}
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.User,
		config.Conf.Mongo.Password,
		config.Conf.Mongo.Address,
		config.Conf.Mongo.DB)
	if err != nil {
		log.Logger.Error("连接mongo错误：", err)
		return
	}
	if scene.Configuration == nil {
		scene.Configuration = new(model.Configuration)
		scene.Configuration.Variable = []*model.KV{}
		scene.Configuration.Mu = sync.Mutex{}
	}
	if scene.Configuration.Variable == nil {
		scene.Configuration.Variable = []*model.KV{}
		scene.Configuration.Mu = sync.Mutex{}
	}
	if scene.Configuration.ParameterizedFile != nil {
		p := scene.Configuration.ParameterizedFile
		if p.VariableNames == nil {
			p.VariableNames = new(model.VariableNames)
		}
		p.VariableNames.Mu = sync.Mutex{}
		p.ReadFile()
	}

	scene.Debug = model.All
	defer mongoClient.Disconnect(context.TODO())
	mongoCollection := model.NewCollection(config.Conf.Mongo.DB, config.Conf.Mongo.SceneDebugTable, mongoClient)
	golink.DisposeScene(wg, gid, model.SceneType, scene, nil, nil, mongoCollection)

	wg.Wait()

}

// DebugApi api调试
func DebugApi(Api model.Api) {
	event := model.Event{}
	event.Api = Api
	event.Weight = 100
	event.Id = "接口调试"
	wg := &sync.WaitGroup{}
	// 新建mongo客户端连接，用于发送debug数据
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.User,
		config.Conf.Mongo.Password,
		config.Conf.Mongo.Address,
		config.Conf.Mongo.DB)
	if err != nil {
		log.Logger.Error("连接mongo错误：", err)
		return
	}
	defer mongoClient.Disconnect(context.TODO())
	mongoCollection := model.NewCollection(config.Conf.Mongo.DB, config.Conf.Mongo.ApiDebugTable, mongoClient)
	configuration := new(model.Configuration)
	configuration.Variable = []*model.KV{}
	configuration.Mu = sync.Mutex{}
	wg.Add(1)
	go golink.DisposeRequest(wg, nil, nil, nil, configuration, event, mongoCollection)
	wg.Wait()

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
			if sceneTestResultDataMsg.PlanId == 0 {
				sceneTestResultDataMsg.PlanId = resultDataMsg.PlanId
			}
			if sceneTestResultDataMsg.PlanName == "" {
				sceneTestResultDataMsg.PlanName = resultDataMsg.PlanName
			}
			if sceneTestResultDataMsg.SceneId == 0 {
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
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanId == 0 {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanId = resultDataMsg.PlanId
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanName == "" {
				sceneTestResultDataMsg.Results[resultDataMsg.TargetId].PlanName = resultDataMsg.PlanName
			}
			if sceneTestResultDataMsg.Results[resultDataMsg.TargetId].SceneId == 0 {
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
