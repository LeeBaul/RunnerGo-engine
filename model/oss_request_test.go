package model

import (
	"fmt"
	"testing"
)

func TestDownLoad(t *testing.T) {

	client := NewOssClient("", "", "")
	bucket, _ := client.Bucket("")
	//bucket.GetConfig()
	list, _ := bucket.ListObjects()
	fmt.Println("1231231", list.Objects)
	DownLoad(client, "", "D:\\123\\9f07f4e6-2539-475b-bc3f-3c3ea188eeea.txt", "")
}
