// Package client http 客户端
package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/tools"
	"time"
)

// HTTPRequest HTTP 请求
// method 方法 GET POST
// url 请求的url
// body 请求的body
// headers 请求头信息
// timeout 请求超时时间

func HTTPRequest(method, url string, body []byte, headers map[string]string, timeout int) (resp *fasthttp.Response, requestTime uint64, sendBytes int, err error) {

	client := fastClient(timeout)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(method)
	if headers != nil {
		str, _ := json.Marshal(headers)
		req.Header.SetContentEncodingBytes(str)
	}
	switch method {
	case "POST":
		req.SetRequestURI(url)
		req.SetBody(body)
	case "GET":
		req.SetRequestURI(url)
	}
	sendBytes = req.Header.ContentLength()
	resp = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	startTime := time.Now().UnixMilli()
	if err = client.Do(req, resp); err != nil {
		log.Logger.Error("请求错误", err)
	}
	fmt.Println("req:", req)
	fmt.Println("resp:", resp)
	requestTime = tools.TimeDifference(startTime)

	return

}

// 获取fasthttp客户端
func fastClient(timeOut int) *fasthttp.Client {
	return &fasthttp.Client{
		Name:                     config.Config["httpClientName"].(string),
		NoDefaultUserAgentHeader: config.Config["httpNoDefaultUserAgentHeader"].(bool),
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          int(config.Config["httpClientMaxConnsPerHost"].(int64)),
		MaxIdleConnDuration:      time.Duration(config.Config["httpClientMaxIdleConnDuration"].(int64)) * time.Millisecond,
		ReadTimeout:              time.Duration(int64(timeOut)) * time.Millisecond,
		WriteTimeout:             time.Duration(config.Config["httpClientWriteTimeout"].(int64)) * time.Millisecond,
		MaxConnWaitTimeout:       time.Duration(config.Config["httpClientMaxConnWaitTimeout"].(int64)) * time.Millisecond,
	}
}
