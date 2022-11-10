// Package server 压测启动
package server

import (
	"RunnerGo-engine/config"
	"RunnerGo-engine/log"
	"RunnerGo-engine/model"
	"RunnerGo-engine/server/execution"
	"RunnerGo-engine/server/golink"
	"RunnerGo-engine/server/heartbeat"
	"RunnerGo-engine/tools"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/mongo"
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
	if plan.ReportId == "" {
		log.Logger.Error("reportId 不能为空")
		return
	}
	topic := config.Conf.Kafka.TopIc
	partition := plan.Partition
	go model.SendKafkaMsg(kafkaProducer, resultDataMsgCh, topic, partition, plan.ReportId)
	var sharedMap = new(sync.Map) // 存储场景中各个事件的状态

	requestCollection := model.NewCollection(config.Conf.Mongo.DB, config.Conf.Mongo.StressDebugTable, mongoClient)
	debugCollection := model.NewCollection(config.Conf.Mongo.DB, config.Conf.Mongo.DebugTable, mongoClient)
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
				if value.Var == kv.Key {
					target = true
					continue
				}
			}
			if !target {
				var variable = new(model.KV)
				variable.Key = value.Var
				variable.Value = value.Val
				scene.Configuration.Variable = append(scene.Configuration.Variable, variable)
			}
		}
	}

	// 分解任务
	TaskDecomposition(plan, wg, resultDataMsgCh, debugCollection, requestCollection, sharedMap, kafkaProducer)
}

// TaskDecomposition 分解任务
func TaskDecomposition(plan *model.Plan, wg *sync.WaitGroup, resultDataMsgCh chan *model.ResultDataMsg, debugCollection, mongoCollection *mongo.Collection, sharedMap *sync.Map, kafkaProducer sarama.SyncProducer) {
	defer close(resultDataMsgCh)
	scene := plan.Scene
	scene.TeamId = plan.TeamId
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
		//teamId := strconv.FormatInt(plan.TeamId, 10)
		//p.DownLoadFile(teamId, plan.ReportId)
		p.UseFile()
	}

	var reportMsg = &model.ResultDataMsg{}
	if plan.MachineNum <= 0 {
		plan.MachineNum = 1
	}
	reportMsg.TeamId = plan.TeamId
	reportMsg.PlanId = plan.PlanId
	reportMsg.SceneId = scene.SceneId
	reportMsg.SceneName = scene.SceneName
	reportMsg.PlanName = plan.PlanName
	reportMsg.ReportId = plan.ReportId
	reportMsg.ReportName = plan.ReportName
	reportMsg.MachineNum = plan.MachineNum
	reportMsg.MachineIp = heartbeat.LocalIp + fmt.Sprintf("_%d", config.Conf.Heartbeat.Port)
	testModelJson, _ := json.Marshal(scene.ConfigTask.ModeConf)

	var startMsg = &model.ResultDataMsg{}
	startMsg.TeamId = plan.TeamId
	startMsg.PlanId = plan.PlanId
	startMsg.SceneId = scene.SceneId
	startMsg.SceneName = scene.SceneName
	startMsg.PlanName = plan.PlanName
	startMsg.ReportId = plan.ReportId
	startMsg.ReportName = plan.ReportName
	startMsg.MachineNum = plan.MachineNum
	startMsg.Timestamp = time.Now().UnixMilli()
	startMsg.Start = true
	log.Logger.Info("任务配置：    ", string(testModelJson))
	resultDataMsgCh <- startMsg
	var msg string
	switch scene.ConfigTask.Mode {
	case model.ConcurrentModel:
		execution.ConcurrentModel(wg, scene, reportMsg, resultDataMsgCh, debugCollection, mongoCollection, sharedMap)
	case model.ErrorRateModel:
		execution.ErrorRateModel(wg, scene, reportMsg, resultDataMsgCh, debugCollection, mongoCollection, sharedMap)
	case model.LadderModel:
		execution.LadderModel(wg, scene, reportMsg, resultDataMsgCh, debugCollection, mongoCollection, sharedMap)
		//case task.QpsModel:
		//	execution.ExecutionQpsModel()
	case model.RTModel:
		msg = execution.RTModel(wg, scene, reportMsg, resultDataMsgCh, debugCollection, mongoCollection, sharedMap)
	case model.QpsModel:
		execution.QPSModel(wg, scene, reportMsg, resultDataMsgCh, debugCollection, mongoCollection, sharedMap)
	default:
		var machines []string
		msg = "任务类型不存在"
		machine := reportMsg.MachineIp
		machines = append(machines, machine)
		tools.SendStopStressReport(machines, plan.ReportId)
	}
	wg.Wait()

	// 发送结束消息时间戳
	startMsg.Start = false
	startMsg.End = true
	startMsg.Timestamp = time.Now().UnixMilli()
	resultDataMsgCh <- startMsg
	debugMsg := make(map[string]interface{})
	debugMsg["report_id"] = plan.ReportId
	debugMsg["end"] = true
	model.Insert(mongoCollection, debugMsg)

	log.Logger.Info("计划:", plan.PlanId, "  ： ", msg)

}

// DebugScene 场景调试
func DebugScene(scene *model.Scene) {
	gid := tools.GetGid()
	wg := &sync.WaitGroup{}
	currentWg := &sync.WaitGroup{}
	mongoClient, err := model.NewMongoClient(
		config.Conf.Mongo.User,
		config.Conf.Mongo.Password,
		config.Conf.Mongo.Address,
		config.Conf.Mongo.DB)
	if err != nil {
		log.Logger.Error("连接mongo错误：", err)
		return
	}
	if scene.Variable != nil {
		if scene.Configuration == nil {
			scene.Configuration = new(model.Configuration)
		}
		if scene.Configuration.Variable == nil {
			scene.Configuration.Variable = []*model.KV{}
		}
		for _, v := range scene.Variable {
			target := false
			for _, sv := range scene.Configuration.Variable {
				if v.Key == sv.Key {
					target = true
				}
			}
			if !target {
				scene.Configuration.Variable = append(scene.Configuration.Variable, v)
			}
		}
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
		//teamId := strconv.FormatInt(scene.TeamId, 10)
		//p.DownLoadFile(teamId, scene.ReportId)
		p.UseFile()
	}
	var sharedMap = new(sync.Map)
	scene.Debug = model.All
	defer mongoClient.Disconnect(context.TODO())
	mongoCollection := model.NewCollection(config.Conf.Mongo.DB, config.Conf.Mongo.SceneDebugTable, mongoClient)
	golink.DisposeScene(sharedMap, wg, currentWg, gid, model.SceneType, scene, nil, nil, mongoCollection)
	currentWg.Wait()
	log.Logger.Info("场景：    ", scene.SceneName, "        调试结束！")

}

// DebugApi api调试
func DebugApi(debugApi model.Api) {

	if debugApi.Variable != nil && len(debugApi.Variable) > 0 {
		for _, value := range debugApi.Variable {
			if debugApi.Parameters == nil {
				debugApi.Parameters = new(sync.Map)
			}
			debugApi.Parameters.Store(value.Key, value.Value)
		}
	}
	event := model.Event{}
	event.Api = debugApi
	event.Weight = 100
	event.Id = "接口调试"
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
	golink.DisposeRequest(nil, nil, nil, configuration, event, mongoCollection)

}
