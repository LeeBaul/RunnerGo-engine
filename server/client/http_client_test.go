package client

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"testing"
)

func TestHTTPRequest(t *testing.T) {

	method := "GET"
	url := "http://www.baidu.com"
	client := fastClient(1000, nil)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.Header.SetMethod(method)
	req.SetRequestURI(url)
	if err := client.Do(req, resp); err != nil {
		fmt.Println("请求错误", err)
	}

}
