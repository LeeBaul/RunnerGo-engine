// Package golink 连接
package golink

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/client"
	"sync"
)

// HttpSend 发送http请求
func HttpSend(eventId string, api model.Api, sceneVariable *sync.Map, requestCollection *mongo.Collection) (bool, int64, uint64, uint, uint, string) {
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
		debugMsg := new(model.DebugMsg)
		debugMsg.EventId = eventId
		debugMsg.ApiId = api.TargetId
		debugMsg.ApiName = api.Name
		debugMsg.Request = make(map[string]interface{})
		requestHeader := make(map[string]interface{})
		_ = json.Unmarshal(req.Header.Header(), &requestHeader)
		debugMsg.Request["header"] = requestHeader

		debugMsg.Request["body"] = string(req.Body())

		debugMsg.Response = make(map[string]interface{})
		responseHeader := make(map[string]interface{})
		_ = json.Unmarshal(req.Header.Header(), &responseHeader)
		debugMsg.Response["header"] = responseHeader

		debugMsg.Response["body"] = string(resp.Body())

		if api.Assert != nil {
			debugMsg.Assertion = make(map[string][]model.AssertionMsg)
			debugMsg.Assertion["assertion"] = assertionMsgList
		}
		if api.Regex != nil {
			debugMsg.Regex = regex
		}
		msg := make(map[string]*model.DebugMsg)
		msg["debug"] = debugMsg
		log.Logger.Info(api.TargetId)
		if requestCollection != nil {
			model.Insert(requestCollection, msg)
		}
	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg
}
