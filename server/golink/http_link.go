// Package golink 连接
package golink

import (
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/model"
	"kp-runner/server/client"
	"sync"
)

// HttpSend 发送http请求
func HttpSend(event model.Event, api model.Api, sceneVariable *sync.Map, requestCollection *mongo.Collection) (bool, int64, uint64, uint, uint, string) {
	var (
		isSucceed     = true
		errCode       = model.NoError
		contentLength = uint(0)
		errMsg        = ""
	)

	resp, req, requestTime, sendBytes, err := client.HTTPRequest(api.Method, api.Request.URL, api.Request.Body, api.Request.Query,
		api.Request.Header, api.Request.Auth, api.Timeout)
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	defer fasthttp.ReleaseRequest(req)
	var regex []map[string]interface{}
	if api.Regex != nil {
		for _, regular := range api.Regex {
			reg := make(map[string]interface{})
			value := regular.Extract(string(resp.Body()), sceneVariable)
			reg[regular.Var] = value
			regex = append(regex, reg)
		}
	}
	if err != nil {
		isSucceed = false
		errMsg = err.Error()
	}
	var assertionMsgList []model.AssertionMsg
	// 断言验证
	if api.Assert != nil {
		var assertionMsg = model.AssertionMsg{}
		var (
			code    = int64(10000)
			succeed = true
			msg     = ""
		)
		for _, v := range api.Assert {
			code, succeed, msg = v.VerifyAssertionText(resp)
			if succeed != true {
				errCode = code
				isSucceed = succeed
				errMsg = msg
			}
			assertionMsg.Code = code
			assertionMsg.IsSucceed = succeed
			assertionMsg.Msg = msg
			assertionMsgList = append(assertionMsgList, assertionMsg)
		}
	}
	// 接收到的字节长度
	contentLength = uint(resp.Header.ContentLength())

	// 开启debug模式后，将请求响应信息写入到mongodb中
	if api.Debug == true {
		debugMsg := make(map[string]interface{})

		debugMsg["uuid"] = api.Uuid.String()
		debugMsg["event_id"] = event.Id
		debugMsg["api_id"] = api.TargetId
		debugMsg["api_name"] = api.Name
		debugMsg["request_time"] = requestTime
		debugMsg["request_code"] = resp.StatusCode()

		debugMsg["request_header"] = req.Header.String()
		debugMsg["request_body"] = string(req.Body())

		debugMsg["response_header"] = resp.Header.String()
		debugMsg["response_body"] = string(resp.Body())
		debugMsg["response_bytes"] = resp.Header.ContentLength()
		if err != nil {
			debugMsg["response_msg"] = err.Error()
		}
		switch isSucceed {
		case false:
			debugMsg["status"] = model.Failed
		case true:
			debugMsg["status"] = model.Success
		}

		debugMsg["next_list"] = event.NextList

		if api.Assert != nil {
			debugMsg["assertion"] = assertionMsgList
		}
		if api.Regex != nil {
			debugMsg["regex"] = regex
		}

		if requestCollection != nil {
			model.Insert(requestCollection, debugMsg)
		}
	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg
}
