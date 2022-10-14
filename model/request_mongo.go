package model

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kp-runner/log"
)

func NewMongoClient(user, password, host, db string) (mongoClient *mongo.Client, err error) {
	//clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s/%s", user, password, host, db))
	conf := fmt.Sprintf("mongodb://%s:%s@%s/%s", user, password, host, db)

	clientOptions := options.Client().ApplyURI(conf)
	mongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return
	}

	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		return
	}

	log.Logger.Info("mongo数据库建连成功")
	return
}

func NewCollection(db, table string, mongoClient *mongo.Client) (collection *mongo.Collection) {
	collection = mongoClient.Database(db).Collection(table)
	return
}

func Insert(collection *mongo.Collection, msg interface{}) {
	_, err := collection.InsertOne(context.TODO(), msg)
	if err != nil {
		log.Logger.Error("向mongo写入数据错误:", err)
	}
}

func QueryDebugStatus(collection *mongo.Collection, reportId string) string {

	filter := bson.D{{"report_id", reportId}}
	singleResult := collection.FindOne(context.TODO(), filter)
	cur, err := singleResult.DecodeBytes()
	if err != nil {
		return StopDebug
	}
	list, err := cur.Elements()
	if err != nil {
		return StopDebug
	}
	for _, value := range list {
		if value.Key() == "debug" {
			return string(value.Value().Value)
		}
	}
	return StopDebug
}
