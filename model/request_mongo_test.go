package model

import (
	"context"
	"fmt"
	"testing"
)

func TestQueryDebugStatus(t *testing.T) {

	mongoClient, err := NewMongoClient(
		"kunpeng",
		"kYjJpU8BYvb4EJ9x",
		"172.17.18.255:27017",
		"kunpeng")
	if err != nil {
		fmt.Println("连接mongo错误：", err)
		return
	}
	defer mongoClient.Disconnect(context.TODO())

	debugCollection := NewCollection("kunpeng", "debug_status", mongoClient)
	fmt.Println(QueryDebugStatus(debugCollection, 1298))
}
