package client

import (
	"crypto/tls"
	"github.com/valyala/fasthttp"
	"kp-runner/config"
	"kp-runner/log"
	"kp-runner/model"
	"strings"
	"time"
)

func HTTPRequest(method, url string, body *model.Body, query *model.Query, header *model.Header, auth *model.Auth, timeout int64) (resp *fasthttp.Response, req *fasthttp.Request, requestTime uint64, sendBytes float64, err error, timestamp int64, str string) {

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

	if method == "GET" {
		if query != nil && query.Parameter != nil {
			var temp []*model.VarForm
			for _, v := range query.Parameter {
				if v.IsChecked == 1 {
					if !strings.Contains(url, v.Key) {
						temp = append(temp, v)
					}
				}
			}
			for k, v := range temp {
				if k == 0 {
					url += "?" + v.Key + "=" + v.Value.(string)
				} else {
					url += "&" + v.Key + "=" + v.Value.(string)
				}
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
	if err = client.Do(req, resp); err != nil {
		log.Logger.Error("请求错误", err)
	}

	requestTime = uint64(time.Since(startTime))
	sendBytes = float64(req.Header.ContentLength()) / 1024
	timestamp = time.Now().UnixMilli()
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
