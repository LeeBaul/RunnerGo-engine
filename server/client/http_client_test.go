package client

import (
	"fmt"
	"net/url"
	"testing"
)

func TestHTTPRequest(t *testing.T) {

	a := url.URL{Scheme: "ws", Host: "localhost:8080"}
	fmt.Println(a.String())
}
