package model

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kp-runner/log"
)

func NewMongoClient(user, password, host string) (mongoClient *mongo.Client, err error) {
	clientOptions := options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s@%s", user, password, host))
	mongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return
	}

	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		return
	}

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
