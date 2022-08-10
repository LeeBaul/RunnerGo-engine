// Package golink 连接
package golink

import (
	"fmt"
	"kp-runner/model"
	request2 "kp-runner/model/request"
	"kp-runner/server/client"
)

func webSocketSend(request request2.Request) (bool, int, uint64, int, int64) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = model.NoError
		contentLength = int64(0)
	)
	headers := map[string][]string{}
	for key, value := range request.Headers {
		headers[key] = []string{value}
	}
	resp, requestTime, sendBytes, err := client.WebSocketRequest(request.URL, request.GetBody(), headers, request.Timeout)

	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		contentLength = int64(len(resp))
	}
	fmt.Println("resp", string(resp))
	return isSucceed, errCode, requestTime, sendBytes, contentLength
}
