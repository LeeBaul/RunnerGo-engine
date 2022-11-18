package model

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
	"testing"
	"time"
)

func TestInitRedisClient(t *testing.T) {
	fc := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	// set url
	req.Header.SetMethod("GET")
	url := ""
	req.SetRequestURI(url)
	resp := fasthttp.AcquireResponse()

	if err := fc.Do(req, resp); err != nil {
		fmt.Println("请求错误", err)
	}
	strs := strings.Split(string(resp.Body()), "\n")
	for _, str := range strs {
		fmt.Println("str:            ", str)
		time.Sleep(5 * time.Second)
	}

}
