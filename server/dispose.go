// Package server 压测启动
package server

import (
	"kp-runner/config"
	"kp-runner/model"
	"kp-runner/server/execution"
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
func Execution(plan *model.Plan) {

	// 设置kafka消费者
	kafkaProducer := model.NewKafkaProducer([]string{config.Config["kafkaAddress"].(string)})
	// 设置接收数据缓存
	ch := make(chan *model.ResultDataMsg, 10000)

	go model.SendKafkaMsg(kafkaProducer, ch)
	defer close(ch)
	switch plan.ConfigTask.TestModel.Type {
	case model.ConcurrentModel:
		execution.ExecutionConcurrentModel(
			kafkaProducer,
			ch,
			plan)
	case model.ErrorRateModel:
		execution.ExecutionErrorRateModel(
			kafkaProducer,
			plan,
			ch)
	case model.LadderModel:
		execution.ExecutionLadderModel(
			kafkaProducer,
			plan,
			ch)
		//case task.TpsModel:
		//	execution.ExecutionTpsModel()
		//case task.QpsModel:
		//	execution.ExecutionQpsModel()
	case model.RTModel:
		execution.ExecutionRTModel(kafkaProducer,
			plan,
			ch)
	default:

	}
}

// 计算测试结果
//
//func ReceivingResults(resultDataMsgCh <-chan *model.ResultDataMsg, apiTestResultDataMsgCh chan *model.ApiTestResultDataMsg, sceneTestResultDataMsgCh chan<- *model.SceneTestResultDataMsg) {
//	var (
//		sceneTestResultDataMsg = &model.SceneTestResultDataMsg{}
//		sceneMap               = make(map[string]tools.MyUint64List)
//		apiTestResultDataMsg   = &model.ApiTestResultDataMsg{}
//	)
//	for {
//		ticker := time.NewTicker(1 * time.Second)
//		select {
//		case resultDataMsg := <-resultDataMsgCh:
//			sceneMap[resultDataMsg.ApiId] = append(sceneMap[resultDataMsg.ApiId], resultDataMsg.RequestTime)
//			sort.Sort(sceneMap[resultDataMsg.ApiId])
//		case <-ticker.C:
//			for k, v := range sceneMap {
//				sort.Sort(v)
//				minRequestTime = v[0]
//				maxRequestTime = v[len(v)-1]
//
//			}
//		}
//
//	}
//}
