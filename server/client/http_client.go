// Package client http 客户端
package client

import (
	"crypto/tls"
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

func HTTPRequest(method, url string, body string, headers map[string]string, timeout int) (resp *fasthttp.Response, requestTime uint64, sendBytes int, err error) {

	client := fastClient(timeout)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(method)
	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}

	}

	req.SetRequestURI(url)
	req.SetBodyString(body)

	sendBytes = req.Header.ContentLength()
	resp = fasthttp.AcquireResponse()

	startTime := time.Now().UnixMilli()
	if err = client.Do(req, resp); err != nil {
		log.Logger.Error("请求错误", err)
	}
	requestTime = tools.TimeDifference(startTime)

	return
}

// 获取fasthttp客户端
func fastClient(timeOut int) *fasthttp.Client {
	fc := &fasthttp.Client{
		Name:                     config.Config["httpClientName"].(string),
		NoDefaultUserAgentHeader: config.Config["httpNoDefaultUserAgentHeader"].(bool),
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          int(config.Config["httpClientMaxConnsPerHost"].(int64)),
		MaxIdleConnDuration:      time.Duration(config.Config["httpClientMaxIdleConnDuration"].(int64)) * time.Millisecond,
		ReadTimeout:              time.Duration(int64(timeOut)) * time.Millisecond,
		WriteTimeout:             time.Duration(config.Config["httpClientWriteTimeout"].(int64)) * time.Millisecond,
		MaxConnWaitTimeout:       time.Duration(config.Config["httpClientMaxConnWaitTimeout"].(int64)) * time.Millisecond,
	}
	log.Logger.Info("MaxConnWaitTimeout", fc.MaxConnWaitTimeout)
	return fc

}
