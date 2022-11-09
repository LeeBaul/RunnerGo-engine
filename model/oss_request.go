package model

import (
	log2 "RunnerGo-engine/log"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func NewOssClient(endpoint, accessKeyID, accessKeySecret string) (client *oss.Client) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		log2.Logger.Error("创建oss客户端失败:", client)
	}
	return
}

func DownLoad(client *oss.Client, formPath, toPath, bucketName string) (err error) {
	if client == nil {
		return
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		log2.Logger.Error("获取储存空间失败", err)
		return
	}
	err = bucket.GetObjectToFile(formPath, toPath)
	if err != nil {
		log2.Logger.Error(err)
	}
	return

}
