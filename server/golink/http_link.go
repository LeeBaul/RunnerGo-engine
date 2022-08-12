// Package golink 连接
package golink

import (
	"kp-runner/model"
	"kp-runner/server/client"
	"kp-runner/tools"
)

// httpSend 发送http请求
func httpSend(request model.Request, globalVariable map[string]string) (bool, int, uint64, int, int64) {
	var (
		// startTime = time.Now()
		isSucceed     = true
		errCode       = model.NoError
		contentLength = int64(0)
	)
	// 如果接口中没有定义全局（接口）变量
	if request.Parameterizes == nil {
		request.Parameterizes = make(map[string]string)
	}
	for k, v := range request.Parameterizes {
		key := tools.VariablesMatch(k)
		if value, ok := globalVariable[key]; ok {
			request.Parameterizes[k] = value
		}

		value := tools.VariablesMatch(v)
		if value1, ok := globalVariable[value]; ok {
			request.Parameterizes[v] = value1
		}
	}
	resp, requestTime, sendBytes, err := client.HTTPRequest(request.Parameterizes, request.Method, request.URL, request.GetBody(),
		request.Headers, request.Timeout)
	if request.Regulars != nil {
		for _, regular := range request.Regulars {
			regular.Extract(resp.String(), globalVariable)
		}
	}
	if err != nil {
		isSucceed = false
		errCode = model.RequestError // 请求错误
	} else {
		// 接收到的字节长度
		contentLength = int64(resp.Header.ContentLength())
	}
	return isSucceed, errCode, requestTime, sendBytes, contentLength
}
