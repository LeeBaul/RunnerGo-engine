package golink

import (
	"kp-runner/model"
	"strconv"
	"sync"
	"time"
)

func Dispose(eventList []model.Event, ch chan *model.ResultDataMsg, plan *model.Plan, globalVariable *sync.Map) {

	for _, event := range eventList {
		switch event.EventType {
		case model.RequestType:
			DisposeRequest(ch, plan, event.Request, globalVariable)
		case model.ControllerType:
			DisposeController(event.Controller, globalVariable)
		}
	}
}

func DisposeRequest(ch chan<- *model.ResultDataMsg, plan *model.Plan,
	request *model.Request, globalVariable *sync.Map) {

	requestResults := &model.ResultDataMsg{
		PlanId:     plan.PlanID,
		PlanName:   plan.PlanName,
		SceneId:    plan.Scene.SceneId,
		SceneName:  plan.Scene.SceneName,
		ReportId:   plan.ReportId,
		ReportName: plan.ReportName,
	}
	if request.Parameterizes == nil {
		request.Parameterizes = make(map[string]string)
	}
	request.ReplaceParameterizes(globalVariable)

	// 如果接口的变量中没有全局变量中的key，那么将全局变量添加到接口变量中
	globalVariable.Range(func(key, value any) bool {
		if _, ok := request.Parameterizes[key.(string)]; !ok {
			request.Parameterizes[key.(string)] = value.(string)
		}
		return true
	})

	// 将参数化中的数据赋值给请求中的变量里
	if plan.Scene.Configuration.ParameterizedFile.VariableNames != nil {
		for variableName, _ := range plan.Scene.Configuration.ParameterizedFile.VariableNames.VarMapList {
			if _, ok := request.Parameterizes[variableName]; !ok {
				plan.Scene.Configuration.Mu.Lock()
				p := plan.Scene.Configuration.ParameterizedFile
				request.Parameterizes[variableName] = p.UseVar(variableName)
				plan.Scene.Configuration.Mu.Unlock()
			}
		}
	}

	request.ReplaceBodyParameterizes()
	request.ReplaceHeaderParameterizes()
	request.ReplaceUrlParameterizes()

	switch request.Form {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg := httpSend(request, globalVariable)
		requestResults.ApiName = request.ApiName
		requestResults.RequestTime = requestTime
		requestResults.ErrorType = errCode
		requestResults.IsSucceed = isSucceed
		requestResults.SendBytes = int64(sendBytes)
		requestResults.ReceivedBytes = contentLength
		requestResults.ErrorMsg = errMsg
		ch <- requestResults
	case model.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, contentLength := webSocketSend(request)
		requestResults.ApiName = request.ApiName
		requestResults.RequestTime = requestTime
		requestResults.ErrorType = errCode
		requestResults.IsSucceed = isSucceed
		requestResults.SendBytes = int64(sendBytes)
		requestResults.ReceivedBytes = contentLength
		ch <- requestResults
	case model.FormTypeGRPC:
		//isSucceed, errCode, requestTime, sendBytes, contentLength := rpcSend(request)
		//requestResults.ApiName = request.ApiName
		//requestResults.RequestTime = requestTime
		//requestResults.ErrorType = errCode
		//requestResults.IsSucceed = isSucceed
		//requestResults.SendBytes = int64(sendBytes)
		//requestResults.ReceivedBytes = contentLength
		//ch <- requestResults

	}
	if request.Requests != nil && request.Requests[0] != nil {
		for _, requestIndividual := range request.Requests {
			DisposeRequest(ch, plan, requestIndividual, globalVariable)
		}
	}
	if request.Controllers != nil && request.Controllers[0] != nil {
		for _, controllerIndividual := range request.Controllers {
			DisposeController(controllerIndividual, globalVariable)
		}

	}
}

func DisposeController(controller *model.Controller, globalVariable *sync.Map) {
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
