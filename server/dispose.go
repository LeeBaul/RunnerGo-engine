// Package server 压测启动
package server

import (
	"github.com/robfig/cron/v3"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/execution"
	"sort"
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

	status := model.QueryTimingTaskStatus(plan.PlanID + ":" + plan.Scene.SceneId + ":" + "timing")
	if status == false {
		c.Remove(id)
	}

}

// ExecutionPlan 执行计划
func ExecutionPlan(plan *model.Plan) {

	// 设置kafka消费者
	kafkaProducer := model.NewKafkaProducer([]string{config.Config["kafkaAddress"].(string)})
	// 设置接收数据缓存
	ch := make(chan *model.ResultDataMsg, 10000)
	// 任务状态channel，
	statusCh := make(chan bool, 1)
	// 查询计划状态

	sceneTestResultDataMsgCh := make(chan *model.SceneTestResultDataMsg, 1)
	go model.QueryPlanStatus(plan.PlanID+":"+plan.Scene.SceneId+":status", statusCh)

	// 计算测试结果
	go ReceivingResults(ch, sceneTestResultDataMsgCh)
	go model.SendKafkaMsg(kafkaProducer, sceneTestResultDataMsgCh)
	switch plan.ConfigTask.TestModel.Type {
	case model.ConcurrentModel:
		execution.ExecutionConcurrentModel(
			statusCh,
			ch,
			plan)
	case model.ErrorRateModel:
		execution.ExecutionErrorRateModel(
			statusCh,
			plan,
			ch)
	case model.LadderModel:
		execution.ExecutionLadderModel(
			statusCh,
			plan,
			ch)
		//case task.TpsModel:
		//	execution.ExecutionTpsModel()
		//case task.QpsModel:
		//	execution.ExecutionQpsModel()
	case model.RTModel:
		execution.ExecutionRTModel(
			statusCh,
			plan,
			ch)
	default:
		close(ch)
	}
}

// ReceivingResults 计算并发送测试结果
func ReceivingResults(resultDataMsgCh <-chan *model.ResultDataMsg, sceneTestResultDataMsgCh chan *model.SceneTestResultDataMsg) {
	var (
		sceneTestResultDataMsg = new(model.SceneTestResultDataMsg)
		requestTimeMap         = make(map[string]model.RequestTimeList)
	)
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
func timeLineCalculate(line int, requestTimeList model.RequestTimeList) (requestTime uint64) {
	if line > 0 && line < 100 {
		proportion := float64(line) / 100
		value := proportion * float64(len(requestTimeList))
		requestTime = requestTimeList[int(value)]
	}
	return

}
