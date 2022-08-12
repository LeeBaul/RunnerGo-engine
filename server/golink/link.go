package golink

import (
	"kp-runner/model"
	"sync"
)

func Send(ch chan<- *model.TestResultDataMsg, plan model.Plan, wg *sync.WaitGroup,
	request model.Request, globalVariable map[string]string) {
	defer wg.Done()
	// fmt.Printf("启动协程 编号:%05d \n", chanID)
	requestResults := &model.TestResultDataMsg{
		PlanId:     plan.PlanID,
		PlanName:   plan.PlanName,
		SceneId:    plan.Scene.SceneId,
		SceneName:  plan.Scene.SceneName,
		ReportId:   plan.ReportId,
		ReportName: plan.ReportName,
	}

	switch request.Form {
	case model.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, contentLength := httpSend(request, globalVariable)
		requestResults.ApiName = request.ApiName
		requestResults.RequestTime = requestTime
		requestResults.ErrorType = errCode
		requestResults.IsSucceed = isSucceed
		requestResults.SendBytes = int64(sendBytes)
		requestResults.ReceivedBytes = contentLength
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

}
