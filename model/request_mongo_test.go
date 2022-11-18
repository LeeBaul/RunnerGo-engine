package model

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestQueryDebugStatus(t *testing.T) {

	mongoClient, err := NewMongoClient(
		"",
		"",
		"",
		"")
	if err != nil {
		fmt.Println("连接mongo错误：", err)
		return
	}
	defer mongoClient.Disconnect(context.TODO())

	debugCollection := NewCollection("", "report_data", mongoClient)
	filter := bson.D{{"reportid", "4"}}
	m := make(map[string]interface{})
	debugCollection.FindOne(context.TODO(), filter).Decode(m)
	value, ok := m["data"]
	if ok {
		fmt.Println("123", value)
	}

	//fmt.Println(QueryDebugStatus(debugCollection, 1298))

}
