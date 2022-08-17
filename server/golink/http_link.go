// Package golink 连接
package golink

import (
	"github.com/valyala/fasthttp"
	"kp-runner/model"
	"kp-runner/server/client"
	"sync"
)

// httpSend 发送http请求
func httpSend(request *model.Request, globalVariable *sync.Map) (bool, int, uint64, int, int64, string) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = model.NoError
		contentLength = int64(0)
		errMsg        = ""
	)

	resp, requestTime, sendBytes, err := client.HTTPRequest(request.Method, request.URL, request.Body,
		request.Headers, request.Timeout)
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	if request.Regulars != nil {
		for _, regular := range request.Regulars {
			regular.Extract(string(resp.Body()), globalVariable)
		}
	}

	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
		errMsg = err.Error()
	} else {
		// 断言验证
		if request.Assertions != nil {
			for k, v := range request.Assertions {
				switch request.Assertions[k].Type {
				case model.Text:
					assert := v.AssertionText
					errCode, isSucceed, errMsg = assert.VerifyAssertionText(resp)
					if !isSucceed {
						break
					}
				case model.Regular:
				case model.Json:
				case model.XPath:

				}
			}
		}
		// 接收到的字节长度
		contentLength = int64(resp.Header.ContentLength())
	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength, errMsg
}
