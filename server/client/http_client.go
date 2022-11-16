package client

import (
	"RunnerGo-engine/config"
	"RunnerGo-engine/model"
	"crypto/tls"
	"github.com/valyala/fasthttp"
	"strings"
	"time"
)

func HTTPRequest(method, url string, body *model.Body, query *model.Query, header *model.Header, auth *model.Auth, timeout int64) (resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, sendBytes float64, err error, str string) {

	client := fastClient(timeout)
	req = fasthttp.AcquireRequest()

	// set methon
	req.Header.SetMethod(method)

	// set header
	header.SetHeader(req)

	urls := strings.Split(url, "//")
	if !strings.EqualFold(urls[0], model.HTTP) && !strings.EqualFold(urls[0], model.HTTPS) {
		url = model.HTTP + "//" + url

	}

	urlQuery := req.URI().QueryArgs()

	if strings.Contains(url, "?") && strings.Contains(url, "&") && strings.Contains(url, "=") {
		strs := strings.Split(url, "?")
		url = strs[0]
		queryList := strings.Split(strs[1], "&")
		for i := 0; i < len(queryList); i++ {
			keys := strings.Split(queryList[i], "=")
			urlQuery.Add(keys[0], keys[1])
		}
	}
	if query != nil && query.Parameter != nil {
		for _, v := range query.Parameter {
			if v.IsChecked == 1 {
				by := v.ValueToByte()
				urlQuery.AddBytesV(v.Key, by)
			}
		}
	}

	// set url
	req.SetRequestURI(url)

	// set body
	str = body.SetBody(req)

	// set auth
	auth.SetAuth(req)

	resp = fasthttp.AcquireResponse()

	startTime := time.Now()

	// 发送请求
	err = client.Do(req, resp)
	requestTime = uint64(time.Since(startTime))
	sendBytes = float64(req.Header.ContentLength()) / 1024
	if sendBytes <= 0 {
		sendBytes = float64(len(req.Body())) / 1024
	}
	return
}

// 获取fasthttp客户端
func fastClient(timeOut int64) *fasthttp.Client {
	fc := &fasthttp.Client{
		Name:                     config.Conf.Http.Name,
		NoDefaultUserAgentHeader: config.Conf.Http.NoDefaultUserAgentHeader,
		TLSConfig:                &tls.Config{InsecureSkipVerify: true},
		MaxConnsPerHost:          config.Conf.Http.MaxConnPerHost,
		MaxIdleConnDuration:      config.Conf.Http.MaxIdleConnDuration * time.Millisecond,
		ReadTimeout:              time.Duration(timeOut) * time.Millisecond,
		WriteTimeout:             time.Duration(timeOut) * time.Millisecond,
		MaxConnWaitTimeout:       config.Conf.Http.MaxConnWaitTimeout * time.Millisecond,
	}
	return fc

}
