// Package client http 客户端
package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"kp-runner/config"
	"kp-runner/helper"
	"log"
	"os"
	"time"
)

// logErr err
var logErr = log.New(os.Stderr, "", 0)

// HTTPRequest HTTP 请求
// method 方法 GET POST
// url 请求的url
// body 请求的body
// headers 请求头信息
// timeout 请求超时时间
//func HTTPRequest(method, url string, body []byte, headers map[string]string,
//	timeout time.Duration) (resp *fasthttp.Response, requestTime uint64, err error) {
//	// 跳过证书验证
//	client := fastClient(timeout)
//	req := &fasthttp.Request{}
//	req.Header.SetMethod(method)
//	req.SetRequestURI(url)
//	req.SetBody(body)
//
//	// 在req中设置Host，解决在header中设置Host不生效问题
//	// 设置默认为utf-8编码
//	if _, ok := headers["Content-Type"]; !ok {
//		if headers == nil {
//			headers = make(map[string]string)
//		}
//		headers["Content-Type"] = "application/x-www-form-urlencoded; charset=utf-8"
//	}
//	for key, value := range headers {
//		req.Header.Set(key, value)
//	}
//	startTime := time.Now()
//	err = client.Do(req, resp)
//	requestTime = uint64(helper.DiffNano(startTime))
//	if err != nil {
//		logErr.Println("请求失败:", err)
//		return
//	}
//	return
//}
//

//func HTTPRequest(method, url string, body []byte, headers map[string]string,
//	timeout time.Duration) (resp *fasthttp.Response, requestTime uint64, err error) {
//	client := fastClient(timeout)
//
//}

// FastGet 获取GET请求对象,没有进行资源回收
// @Description:
// @param url
// @param args
// @return *fasthttp.Request
func FastGet(url string, args map[string]interface{}) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	values := helper.ToString(args)
	req.SetRequestURI(url + "?" + values)
	return req
}

// DoGet 发送GET请求,获取响应
// @Description:
// @param url
// @param args
// @return []byte
// @return error
func DoGet(fastClient *fasthttp.Client, url string, args map[string]interface{}) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req) // 用完需要释放资源
	req.Header.SetMethod("GET")
	values := helper.ToString(args)
	req.SetRequestURI(url + "?" + values)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	if err := fastClient.Do(req, resp); err != nil {
		fmt.Println("请求失败:", err.Error())
		return nil, err
	}
	return resp.Body(), nil
}

// FastPostJson POST请求JSON参数,没有进行资源回收
// @Description:
// @param url
// @param args
// @return *fasthttp.Request
func FastPostJson(url string, args map[string]interface{}) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	// 默认是application/x-www-form-urlencoded
	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")
	req.SetRequestURI(url)
	marshal, _ := json.Marshal(args)
	req.SetBody(marshal)
	return req
}

// FastPostForm POST请求表单传参,没有进行资源回收
// @Description:
// @param url
// @param args
// @return *fasthttp.Request
func FastPostForm(url string, args map[string]interface{}) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	// 默认是application/x-www-form-urlencoded
	//req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")
	req.SetRequestURI(url)
	marshal, _ := json.Marshal(args)
	req.BodyWriter().Write([]byte(helper.ToString(args)))
	req.BodyWriter().Write(marshal)
	return req
}

// FastResponse 获取响应,保证资源回收
// @Description:
// @param request
// @return []byte
// @return error
func FastResponse(fastClient *fasthttp.Client, request *fasthttp.Request) ([]byte, error) {
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	defer fasthttp.ReleaseRequest(request)
	if err := fastClient.Do(request, response); err != nil {
		log.Println("响应出错了")
		return nil, err
	}
	return response.Body(), nil
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
