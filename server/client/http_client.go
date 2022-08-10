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

func HTTPRequest(method, url string, body []byte, headers map[string]string, timeout time.Duration) (resp *fasthttp.Response, requestTime uint64, sendBytes int, err error) {

	client := fastClient(timeout)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(method)
	if headers != nil {
		str, _ := json.Marshal(headers)
		req.Header.SetContentEncodingBytes(str)
	}
	fmt.Println(111111111111)
	switch method {
	case "POST":
		req.SetRequestURI(url)
		req.SetBody(body)
	default:
		req.SetRequestURI(url + "?" + string(body))
	}
	sendBytes = req.Header.ContentLength()
	resp = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	startTime := time.Now().UnixMilli()
	if err = client.Do(req, resp); err != nil {
		log.Logger.Error("请求错误", err)
	}
	fmt.Println("req: ", string(req.Body()))
	fmt.Println("resp: ", string(resp.Body()))
	requestTime = tools.TimeDifference(startTime)

	return

}

// 获取fasthttp客户端
func fastClient(timeOut time.Duration) *fasthttp.Client {
	return &fasthttp.Client{
		Name:                     config.Config["httpClientName"].(string),
		NoDefaultUserAgentHeader: config.Config["httpNoDefaultUserAgentHeader"].(bool),
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          config.Config["httpClientMaxConnsPerHost"].(int),
		MaxIdleConnDuration:      config.Config["httpClientMaxIdleConnDuration"].(time.Duration),
		ReadTimeout:              timeOut,
		WriteTimeout:             config.Config["httpClientWriteTimeout"].(time.Duration),
		MaxConnWaitTimeout:       config.Config["httpClientMaxConnWaitTimeout"].(time.Duration),
	}
}
