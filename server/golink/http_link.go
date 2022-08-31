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
func HttpSend(request model.Request, globalVariable *sync.Map, requestCollection *mongo.Collection, debugMsg *model.DebugMsg) (bool, int64, uint64, uint, uint, string) {
	var (
		isSucceed     = true
		errCode       = model.NoError
		contentLength = uint(0)
		errMsg        = ""
	)

	resp, req, requestTime, sendBytes, err := client.HTTPRequest(request.Method, request.URL, request.Body,
		request.Headers, request.Timeout)
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	defer fasthttp.ReleaseRequest(req)
	if request.Regulars != nil {
		for _, regular := range request.Regulars {
			regular.Extract(string(resp.Body()), globalVariable)
		}
	}

	if err != nil {
		isSucceed = false
		errCode = int64(model.RequestError) // 请求错误
		errMsg = err.Error()
	} else {
		// 断言验证
		if request.Assertions != nil {

			var assertionMsgList []model.AssertionMsg
			var assertionMsg = model.AssertionMsg{}
			var (
				code    = int64(10000)
				succeed = true
				msg     = ""
			)
			for k, v := range request.Assertions {
				switch request.Assertions[k].Type {
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

	if request.Debug == true {
		if debugMsg == nil {
			debugMsg = &model.DebugMsg{}
		}

		debugMsg.Request = make(map[string]string)
		debugMsg.Request["header"] = req.Header.String()
		debugMsg.Request["body"] = string(req.Body())
		debugMsg.Response = make(map[string]string)
		debugMsg.Response["header"] = resp.Header.String()
		debugMsg.Response["body"] = string(resp.Body())

		msg := make(map[string]*model.DebugMsg)
		msg["debug"] = debugMsg
		log.Logger.Info(request.ApiId)
		if requestCollection != nil {
			model.Insert(requestCollection, msg)
		}

	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg
}
