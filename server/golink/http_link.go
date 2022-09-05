// Package golink 连接
package golink

import (
	"github.com/valyala/fasthttp"
	"go.mongodb.org/mongo-driver/mongo"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/server/client"
	"sync"
)

// HttpSend 发送http请求
func HttpSend(eventId string, api model.Api, globalVariable *sync.Map, requestCollection *mongo.Collection, debugMsg *model.DebugMsg) (bool, int64, uint64, uint, uint, string) {
	var (
		isSucceed     = true
		errCode       = model.NoError
		contentLength = uint(0)
		errMsg        = ""
	)

	resp, req, requestTime, sendBytes, err := client.HTTPRequest(api.Method, api.URL, api.Body, api.Query,
		api.Header, api.Auth, api.Timeout)
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	defer fasthttp.ReleaseRequest(req)
	if api.Regulars != nil {
		for _, regular := range api.Regulars {
			regular.Extract(string(resp.Body()), globalVariable)
		}
	}
	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
		errMsg = err.Error()
	} else {
		// 断言验证
		if api.Assertions != nil {

			var assertionMsgList []model.AssertionMsg
			var assertionMsg = model.AssertionMsg{}
			var (
				code    = int64(10000)
				succeed = true
				msg     = ""
			)
			for k, v := range api.Assertions {
				switch api.Assertions[k].Type {
				case model.Text:
					assert := v.AssertionText
					code, succeed, msg = assert.VerifyAssertionText(resp)

				case model.Regular:
				case model.Json:
				case model.XPath:

				}
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
	}
	// 开启debug模式后，将请求响应信息写入到mongodb中

	if api.Debug == true {
		if debugMsg == nil {
			debugMsg = &model.DebugMsg{}
		}
		debugMsg.EventId = eventId
		debugMsg.ApiId = api.TargetId
		debugMsg.ApiName = api.Name
		debugMsg.Request = make(map[string]string)
		debugMsg.Request["header"] = req.Header.String()
		debugMsg.Request["body"] = string(req.Body())
		debugMsg.Response = make(map[string]string)
		debugMsg.Response["header"] = resp.Header.String()
		debugMsg.Response["body"] = string(resp.Body())

		msg := make(map[string]*model.DebugMsg)
		msg["debug"] = debugMsg
		log.Logger.Info(api.TargetId)
		if requestCollection != nil {
			model.Insert(requestCollection, msg)
		}

	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg
}
