// Package golink 连接
package golink

import (
	"kp-runner/model"
	"kp-runner/server/client"
)

func webSocketSend(request model.Request) (bool, int64, uint64, uint, uint) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = model.NoError
		contentLength = uint(0)
	)
	headers := map[string][]string{}
	for key, value := range request.Headers {
		headers[key] = []string{value}
	}
	resp, requestTime, sendBytes, err := client.WebSocketRequest(request.URL, request.Body, headers, int(request.Timeout))

	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		contentLength = uint(len(resp))
	}
	return isSucceed, int64(errCode), requestTime, sendBytes, contentLength
}
