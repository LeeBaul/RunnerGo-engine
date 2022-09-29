package model

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	log2 "kp-runner/log"
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
	log2.Logger.Debug("formPath............", formPath)
	log2.Logger.Debug("topath.............", toPath)
	err = bucket.GetObjectToFile(formPath, toPath)
	if err != nil {
		fmt.Println("111111111111", err)
		log2.Logger.Error(err)
	}
	return

}
