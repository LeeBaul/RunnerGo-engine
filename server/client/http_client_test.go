package client

import (
	"fmt"
	"testing"
	"time"
)

func TestHTTPRequest(t *testing.T) {

	method := "GET"
	url := "http://localhost:9090/api/v1/query_range?query=up&amp;start=2015-07-01T20:10:30.781Z&amp;end=2015-07-01T20:11:00.781Z&amp;step=15s"
	body := []byte("123")
	headers := make(map[string]string)
	var timeout time.Duration
	timeout = 100

	resp, respTime, err := HTTPRequest(method, url, body, headers, timeout)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)
	fmt.Println(respTime)
}
