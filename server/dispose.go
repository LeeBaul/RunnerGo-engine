// Package server 压测启动
package server

import (
	"kp-runner/config"
	"kp-runner/model"
	"kp-runner/server/execution"
	"sync"
)

const (
	connectionMode = 1 // 1:顺序建立长链接 2:并发建立长链接

)

// init 注册验证器
func init() {

	// http
	//execution.RegisterVerifyHTTP("statusCode", verify.HTTPStatusCode)
	//execution.RegisterVerifyHTTP("json", verify.HTTPJson)

	// webSocket
	//execution.RegisterVerifyWebSocket("json", verify.WebSocketJSON)
}

// Execution 执行计划
func Execution(plan model.Plan) {
	// 设置kafka消费者
	kafkaProducer := model.NewKafkaProducer([]string{config.Config["kafkaAddress"].(string)})
	wg := &sync.WaitGroup{}
	// 设置接收数据缓存
	ch := make(chan *model.TestResultDataMsg, 10000)
	go model.SendKafkaMsg(kafkaProducer, ch)
	defer close(ch)
	switch plan.ConfigTask.TestModel.Type {
	case model.ConcurrentModel:
		execution.ExecutionConcurrentModel(
			kafkaProducer,
			wg,
			ch,
			plan)
	case model.ErrorRateModel:
		execution.ExecutionErrorRateModel(
			kafkaProducer,
			wg,
			plan,
			ch)
	case model.LadderModel:
		execution.ExecutionLadderModel(kafkaProducer,
			wg,
			plan,
			ch)
		//case task.TpsModel:
		//	execution.ExecutionTpsModel()
		//case task.QpsModel:
		//	execution.ExecutionQpsModel()
	case model.RTModel:
		execution.ExecutionRTModel(kafkaProducer,
			wg,
			plan,
			ch)
	default:

	}
	wg.Wait()
}
