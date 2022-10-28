package model

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestQueryReport(t *testing.T) {
	//client := NewEsClient("http://172.17.101.191:9200", "elastic", "ZSrfx4R6ICa3skGBpCdf")
	//if client == nil {
	//	fmt.Println("client", nil)
	//}
	//
	//result := QueryReport(client, "report", "667")
	//fmt.Println

	var a interface{}
	a = 123
	by, _ := json.Marshal(a)
	fmt.Println(string(by), a)
}
