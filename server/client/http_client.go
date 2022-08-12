// Package client http 客户端
package client

import (
	"crypto/tls"
	"encoding/json"
	"github.com/valyala/fasthttp"
	"kp-runner/config"
	"kp-runner/log"
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

func HTTPRequest(parameterizes map[string]string, method, url string, body []byte, headers map[string]string, timeout int) (resp *fasthttp.Response, requestTime uint64, sendBytes int, err error) {

	client := fastClient(timeout)
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(method)
	if headers != nil {
		str, _ := json.Marshal(headers)
		req.Header.SetContentEncodingBytes(str)
	}

	for k, v := range headers {
		if contentType, ok := headers["Content-Type"]; ok {
			req.Header.SetContentEncoding(contentType)
		}
		// 查找header的key中是否存在变量{{****}}
		key := tools.VariablesMatch(k)
		// 如果header的key和提取出来的值不相等说明存在{{**}}， 那么我们需要将{{***}}替换为真实的值
		if strings.Compare(key, k) != 0 {
			// 如果在接口定义的变量中存在这个变量，那么我们将变量替换成实际的值
			if value, ok := parameterizes[key]; ok {
				// 将原来带变量表达式的str，替换成真正的str
				realVar := strings.Replace(k, "{{"+key+"}}", value, -1)
				headers[realVar] = headers[k]
				delete(headers, k)
			}
		}
		value := tools.VariablesMatch(v)
		if strings.Compare(value, v) != 0 {
			if value1, ok := parameterizes[value]; ok {
				realVar := strings.Replace(v, "{{"+value+"}}", value1, -1)
				headers[k] = realVar
			}
		}
	}
	urlVar := tools.VariablesMatch(url)
	if strings.Compare(url, urlVar) != 0 {
		if value, ok := parameterizes[urlVar]; ok {
			url = strings.Replace(url, "{{"+urlVar+"}}", value, -1)
		}
	}
	realBody := string(body)
	bodyVar := tools.VariablesMatch(realBody)
	if strings.Compare(bodyVar, urlVar) != 0 {
		if value, ok := parameterizes[urlVar]; ok {
			realBody = strings.Replace(url, "{{"+bodyVar+"}}", value, -1)
		}
	}
	body = []byte(realBody)

	req.SetRequestURI(url)
	req.SetBody(body)

	sendBytes = req.Header.ContentLength()
	resp = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	startTime := time.Now().UnixMilli()
	if err = client.Do(req, resp); err != nil {
		log.Logger.Error("请求错误", err)
	}
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
