package golink

import (
	"fmt"
	"kp-runner/model"
	"kp-runner/model/plan"
	request2 "kp-runner/model/request"
	"sync"
)

func Send(ch chan<- *model.TestResultDataMsg, plan plan.Plan, wg *sync.WaitGroup,
	request request2.Request) {
	defer wg.Done()
	// fmt.Printf("启动协程 编号:%05d \n", chanID)
	requestResults := &model.TestResultDataMsg{
		PlanId:     plan.PlanID,
		PlanName:   plan.PlanName,
		SceneId:    plan.Scene.SceneID,
		SceneName:  plan.Scene.SceneName,
		ReportId:   plan.ReportId,
		ReportName: plan.ReportName,
	}
	fmt.Println("requestResults", requestResults)
	switch request.Form {
	case request2.FormTypeHTTP:
		isSucceed, errCode, requestTime, sendBytes, contentLength := httpSend(request)
		requestResults.ApiName = request.ApiName
		requestResults.RequestTime = requestTime
		requestResults.ErrorType = errCode
		requestResults.IsSucceed = isSucceed
		requestResults.SendBytes = int64(sendBytes)
		requestResults.ReceivedBytes = contentLength
		ch <- requestResults
	case request2.FormTypeWebSocket:
		isSucceed, errCode, requestTime, sendBytes, contentLength := webSocketSend(request)
		requestResults.ApiName = request.ApiName
		requestResults.RequestTime = requestTime
		requestResults.ErrorType = errCode
		requestResults.IsSucceed = isSucceed
		requestResults.SendBytes = int64(sendBytes)
		requestResults.ReceivedBytes = contentLength
		ch <- requestResults
	case request2.FormTypeGRPC:
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
