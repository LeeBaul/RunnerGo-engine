// Package golink 连接
package golink

import (
	"kp-runner/model"
	"kp-runner/server/client"
)

func webSocketSend(api model.Api) (bool, int64, uint64, uint, uint) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = model.NoError
		contentLength = uint(0)
	)
	headers := map[string][]string{}
	for _, header := range api.Request.Header.Parameter {
		if header.IsChecked == 1 {
			headers[header.Key] = []string{header.Value.(string)}
		}

	}
	resp, requestTime, sendBytes, err := client.WebSocketRequest(api.Request.URL, api.Request.Body.ToString(), headers, int(api.Timeout))

	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		contentLength = uint(len(resp))
	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength
}
