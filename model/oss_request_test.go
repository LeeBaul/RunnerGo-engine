package model

import (
	"fmt"
	"testing"
)

func TestDownLoad(t *testing.T) {

	client := NewOssClient("http://oss-cn-beijing.aliyuncs.com", "LTAI5tEAzFMCX559VD8mRDoZ", "5IV7ZpVx95vBHZ3Y74jr9amaMtXpCQ")
	bucket, _ := client.Bucket("apipost")
	//bucket.GetConfig()
	list, _ := bucket.ListObjects()
	fmt.Println("1231231", list.Objects)
	DownLoad(client, "kunpeng/test/9f07f4e6-2539-475b-bc3f-3c3ea188eeea.txt", "D:\\123\\9f07f4e6-2539-475b-bc3f-3c3ea188eeea.txt", "apipost")
}
