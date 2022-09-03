// Package client http 客户端
package client

import (
	"crypto/tls"
	"github.com/valyala/fasthttp"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/model"
	"kp-runner/tools"
	"strings"
	"time"
)

// HTTPRequest HTTP 请求
// method 方法 GET POST
// url 请求的url
// body 请求的body
// headers 请求头信息
// timeout 请求超时时间

func HTTPRequest(method, url string, body string, query, header []model.VarForm, timeout int64) (resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, sendBytes uint, err error) {

	client := fastClient(timeout)
	req = fasthttp.AcquireRequest()

	req.Header.SetMethod(method)
	if header != nil {
		for _, v := range header {
			if v.Enable == true {
				if strings.EqualFold(v.Name, "content-type") {
					req.Header.SetContentType(v.Value.(string))
				}
				if strings.EqualFold(v.Name, "host") {
					req.Header.SetHost(v.Value.(string))
				}
				req.Header.Set(v.Name, v.Value.(string))
			}
		}

	}

	if method == "GET" {
		if query != nil {
			for _, v := range query {
				if v.Enable == true {
					url += "?" + v.Value.(string)
				}
			}
		}
	}
	req.SetRequestURI(url)

	req.SetBodyString(body)

	resp = fasthttp.AcquireResponse()

	startTime := time.Now().UnixMilli()
	if err = client.Do(req, resp); err != nil {
		log.Logger.Error("请求错误", err)
	}
	requestTime = tools.TimeDifference(startTime)
	sendBytes = uint(req.Header.ContentLength())
	log.Logger.Info("req", string(req.Body()))
	return
}

// 获取fasthttp客户端
func fastClient(timeOut int64) *fasthttp.Client {
	fc := &fasthttp.Client{
		Name:                     config.Config["httpClientName"].(string),
		NoDefaultUserAgentHeader: config.Config["httpNoDefaultUserAgentHeader"].(bool),
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          int(config.Config["httpClientMaxConnsPerHost"].(int64)),
		MaxIdleConnDuration:      time.Duration(config.Config["httpClientMaxIdleConnDuration"].(int64)) * time.Millisecond,
		ReadTimeout:              time.Duration(timeOut) * time.Millisecond,
		WriteTimeout:             time.Duration(config.Config["httpClientWriteTimeout"].(int64)) * time.Millisecond,
		MaxConnWaitTimeout:       time.Duration(config.Config["httpClientMaxConnWaitTimeout"].(int64)) * time.Millisecond,
	}
	return fc

}
