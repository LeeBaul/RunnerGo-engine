// Package golink 连接
package golink

import (
	request2 "kp-runner/model/request"

	"kp-runner/model"
	"kp-runner/server/client"
)

// httpSend 发送http请求
func httpSend(request request2.Request) (bool, int, uint64, int, int64) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = model.NoError
		contentLength = int64(0)
	)
	resp, requestTime, sendBytes, err := client.HTTPRequest(request.Method, request.URL, request.GetBody(),
		request.Headers, request.Timeout)

	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		contentLength = int64(resp.Header.ContentLength())
	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength
}
