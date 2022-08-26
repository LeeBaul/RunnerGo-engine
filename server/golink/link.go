package golink

import (
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/model"
	"strconv"
	"sync"
	"time"
)

// DisposeScene 对场景进行处理
func DisposeScene(eventList []model.Event, ch chan *model.ResultDataMsg, plan *model.Plan, globalVariable *sync.Map, wg *sync.WaitGroup, requestCollection, responseCollection *mongo.Collection, options ...int64) {

	for _, event := range eventList {
		switch event.EventType {
		case model.RequestType:
			DisposeRequest(ch, plan, event.Request, globalVariable, wg, requestCollection, responseCollection, options[0], options[1])
		case model.ControllerType:
			DisposeController(event.Controller, globalVariable, requestCollection, responseCollection, options[0], options[1])
		}
	}
}

// DisposeRequest 开始对请求进行处理
func DisposeRequest(ch chan<- *model.ResultDataMsg, plan *model.Plan,
	request *model.Request, globalVariable *sync.Map, wg *sync.WaitGroup, requestCollection, responseCollection *mongo.Collection, options ...int64) {

	if request.Weight != 100 && request.Weight != 0 {
		proportion := options[1] / int64(100-request.Weight)
		if options[0]%proportion == 0 {
			return
		}

	}
	requestResults := &model.ResultDataMsg{
		PlanId:                plan.PlanID,
		PlanName:              plan.PlanName,
		SceneId:               plan.Scene.SceneId,
		SceneName:             plan.Scene.SceneName,
		ReportId:              plan.ReportId,
		ReportName:            plan.ReportName,
		ApiId:                 request.ApiId,
		ApiName:               request.ApiName,
		CustomRequestTimeLine: request.CustomRequestTime,
		ErrorThreshold:        request.ErrorThreshold,
	}
	if request.Parameterizes == nil {
		request.Parameterizes = &sync.Map{}
	}
	request.ReplaceParameterizes(globalVariable)

	// 如果接口的变量中没有全局变量中的key，那么将全局变量添加到接口变量中
	globalVariable.Range(func(key, value any) bool {
		if _, ok := request.Parameterizes.Load(key); !ok {
			request.Parameterizes.Store(key, value)
		}
		return true
	})

	// 将参数化中的数据赋值给请求中的变量里
	if plan.Scene.Configuration.ParameterizedFile.VariableNames != nil {
		for variableName, _ := range plan.Scene.Configuration.ParameterizedFile.VariableNames.VarMapList {
			if _, ok := request.Parameterizes.Load(variableName); !ok {
				plan.Scene.Configuration.Mu.Lock()
				p := plan.Scene.Configuration.ParameterizedFile
				request.Parameterizes.Store(variableName, p.UseVar(variableName))
				plan.Scene.Configuration.Mu.Unlock()
			}
		}
	}

	request.ReplaceBodyParameterizes()
	request.ReplaceHeaderParameterizes()
	request.ReplaceUrlParameterizes()

	var (
		isSucceed     = false
		errCode       = 0
		requestTime   = uint64(0)
		sendBytes     = uint(0)
		contentLength = uint(0)
		errMsg        = ""
	)
	switch request.Form {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg = httpSend(request, globalVariable, requestCollection, responseCollection)
	case model.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, contentLength = webSocketSend(request)
	case model.FormTypeGRPC:
		//isSucceed, errCode, requestTime, sendBytes, contentLength := rpcSend(request)
	default:
		return

	}
	requestResults.ApiName = request.ApiName
	requestResults.RequestTime = requestTime
	requestResults.ErrorType = errCode
	requestResults.IsSucceed = isSucceed
	requestResults.SendBytes = uint64(sendBytes)
	requestResults.ReceivedBytes = uint64(contentLength)
	requestResults.ErrorMsg = errMsg
	ch <- requestResults
	if request.Requests != nil && request.Requests[0] != nil {
		for _, requestIndividual := range request.Requests {
			wg.Add(1)
			go func(requestIndividual *model.Request) {
				DisposeRequest(ch, plan, requestIndividual, globalVariable, wg, requestCollection, responseCollection, options[0], options[1])
				wg.Done()
			}(requestIndividual)
		}
	}
	if request.Controllers != nil && request.Controllers[0] != nil {
		for _, controllerIndividual := range request.Controllers {
			wg.Add(1)
			go func(controllerIndividual *model.Controller) {
				DisposeController(controllerIndividual, globalVariable, requestCollection, responseCollection, options[0], options[1])
			}(controllerIndividual)

		}
	}
}

// DisposeController 处理控制器
func DisposeController(controller *model.Controller, globalVariable *sync.Map, requestCollection, responseCollection *mongo.Collection, options ...int64) {
	switch controller.ControllerType {
	case model.IfControllerType:
		if v, ok := globalVariable.Load(controller.IfController.Key); ok {
			controller.IfController.PerForm(v.(string))
		}
	case model.CollectionType:
		// 集合点, 待开发
	case model.WaitControllerType: // 等待控制器
		timeWait, _ := strconv.Atoi(controller.WaitController.WaitTime)
		time.Sleep(time.Duration(timeWait) * time.Millisecond)
	}
}
