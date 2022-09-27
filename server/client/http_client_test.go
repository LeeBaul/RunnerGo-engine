package client

import (
	"fmt"
	"path"
	"testing"
)

func TestHTTPRequest(t *testing.T) {

	//method := "GET"
	//url := "http://www.baidu.com"
	//client := fastClient(1000, nil)
	//req := fasthttp.AcquireRequest()
	//resp := fasthttp.AcquireResponse()
	//req.Header.SetMethod(method)
	//bodyBuf := &bytes.Buffer{}
	//bodyWriter := multipart.FileHeader{}(bodyBuf)
	//_ = bodyWriter.WriteField("param", string([]byte("")))
	//writer := fasthttp.StreamWriter()
	//req
	//req.SetRequestURI(url)
	//if err := client.Do(req, resp); err != nil {
	//	fmt.Println("请求错误", err)
	//}

	fmt.Println(path.Base("D://name.txt"))

}
